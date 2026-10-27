package data

import (
	"strings"
	"testing"
)

// The Storting index is a copy of the Storting's own lists, so the tests are about
// the copy being complete and internally consistent. Seat counts are the number
// every "N of 9 parties" claim on the site rests on, so they get the hardest check.
func TestStortingIndexIsComplete(t *testing.T) {
	reps := Representatives()
	// Pinned to the session's size. A floor would pass while representatives were
	// quietly dropped, and the count is a fact of the artefact, not a preference.
	if len(reps) != 169 {
		t.Fatalf("index holds %d representatives, the Storting has 169", len(reps))
	}
	known := map[string]bool{}
	for _, p := range Store.Parties {
		known[p.ID] = true
	}
	for _, r := range reps {
		if r.ID == "" || r.LastName == "" {
			t.Errorf("incomplete representative: %+v", r)
		}
		if !known[r.PartyID] {
			t.Errorf("%s %s: unknown party id %q", r.FirstName, r.LastName, r.PartyID)
		}
		if r.County == "" {
			t.Errorf("%s %s has no county", r.FirstName, r.LastName)
		}
	}
	if len(Committees()) != 16 {
		t.Errorf("%d committees, the session has 16", len(Committees()))
	}
	seen := map[string]bool{}
	for _, c := range Committees() {
		if c.ID == "" || c.Name == "" {
			t.Errorf("incomplete committee: %+v", c)
		}
		if seen[c.ID] {
			t.Errorf("duplicate committee id %q", c.ID)
		}
		seen[c.ID] = true
	}
}

// The reconciliation that makes the seat counts trustworthy: they are derived from
// the representative list, and they must equal the per-party row totals on a
// recorded votering. Two separately produced artefacts of the same authority.
func TestSeatCountsMatchTheVoteRecord(t *testing.T) {
	counts := SeatCounts()
	voteringID, perParty, method := VoteCrossCheck()
	if voteringID == 0 || len(perParty) == 0 {
		t.Fatal("no cross-check stored: the seat counts rest on one artefact only")
	}
	if !strings.Contains(method, "must equal") {
		t.Errorf("cross-check method does not state what it asserts: %q", method)
	}
	total := 0
	for id, seats := range counts {
		code, ok := PartyCode(id)
		if !ok {
			t.Errorf("party %s has no source code, so it cannot be cross-checked", id)
			continue
		}
		got, present := perParty[code]
		if !present {
			t.Errorf("party %s (%s) is absent from votering %d", id, code, voteringID)
			continue
		}
		if got != seats {
			t.Errorf("party %s: %d seats in the representative list, %d rows on votering %d",
				id, seats, got, voteringID)
		}
		total += got
	}
	if total != 169 {
		t.Errorf("vote record accounts for %d representatives, the Storting has 169", total)
	}
	t.Logf("%d parties cross-checked against votering %d, %d seats total", len(counts), voteringID, total)
}

func TestSeatCountsSumToTheStorting(t *testing.T) {
	counts := SeatCounts()
	sum := 0
	for _, n := range counts {
		if n == 0 {
			t.Error("a party with zero seats is in the index: it would divide by zero in a share")
		}
		sum += n
	}
	if sum != SeatTotal() {
		t.Errorf("seats sum to %d, index holds %d representatives", sum, SeatTotal())
	}
	if len(counts) != 9 {
		t.Errorf("%d parties have seats, the dataset scores 9", len(counts))
	}
}

// The ranking must be deterministic, or a page listing parties by strength would
// reorder itself between requests.
func TestSeatRankingIsDeterministic(t *testing.T) {
	ranking := SeatRanking()
	counts := SeatCounts()
	if len(ranking) != len(counts) {
		t.Fatalf("ranking lists %d parties, there are %d", len(ranking), len(counts))
	}
	for i := 1; i < len(ranking); i++ {
		prev, cur := counts[ranking[i-1]], counts[ranking[i]]
		if cur > prev {
			t.Errorf("ranking is not descending at %d: %s %d after %s %d",
				i, ranking[i], cur, ranking[i-1], prev)
		}
		if cur == prev && ranking[i-1] > ranking[i] {
			t.Errorf("tie between %s and %s is not broken by id", ranking[i-1], ranking[i])
		}
	}
}

// The refusal is data, not a comment: a caller must be able to discover that
// committee membership was not collected, and why, without reading the builder.
func TestCommitteeMembershipIsRefusedWithEvidence(t *testing.T) {
	refusal, ok := CommitteeMembershipRefusal()
	if !ok {
		t.Fatal("no committee-membership refusal recorded: the gap would look like an omission")
	}
	if len(refusal.Reason) < 40 {
		t.Errorf("refusal reason is too thin to be useful: %q", refusal.Reason)
	}
	if len(refusal.Attempts) < 5 {
		t.Errorf("only %d attempts recorded; the next run would repeat the search", len(refusal.Attempts))
	}
	for _, a := range refusal.Attempts {
		if !strings.Contains(a, "404") && !strings.Contains(a, "400") {
			t.Errorf("attempt does not record its outcome: %q", a)
		}
	}
}

func TestStortingCarriesProvenance(t *testing.T) {
	name, api, human, retrieved := StortingSourceInfo()
	if name == "" || api == "" || human == "" || retrieved == "" {
		t.Errorf("incomplete provenance: %q %q %q %q", name, api, human, retrieved)
	}
	if StortingSession() != "2025-2026" {
		t.Errorf("session = %q", StortingSession())
	}
}
