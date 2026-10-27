package analysis

import (
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

func vote(id int, sides map[string]string) data.Vote {
	return data.Vote{VoteringID: id, Forslag: "forslag", Parties: sides}
}

// The comparison rules are the whole value of this package, so each one gets its
// own case rather than one happy-path test.
func TestVoteAgreementCountsOnlyVotesBothTookASideOn(t *testing.T) {
	votes := []data.Vote{
		vote(1, map[string]string{"ap": "for", "h": "for"}),    // agree
		vote(2, map[string]string{"ap": "mot", "h": "for"}),    // disagree
		vote(3, map[string]string{"ap": "for"}),                // h absent, not compared
		vote(4, map[string]string{"ap": "for", "h": "absent"}), // explicitly absent
		vote(5, map[string]string{"ap": "split", "h": "for"}),  // split is a side, and disagrees
	}
	got := VoteAgreement(votes, []string{"ap", "h"}, 1)
	if len(got) != 1 {
		t.Fatalf("want one pair, got %d", len(got))
	}
	p := got[0]
	if p.Compared != 3 {
		t.Errorf("Compared = %d, want 3 (votes 1, 2 and 5)", p.Compared)
	}
	if p.Agree != 1 {
		t.Errorf("Agree = %d, want 1 (only vote 1)", p.Agree)
	}
	if p.Disagree() != 2 {
		t.Errorf("Disagree = %d, want 2", p.Disagree())
	}
}

func TestSplitAgreesOnlyWithSplit(t *testing.T) {
	votes := []data.Vote{
		vote(1, map[string]string{"ap": "split", "h": "split"}),
		vote(2, map[string]string{"ap": "split", "h": "for"}),
	}
	p := VoteAgreement(votes, []string{"ap", "h"}, 1)[0]
	if p.Compared != 2 || p.Agree != 1 {
		t.Errorf("compared %d agreed %d, want 2 and 1", p.Compared, p.Agree)
	}
}

func TestMinComparedFiltersPairsThatWereBarelyCompared(t *testing.T) {
	votes := []data.Vote{vote(1, map[string]string{"ap": "for", "h": "for"})}
	if got := VoteAgreement(votes, []string{"ap", "h"}, 2); len(got) != 0 {
		t.Errorf("a pair compared once must not appear with minCompared 2, got %+v", got)
	}
	if got := VoteAgreement(votes, []string{"ap", "h"}, 1); len(got) != 1 {
		t.Errorf("a pair compared once must appear with minCompared 1, got %+v", got)
	}
}

func TestEveryPairOfNinePartiesIsProduced(t *testing.T) {
	parties := []string{"ap", "frp", "h", "sv", "sp", "rodt", "mdg", "krf", "v"}
	sides := map[string]string{}
	for _, p := range parties {
		sides[p] = "for"
	}
	got := VoteAgreement([]data.Vote{vote(1, sides)}, parties, 1)
	if want := len(parties) * (len(parties) - 1) / 2; len(got) != want {
		t.Fatalf("got %d pairs, want %d for %d parties", len(got), want, len(parties))
	}
	for _, p := range got {
		if p.A != "ap" && p.B != "ap" && p.Compared != 1 {
			t.Errorf("pair %s/%s was not compared: %+v", p.A, p.B, p)
		}
	}
}

func TestPctOfNothingIsZeroNotNaN(t *testing.T) {
	if got := (Pair{}).Pct(); got != 0 {
		t.Errorf("Pct of an uncompared pair = %v, want 0", got)
	}
}

func TestMostAlignedPicksBothEnds(t *testing.T) {
	most, least, ok := MostAligned([]Pair{
		{A: "ap", B: "h", Compared: 10, Agree: 9},
		{A: "ap", B: "v", Compared: 10, Agree: 4},
		{A: "h", B: "v", Compared: 10, Agree: 7},
	})
	if !ok {
		t.Fatal("ok = false with three pairs")
	}
	if most.A != "ap" || most.B != "h" {
		t.Errorf("most = %s/%s, want ap/h", most.A, most.B)
	}
	if least.A != "ap" || least.B != "v" {
		t.Errorf("least = %s/%s, want ap/v", least.A, least.B)
	}
	if _, _, ok := MostAligned(nil); ok {
		t.Error("ok = true with no pairs")
	}
}

// And once against the real index, so the derivation is exercised on the data it
// will actually run on rather than on fixtures alone.
func TestVoteAgreementOnTheRealIndex(t *testing.T) {
	votes := data.Votes()
	if len(votes) == 0 {
		t.Fatal("no votes in the index")
	}
	parties := []string{"ap", "frp", "h", "sv", "sp", "rodt", "mdg", "krf", "v"}
	counts := VotesByParty(votes, parties)
	for _, p := range parties {
		if counts[p] == 0 {
			t.Errorf("party %q took a side on no indexed vote", p)
		}
	}
	pairs := VoteAgreement(votes, parties, 5)
	if len(pairs) == 0 {
		t.Fatal("no pair was compared at least 5 times: the index is too thin to say anything")
	}
	most, least, ok := MostAligned(pairs)
	if !ok {
		t.Fatal("no most-aligned pair")
	}
	if most.Pct() < least.Pct() {
		t.Errorf("most %v < least %v", most.Pct(), least.Pct())
	}
	t.Logf("%d votes, %d pairs over 5 comparisons; most aligned %s/%s %.0f%% (%d of %d), least %s/%s %.0f%%",
		len(votes), len(pairs), most.A, most.B, most.Pct(), most.Agree, most.Compared,
		least.A, least.B, least.Pct())
}
