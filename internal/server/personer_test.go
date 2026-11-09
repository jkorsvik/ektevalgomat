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
