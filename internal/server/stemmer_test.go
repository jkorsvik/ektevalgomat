package server

import (
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
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
