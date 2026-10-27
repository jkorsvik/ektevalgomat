package server

import (
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/analysis"
	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The vote-alignment page is a derived view, so it gets the same treatment as the
// coverage claims: the numbers it prints are recomputed here from the dataset, and
// the page is loaded over HTTP rather than inspected through the view model. A view
// that renders in a struct but 500s on the wire is not a page.
func TestStemmerPageRendersAndDerivesItsNumbers(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/stemmer", nil))
	if w.Code != 200 {
		t.Fatalf("/stemmer answered %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Hvem stemmer med hvem") {
		t.Fatal("/stemmer rendered without its heading")
	}

	// Recompute independently: the same derivation the page uses, from the same
	// data, and require the exact string the page must carry.
	ids := make([]string, 0, len(data.Store.Parties))
	short := map[string]string{}
	for _, p := range data.Store.Parties {
		ids = append(ids, p.ID)
		short[p.ID] = p.Short
	}
	pairs := analysis.VoteAgreement(data.Votes(), ids, minComparedForRanking)
	if len(pairs) == 0 {
		t.Skip("no pair is compared often enough to rank; nothing to assert")
	}
	most, least, ok := analysis.MostAligned(pairs)
	if !ok {
		t.Fatal("MostAligned returned ok=false with a non-empty list")
	}
	for _, p := range []analysis.Pair{most, least} {
		want := fmt.Sprintf("%d %%", int(p.Pct()+0.5))
		if !strings.Contains(body, want) {
			t.Errorf("page does not carry %q for the pair %s/%s (%.1f%%)", want, p.A, p.B, p.Pct())
		}
		if !strings.Contains(body, short[p.A]) || !strings.Contains(body, short[p.B]) {
			t.Errorf("page does not name %s/%s, whose figure it prints", short[p.A], short[p.B])
		}
	}

	// The denominator is the whole point of the page: a percentage with no
	// comparison count reads as stronger evidence than it is.
	if !strings.Contains(body, fmt.Sprintf("av %d", most.Compared)) {
		t.Errorf("page does not state the denominator %d next to the most-aligned pair", most.Compared)
	}
	t.Logf("rendered %d bytes; most aligned %s/%s %.0f%% of %d",
		len(body), most.A, most.B, most.Pct(), most.Compared)
}

// A party that took a side on no indexed vote would render as a row of dashes,
// which reads as "this party did not vote" rather than "the index does not cover
// it". The page's per-party counts exist to make that distinction visible.
func TestStemmerPageCountsVotesPerParty(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/stemmer", nil))
	body := w.Body.String()
	if !strings.Contains(body, "Voteringer med stemme") {
		t.Fatal("the per-party denominator table is missing")
	}
	counts := analysis.VotesByParty(data.Votes(), partyIDs(data.Store))
	zero := 0
	for pid, n := range counts {
		if n == 0 {
			zero++
			t.Logf("party %s took a side on no indexed vote", pid)
		}
	}
	if zero == len(counts) {
		t.Fatal("no party has a vote at all: the index is empty, which the page cannot say honestly")
	}
}

func partyIDs(ds *data.Dataset) []string {
	ids := make([]string, 0, len(ds.Parties))
	for _, p := range ds.Parties {
		ids = append(ids, p.ID)
	}
	return ids
}

// The alignment page must say which sessions it covers. A matrix built from three
// of the eighteen sessions the API serves reads as if it covers the record, and the
// reader has no other way to find out. The sessions are derived from the index, so
// widening the corpus without updating the page fails here.
func TestStemmerPageStatesItsSessionCoverage(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/stemmer", nil))
	if w.Code != 200 {
		t.Fatalf("/stemmer answered %d, want 200", w.Code)
	}
	body := w.Body.String()

	// Derived from the raw votes, not from data.VoteSessions(): a test that asks the
	// same helper the page asks cannot notice the helper losing a session. This was
	// found by mutation — truncating VoteSessions() to one session passed, because
	// page and test both counted one and agreed.
	want := map[string]bool{}
	withSession := 0
	for _, v := range data.Votes() {
		if v.Session != "" {
			want[v.Session] = true
			withSession++
		}
	}
	if len(want) == 0 {
		t.Fatal("no vote records a session: the coverage claim cannot be checked")
	}
	for ses := range want {
		if !strings.Contains(body, ses) {
			t.Errorf("the page does not name session %s, which the index was built from", ses)
		}
	}
	if !strings.Contains(body, strconv.Itoa(len(want))+" sesjoner") {
		t.Errorf("the page does not state that the index covers %d sessions", len(want))
	}
	if withSession == 0 || withSession > data.VoteCount() {
		t.Fatalf("VotesWithSession() = %d, want between 1 and %d", withSession, data.VoteCount())
	}
	if !strings.Contains(body, strconv.Itoa(withSession)+" av "+strconv.Itoa(data.VoteCount())) {
		t.Errorf("the page does not state %d of %d votes have a recorded session", withSession, data.VoteCount())
	}
	t.Logf("%d sessions, %d of %d votes with a session", len(want), withSession, data.VoteCount())
}
