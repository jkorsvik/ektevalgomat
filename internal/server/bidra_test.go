package server

import (
	"html"
	"log/slog"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
	"github.com/jkorsvik/ektevalgomat/internal/scoring"
)

// /bidra exists to point a volunteer at real work. Every number and every row
// on it is derived, so the failure mode is not a typo in prose: it is the page
// listing a cell that is already coded, or missing one that is not. Both send
// someone to do work that does not need doing, which is how a contribution
// page loses the people it was built for. So this recomputes the gap set from
// the dataset independently and compares it with what the page rendered.
func TestBidraListsExactlyTheUncodedCells(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/bidra", nil))
	if w.Code != 200 {
		t.Fatalf("/bidra: status %d", w.Code)
	}
	body := w.Body.String()

	// Independent truth: every statement/party cell with no coded position.
	type cell struct{ statement, party string }
	want := map[cell]bool{}
	noSource, uncoded := 0, 0
	for _, s := range data.Store.Statements {
		for _, p := range data.Store.Parties {
			c := scoring.ComputeCell(data.Store, s.ID, p.ID)
			if c.PartyPosition != nil {
				continue
			}
			want[cell{s.ID, p.ID}] = true
			if c.HasEvidence() {
				uncoded++
			} else {
				noSource++
			}
		}
	}
	if len(want) == 0 {
		t.Fatal("no gap cells found: this test would pass on any page")
	}

	// Every rendered row must name a real gap. Rows carry the statement id in
	// the link text and the party id in the party link, so parse those.
	// (?s): the two links sit on different lines, and Go's `.` does not cross
	// a newline by default. Without it this matched nothing and the test
	// reported every gap as omitted, which is a failure of the test, not the page.
	rows := regexp.MustCompile(`(?s)<tr>\s*<td><a href="/utsagn/\d+">(utsagn-[0-9]+)</a>.*?<a href="/parti/([a-z]+)">`).FindAllStringSubmatch(body, -1)
	got := map[cell]bool{}
	for _, r := range rows {
		got[cell{r[1], r[2]}] = true
	}
	if len(got) != len(want) {
		t.Errorf("/bidra lists %d gap cells; the dataset has %d", len(got), len(want))
	}
	for c := range want {
		if !got[c] {
			t.Errorf("/bidra omits a gap: %s / %s", c.statement, c.party)
		}
	}
	for c := range got {
		if !want[c] {
			t.Errorf("/bidra lists %s / %s, which has a coded position: that is work already done", c.statement, c.party)
		}
	}

	// The two kinds of gap must be counted separately and must add up. One
	// "missing" total would hide that one of them needs finding a source and
	// the other needs reading one that is already linked.
	if !strings.Contains(body, "<strong>"+strconv.Itoa(noSource)+"</strong>") {
		t.Errorf("/bidra does not state the %d cells that have no source at all", noSource)
	}
	if !strings.Contains(body, "<strong>"+strconv.Itoa(uncoded)+"</strong>") {
		t.Errorf("/bidra does not state the %d cells that have evidence but no position", uncoded)
	}
	if !strings.Contains(body, strconv.Itoa(len(want))+" partiposisjoner") {
		t.Errorf("/bidra does not state the total of %d cells", len(want))
	}
}

// A row a volunteer cannot act on is decoration. Each one has to carry a
// prefilled issue that names the statement and the party.
func TestEveryBidraRowHasAUsableIssueLink(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/bidra", nil))
	body := w.Body.String()

	rows := strings.Count(body, "<tr>")
	links := regexp.MustCompile(`issues/new\?[^"]*template=rapporter-kilde\.md[^"]*`).FindAllString(body, -1)
	if len(links) != rows-1 { // one <tr> is the header
		t.Errorf("%d rows but %d prefilled issues: every row needs one", rows-1, len(links))
	}
	for _, l := range links {
		if !strings.Contains(l, "title=") {
			t.Errorf("issue link with no title: GitHub would open a blank required field: %s", l[:80])
		}
	}
	// html/template escapes the + in a query string to &#43;, so the raw
	// output cannot be searched for the encoded text as written.
	plain := html.UnescapeString(body)
	if !strings.Contains(plain, "Utsagn%3A+") {
		t.Error("issue bodies do not name the statement")
	}
	if !strings.Contains(plain, "Parti%3A+") {
		t.Error("issue bodies do not name the party")
	}
	if !strings.Contains(plain, "Utsagnet%3A") {
		t.Error("issue bodies do not quote the statement text")
	}
}

// The tier counts are the page's second job: which sources are worth
// replacing. They must come from the dataset, not from prose.
func TestBidraTierCountsComeFromTheDataset(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/bidra", nil))
	body := w.Body.String()

	truth := map[string]int{}
	for _, e := range data.Store.Evidence {
		truth[e.SourceTier]++
	}
	found := 0
	for _, m := range regexp.MustCompile(`Kildegrad ([A-D])</strong>: (\d+) kilder`).FindAllStringSubmatch(body, -1) {
		found++
		tier, n := m[1], m[2]
		if strconv.Itoa(truth[tier]) != n {
			t.Errorf("/bidra says %s sources at tier %s; the dataset has %d", n, tier, truth[tier])
		}
	}
	if found == 0 {
		t.Fatal("/bidra states no tier counts: this test checked nothing")
	}
	// Tier A is the primary-source tier and is not a gap, so listing it as
	// weak would tell volunteers to replace sources that need no replacing.
	if strings.Contains(body, "Kildegrad A</strong>:") {
		t.Error("/bidra lists tier A among the weak sources")
	}
}

// The challenge route must be reachable from the pages a reader actually
// visits, or the page exists for nobody. Every page, not two of them: the
// first version linked it from /data and /om only, so eight of ten pages hid
// the only route into it.
func TestBidraIsReachable(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	everyPage := []string{
		"/", "/partier", "/budsjett", "/metode", "/data", "/bidra", "/om",
		"/personvern", "/resultat", "/utsagn/1", "/parti/ap",
	}
	for _, from := range everyPage {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", from, nil))
		if w.Code != 200 {
			t.Fatalf("%s: status %d", from, w.Code)
		}
		if !strings.Contains(w.Body.String(), `href="/bidra"`) {
			t.Errorf("%s does not link to /bidra", from)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/sitemap.xml", nil))
	if !strings.Contains(w.Body.String(), "/bidra") {
		t.Error("sitemap.xml omits /bidra")
	}
}
