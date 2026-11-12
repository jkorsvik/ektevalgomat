package server

import (
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

func TestPersonerPageRenders(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/personer", nil))
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	body := w.Body.String()
	for _, want := range []string{
		"<h1>Personer</h1>",
		"Stortingsrepresentant",
		"Øyeblikksbilde",
		"ikke avsluttet i registeret",
		"Vervregisteret ajourført " + data.VervSnapshotDate(),
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/personer is missing %q", want)
		}
	}
}

// Every person on disk must reach the page: a record the page cannot show is
// not an overview.
func TestPersonerPageShowsEveryRecord(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/personer", nil))
	body := w.Body.String()

	// Names are rendered in the first cell of each person's row group.
	missing := 0
	for _, p := range data.People() {
		if !strings.Contains(body, p.Name) {
			missing++
			if missing <= 3 {
				t.Errorf("person %s is not on /personer", p.ID)
			}
		}
	}
	if missing > 3 {
		t.Errorf("%d records missing from /personer", missing)
	}

	// The role count in the lede must equal the roles actually rendered as rows.
	view := buildPersonerView(data.Store, SnapshotDate)
	if view.RoleTotal != len(data.People())+countExtraRoles() {
		t.Errorf("RoleTotal %d does not match the rendered role rows", view.RoleTotal)
	}
}

func countExtraRoles() int {
	extra := 0
	for _, p := range data.People() {
		extra += len(p.Roles) - 1
	}
	return extra
}

// The register's snapshot date is not a role start and the page must say so.
func TestSnapshotDatesAreLabelled(t *testing.T) {
	view := buildPersonerView(data.Store, SnapshotDate)
	snap, dated := 0, 0
	for _, p := range view.People {
		for _, r := range p.Roles {
			if r.Snapshot {
				snap++
			} else {
				dated++
			}
		}
	}
	if snap == 0 {
		t.Fatal("no role carries the register snapshot; the label path is untested")
	}
	if dated == 0 {
		t.Fatal("no role carries a date the register stated; date extraction is dead")
	}
}

// The register comparison is the only place the page can say what changed
// rather than what is. It must appear when two editions were compared, and it
// must not claim a date the register does not state.
func TestRegisterHistorySection(t *testing.T) {
	hist := data.RegisterHistoryDiff()
	if len(hist.ComparedBetween) != 2 {
		t.Skip("no two register editions compared in this build")
	}
	h := New(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/personer", nil))
	body := w.Body.String()

	for _, want := range []string{
		"Verv som er borte siden " + hist.ComparedBetween[0],
		hist.ComparedBetween[1],
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/personer is missing %q", want)
		}
	}
	if hist.Counts.RolesEnded == 0 {
		t.Error("history has no ended roles; the comparison produced nothing")
	}
	// The wording must not turn the comparison into a claim about a date the
	// register never published.
	if strings.Contains(body, "tok slutt ") {
		t.Error("the page claims an end date the register does not state")
	}
}

// The party table must be deterministic. Sorting on count alone left equal
// counts in map-iteration order, so the table reshuffled on every request.
func TestPartyTableOrderIsDeterministic(t *testing.T) {
	first := buildPersonerView(data.Store, SnapshotDate).Parties
	for i := 0; i < 5; i++ {
		again := buildPersonerView(data.Store, SnapshotDate).Parties
		for j := range first {
			if first[j] != again[j] {
				t.Fatalf("party order changed between builds at %d: %v vs %v", j, first[j], again[j])
			}
		}
	}
	for i := 1; i < len(first); i++ {
		if first[i-1].Count < first[i].Count {
			t.Errorf("party %s comes after a smaller count", first[i].Party)
		}
		if first[i-1].Count == first[i].Count && first[i-1].Party > first[i].Party {
			t.Errorf("ties not sorted by name: %s before %s", first[i-1].Party, first[i].Party)
		}
	}
}

// The vote sentence is the first thing on the site that says how one named person
// voted, so the wording has to match the numbers behind it. Absence is never
// mentioned: a missing row and an absent member are indistinguishable here.
func TestVoteSentenceStatesWhatTheNumbersSay(t *testing.T) {
	cases := []struct {
		voted, divisions, deviation int
		want                        string
	}{
		{45, 101, 0, "Stemte i 45 av 101 divisjoner der Stortinget førte navn. Aldri mot eget partis linje."},
		{45, 101, 1, "Stemte i 45 av 101 divisjoner der Stortinget førte navn. 1 gang mot eget partis linje."},
		{45, 101, 3, "Stemte i 45 av 101 divisjoner der Stortinget førte navn. 3 ganger mot eget partis linje."},
		{1234, 2000, 0, "Stemte i 1 234 av 2 000 divisjoner der Stortinget førte navn. Aldri mot eget partis linje."},
	}
	for _, c := range cases {
		if got := voteSentence(c.voted, c.divisions, c.deviation); got != c.want {
			t.Errorf("voteSentence(%d,%d,%d)\n got: %s\nwant: %s", c.voted, c.divisions, c.deviation, got, c.want)
		}
	}
}

// The page must carry the artefact's coverage, because a reader who sees "stemte
// i 45 av 101 divisjoner" needs to know that 101 is not every division held.
func TestPersonerPageStatesVoteCoverage(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/personer")
	if !strings.Contains(body, "Navngitte stemmer:") {
		t.Fatal("the page does not state how much of the vote record is per-person")
	}
	if !strings.Contains(body, "mot eget partis linje") && !strings.Contains(body, "Aldri mot eget partis linje") {
		t.Error("no representative carries a vote sentence at all")
	}
}

// Committee seats are a snapshot, so the wording must not read like a dated role,
// and a leader must be distinguishable from an ordinary member.
func TestCommitteeSentenceNamesSeatsAndRoles(t *testing.T) {
	// Order is the data layer's job (CommitteesForPerson puts the leader first);
	// this function only formats. The test feeds the order it would receive.
	seats := []data.CommitteeSeat{
		{CommitteeName: "Kontroll- og konstitusjonskomiteen", Role: "Leder"},
		{CommitteeName: "Finanskomiteen", Role: "Medlem"},
	}
	got := committeeSentence(seats)
	want := "Komité: Kontroll- og konstitusjonskomiteen (leder), Finanskomiteen."
	if got != want {
		t.Errorf("committeeSentence\n got: %s\nwant: %s", got, want)
	}
	if committeeSentence(nil) != "" {
		t.Error("a person with no seat must render nothing, not an empty label")
	}
}

// The page must name the bodies the source publishes no member list for. Silence
// about them would read as "nobody sits there".
func TestPersonerPageStatesCommitteeCoverageAndGaps(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/personer")
	if !strings.Contains(body, "Komitémedlemskap er hentet") {
		t.Fatal("the page does not say where committee membership comes from")
	}
	if !strings.Contains(body, "øyeblikksbilde") {
		t.Error("the page does not state that membership is a snapshot, not a dated role")
	}
	if !strings.Contains(body, "Uten publisert medlemsliste") {
		t.Error("the page does not name the bodies the source omits")
	}
}
