package server

import (
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"regexp"
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
	// The optimum is recomputed here with its own comparison rather than by calling
	// analysis.MostAligned: a test that asks the same helper the page asks cannot
	// notice the helper picking the wrong pair, which is exactly what happened when
	// the two definitions of "most aligned" disagreed on ties.
	most, least := pairs[0], pairs[0]
	better := func(a, b analysis.Pair) bool {
		if a.Pct() != b.Pct() {
			return a.Pct() > b.Pct()
		}
		if a.Compared != b.Compared {
			return a.Compared > b.Compared
		}
		return a.A < b.A
	}
	for _, p := range pairs[1:] {
		if better(p, most) {
			most = p
		}
		if better(least, p) {
			least = p
		}
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
	//
	// Asserted inside the headline sentence, not across the whole document. The page
	// is several megabytes of matrix, so any number appears somewhere in it, and a
	// document-wide Contains let a swapped Most/Least pass.
	headline := regexp.MustCompile(`(?s)Mest enige:.{0,600}?Minst enige:`).FindString(body)
	if headline == "" {
		t.Fatal("the page has no 'Mest enige ... Minst enige' sentence")
	}
	if !strings.Contains(headline, fmt.Sprintf("av %d", most.Compared)) {
		t.Errorf("the headline does not state the denominator %d for the most-aligned pair", most.Compared)
	}
	if !strings.Contains(headline, short[most.A]) || !strings.Contains(headline, short[most.B]) {
		t.Errorf("the headline does not name %s/%s, whose figure it prints", short[most.A], short[most.B])
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

// The vote list exists so a reader can check a figure against the votes behind it,
// but only a cited vote has an evidence row to check against. Listing every indexed
// vote made the page 11 MB once the corpus passed 14 000, which is not a page anyone
// can read. The list is bounded to cited votes and the page states the remainder,
// derived here from the raw votes rather than from the view model.
func TestStemmerListsOnlyCitedVotesAndSaysSo(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/stemmer", nil))
	if w.Code != 200 {
		t.Fatalf("/stemmer answered %d, want 200", w.Code)
	}
	body := w.Body.String()

	cited, uncited := 0, 0
	for _, v := range data.Votes() {
		if len(v.CitedBy) > 0 {
			cited++
		} else {
			uncited++
		}
	}
	if cited == 0 || uncited == 0 {
		t.Fatalf("cited=%d uncited=%d: the split cannot be checked", cited, uncited)
	}

	// Every cited vote is listed, by id.
	for _, v := range data.Votes() {
		if len(v.CitedBy) == 0 {
			continue
		}
		if !strings.Contains(body, fmt.Sprintf(`<td class="mono">%d</td>`, v.VoteringID)) {
			t.Errorf("cited votering %d is not listed", v.VoteringID)
		}
	}

	// The remainder is stated, not silently dropped.
	if !strings.Contains(body, strconv.Itoa(uncited)+" av") {
		t.Errorf("the page does not state that %d votes are unlisted", uncited)
	}

	// And the page stays readable: a regression to listing everything would blow this
	// budget long before anyone opened it.
	if len(body) > 2_000_000 {
		t.Errorf("the page is %d bytes; the vote list has grown back", len(body))
	}
	t.Logf("%d cited listed, %d unlisted, page %d bytes", cited, uncited, len(body))
}
