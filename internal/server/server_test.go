package server

import (
	"bytes"
	"fmt"
	"html"
	"log/slog"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// TestNoVisitLog: /personvern tells visitors that nothing they do on the
// site is stored by us ("ingen besøkslogg"). Serving normal pages must
// therefore write nothing to the log. Errors may still be logged.
func TestNoVisitLog(t *testing.T) {
	var buf bytes.Buffer
	h := New(slog.New(slog.NewJSONHandler(&buf, nil)))
	buf.Reset() // startup lines (dataset loaded) are not visits
	for _, path := range []string{"/", "/utsagn/52", "/resultat", "/personvern", "/static/css/style.css", "/healthz"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Fatalf("%s: status %d", path, w.Code)
		}
	}
	if buf.Len() != 0 {
		t.Errorf("serving pages wrote to the log, but /personvern promises no visit log:\n%s", buf.String())
	}
}

// TestStaticCacheHeaders: versioned assets and fonts are immutable for a
// year. An unversioned asset, a directory listing and an error response are
// not: each can serve different bytes under the same URL later.
func TestStaticCacheHeaders(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	const immutable = "public, max-age=31536000, immutable"
	for _, tc := range []struct {
		path, want string
		code       int
	}{
		{"/static/css/style.css?v=abc", immutable, 200},
		{"/static/fonts/IBMPlexSerif-SemiBold-Latin1.woff2", immutable, 200},
		{"/static/css/style.css", "", 200},
		{"/static/fonts/", "", 200},
		{"/static/fonts/missing.woff2", "", 404},
		{"/static/missing.css?v=abc", "", 404},
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		if w.Code != tc.code {
			t.Errorf("%s: status %d, want %d", tc.path, w.Code, tc.code)
		}
		if got := w.Header().Get("Cache-Control"); got != tc.want {
			t.Errorf("%s: Cache-Control %q, want %q", tc.path, got, tc.want)
		}
	}
}

// TestEveryPageUsesSharedPartials: a page that drops {{template "head"}}
// loses its stylesheet and still returns 200, and a page that drops
// {{template "header"}} loses the mark and the nav. Neither fails at startup, so check
// the rendered HTML of every route.
func TestEveryPageUsesSharedPartials(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	for _, path := range []string{"/", "/om", "/metode", "/personvern", "/resultat", "/utsagn/52"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		body := w.Body.String()
		for _, want := range []string{
			`<link rel="stylesheet" href="/static/css/style.css?v=`,
			`<link rel="preload" href="/static/fonts/IBMPlexSerif-SemiBold-Latin1.woff2"`,
			`<h1 class="site-title"><a class="brand" href="/"><svg class="brand-mark"`,
			`<a href="/metode">Metode</a>`,
		} {
			if w.Code != 200 || !strings.Contains(body, want) {
				t.Errorf("%s (status %d): missing %s", path, w.Code, want)
			}
		}
	}
	// Kildegrad: the letter is real text, never colour or meter alone.
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/utsagn/52", nil))
	if !regexp.MustCompile(`<span class="grade g-([A-E])" title="Kildegrad [A-E] — [^"]+"><b>[A-E]</b>`).MatchString(w.Body.String()) {
		t.Error("/utsagn/52: no kildegrad element with its letter as text")
	}
}

// TestIndexNoscriptFallbackHasRealLinks locks in the fix for a shipped bug:
// the <noscript> fallback claimed "you can still read the whole evidence
// chain directly" but its <ul id="noscript-links"> was an empty shell,
// because it was meant to be filled by the same JavaScript that a
// noscript-context browser, by definition, isn't running. This test fails
// loudly if that regresses to an empty list.
func TestIndexNoscriptFallbackHasRealLinks(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body := w.Body.String()

	start := strings.Index(body, `id="noscript-links"`)
	if start == -1 {
		t.Fatal("index page missing #noscript-links element")
	}
	end := strings.Index(body[start:], "</ul>")
	if end == -1 {
		t.Fatal("no closing </ul> found for #noscript-links")
	}
	block := body[start : start+end]

	links := regexp.MustCompile(`<a href="/utsagn/\d+">`).FindAllString(block, -1)
	if len(links) == 0 {
		t.Fatal("noscript fallback promises the evidence chain is readable " +
			"without JS, but #noscript-links contains zero <a href=\"/utsagn/N\"> " +
			"links — the fallback is a dead end for real noscript users")
	}
}

// TestStrengthLineSaysWhichSourcesCarryTheLabel: on utsagn-06, Ap, Sp and
// MDG list two items (a coded merknad and an uncoded vote) but the label
// is computed from the merknad alone. The line must say so, and only on
// those three cells; "Moderat (2 kilder)" alone would imply both
// sources back the label (2026-10-02 review). The count writes "kilde"
// or "kilder" rather than "kilde(r)" as of the 2026-10-13 klarspråk
// pass, so this expectation tracks the template's own singular/plural.
func TestStrengthLineSaysWhichSourcesCarryTheLabel(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/utsagn/6", nil))
	body := w.Body.String()
	want := "Moderat (2 kilder, styrken bygger på 1 med posisjon)"
	if got := strings.Count(body, want); got != 3 {
		t.Errorf("/utsagn/6: %q appears %d times, want 3 (Ap, Sp, MDG)", want, got)
	}
	if got := strings.Count(body, "styrken bygger på"); got != 3 {
		t.Errorf("/utsagn/6: basis clause appears %d times, want 3 (only where basis < count)", got)
	}
}

// TestUncodedPartyCardShowsNoStrengthRating: a party card whose verdict
// is "Ikke vurdert" has no position, so its Dokumentasjon line must not
// carry dots or a Svak/Moderat/Sterk rating. Ap on utsagn-57 used to
// read "Ikke vurdert" next to "●●● Sterk" (2026-10-02). Checked on every
// statement page, and the card count guards against a vacuous pass.
// The count moved 3 -> 2 -> 1 on 2026-10-13, as Rødt's and then Sp's
// programmer were coded on utsagn-45. Both cards there now carry a
// position, so utsagn-57 Ap is the only uncoded card left in the dataset.
func TestUncodedPartyCardShowsNoStrengthRating(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	uncodedCards := 0
	for _, st := range data.Store.Statements {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", fmt.Sprintf("/utsagn/%d", st.Number), nil))
		if w.Code != 200 {
			t.Fatalf("/utsagn/%d: status %d", st.Number, w.Code)
		}
		cards := strings.Split(w.Body.String(), `class="card party-card"`)[1:]
		for _, card := range cards {
			if !strings.Contains(card, `<span class="party-word">Ikke vurdert</span>`) {
				continue
			}
			uncodedCards++
			// Head and Dokumentasjon line only; the evidence items below
			// may quote any word.
			end := strings.Index(card, "Historikk")
			if end < 0 {
				t.Fatalf("/utsagn/%d: party card has no Historikk line, so this test can no longer find the card head", st.Number)
			}
			head := card[:end]
			for _, banned := range []string{"●", "Sterk", "Moderat", "Svak", "party-value"} {
				if strings.Contains(head, banned) {
					t.Errorf("/utsagn/%d: uncoded party card contains %q", st.Number, banned)
				}
			}
			if !regexp.MustCompile(`Ingen kodet posisjon \(\d+ kilder?, vist som kontekst\)`).MatchString(head) {
				t.Errorf("/utsagn/%d: uncoded party card does not say it has no coded position:\n%s", st.Number, head)
			}
		}
	}
	if uncodedCards != 1 {
		t.Errorf("found %d uncoded party cards, want 1 (utsagn-57 Ap); update this count if the data changed on purpose", uncodedCards)
	}
}

// TestDimensionCoverageNeverOverclaims is the regression guard for the
// stale hand-maintained count that let the front page show a dimension at
// more than 100% built (2026-10 overclaim). Every row's Total must equal
// the number of statements that actually exist in the dataset, and Built
// must never exceed Total. Both summaries are checked so a reintroduced
// hand-maintained map fails here.
func TestDimensionCoverageNeverOverclaims(t *testing.T) {
	ds := data.Store
	counts := map[string]int{}
	for _, s := range ds.Statements {
		if s.Dimension == addendumDimension {
			continue
		}
		counts[s.Dimension]++
	}
	checked := 0
	for _, rows := range [][]dimensionCoverage{dimensionSummary(ds), allDimensionsSummary(ds)} {
		for _, r := range rows {
			checked++
			if r.Built > r.Total {
				t.Errorf("dimension %q: built %d > total %d (would show more than 100%% built)", r.Name, r.Built, r.Total)
			}
			if r.Total != counts[r.Name] {
				t.Errorf("dimension %q: total %d, but %d statements exist in the dataset", r.Name, r.Total, counts[r.Name])
			}
			if r.Built != counts[r.Name] {
				t.Errorf("dimension %q: built %d, but %d statements exist in the dataset", r.Name, r.Built, counts[r.Name])
			}
		}
	}
	if checked == 0 {
		t.Fatal("no dimension rows checked, the guard ran on nothing")
	}
	t.Logf("checked %d dimension rows", checked)
}

// TestFrontPageDimensionTotalsAreDerived renders the front page and checks
// the per-dimension "0 av N besvart" denominators against the number of
// statements that exist. Four dimensions were stale in the old map (Norsk
// handlefrihet, Universell grunnmur, Robusthet / adaptabilitet and
// Finansiering av fellesskapet); each must now show its real count.
func TestFrontPageDimensionTotalsAreDerived(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	body := w.Body.String()
	want := map[string]int{
		"norsk-handlefrihet":           7, // old map said 6
		"universell-grunnmur":          7, // old map said 6
		"robusthet-adaptabilitet":      5, // old map said 4
		"finansiering-av-fellesskapet": 7, // old map said 5
	}
	for slug, n := range want {
		re := regexp.MustCompile(`id="dim-row-` + regexp.QuoteMeta(slug) + `"[\s\S]*?0 av (\d+) besvart`)
		m := re.FindStringSubmatch(body)
		if m == nil {
			t.Errorf("front page has no rendered row for dimension slug %q", slug)
			continue
		}
		if m[1] != fmt.Sprintf("%d", n) {
			t.Errorf("dimension %q shows denominator %s, want %d", slug, m[1], n)
		}
	}
}

// TestFooterLinksThePublicSource: the footer must name the repository a
// visitor can actually open, and must never name one they cannot.
//
// This replaced TestFooterDoesNotClaimOpenSource, which asserted the
// opposite ("Kildekoden er ikke offentlig tilgjengelig ennå") because the
// site lived inside a private monorepo. The code is now published on its
// own, so the honest line is a link to it. The rule that survives both
// versions is the one worth keeping: the footer may only claim what a
// visitor can verify in one click.
func TestFooterLinksThePublicSource(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	paths := []string{"/", "/om", "/metode", "/personvern", "/resultat", "/utsagn/52"}
	checked := 0
	for _, path := range paths {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Fatalf("%s: status %d", path, w.Code)
		}
		body := w.Body.String()
		checked++
		if !strings.Contains(body, "https://github.com/jkorsvik/ektevalgomat") {
			t.Errorf("%s: footer does not link the public source repository", path)
		}
		// The footer used to say the source was unavailable, because it
		// lived in a private repository. Neither the claim nor a link to
		// that private repository may come back.
		for _, banned := range []string{"ikke offentlig tilgjengelig"} {
			if strings.Contains(body, banned) {
				t.Errorf("%s: footer still contains %q", path, banned)
			}
		}
		// The link without the licence would be a claim with no terms.
		if !strings.Contains(body, "AGPL") {
			t.Errorf("%s: footer links the source but names no licence", path)
		}
	}
	if checked == 0 {
		t.Fatal("no pages checked")
	}
}

// TestProgrammesIndexLinksEveryPartyPage: /parti/{id} has been live since
// 2026-10-03, but /partier never linked to it, so nine working pages were
// reachable only by typing the URL. This pins the links.
func TestProgrammesIndexLinksEveryPartyPage(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/partier", nil))
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	body := w.Body.String()
	linked := map[string]bool{}
	for _, m := range regexp.MustCompile(`href="/parti/([a-z]+)"`).FindAllStringSubmatch(body, -1) {
		linked[m[1]] = true
	}
	parties := data.Store.Parties
	if len(parties) < 9 {
		t.Fatalf("only %d parties in the dataset, expected at least 9", len(parties))
	}
	for _, p := range parties {
		if !linked[p.ID] {
			t.Errorf("/partier does not link to /parti/%s", p.ID)
		}
	}
}

// TestProgrammesIndexOffersExactlyTheCoveredThemes: the theme navigator must
// offer every dimension the programmes cover, and nothing else. Comparing
// against the set derived from the data, rather than spot-checking one slug,
// is what gives this teeth: dropping the filter lets a theme with no
// programme points through, and the sets then differ.
func TestProgrammesIndexOffersExactlyTheCoveredThemes(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/partier", nil))
	body := w.Body.String()

	got := map[string]bool{}
	for _, m := range regexp.MustCompile(`href="/partier\?tema=([a-z0-9-]+)"`).FindAllStringSubmatch(body, -1) {
		if got[m[1]] {
			t.Errorf("theme %q linked more than once", m[1])
		}
		got[m[1]] = true
	}

	want := map[string]bool{}
	for _, prog := range data.Store.Programs {
		for _, pt := range prog.Points {
			if pt.Dimension != "" {
				want[dimensionSlug(pt.Dimension)] = true
			}
		}
	}
	if len(want) < 9 {
		t.Fatalf("only %d dimensions have programme points, expected at least 9", len(want))
	}
	for s := range want {
		if !got[s] {
			t.Errorf("a dimension with programme points is not offered: %q", s)
		}
	}
	for s := range got {
		if !want[s] {
			t.Errorf("a theme is offered that has no programme points: %q", s)
		}
	}
}

// TestThemeViewShowsEveryPartyOnOneDimension: ?tema= must narrow to exactly
// one dimension per party, so the parties can be read against each other,
// and it must not quietly drop a party that promises nothing on the theme.
func TestThemeViewShowsEveryPartyOnOneDimension(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/partier?tema=barns-muligheter", nil))
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Viser bare temaet") {
		t.Error("the focused view does not say which theme is shown")
	}
	sections := len(regexp.MustCompile(`id="parti-`).FindAllString(body, -1))
	if sections != len(data.Store.Parties) {
		t.Errorf("focused view has %d party sections, want %d", sections, len(data.Store.Parties))
	}
	groups := strings.Count(body, `class="program-group"`)
	if groups == 0 || groups > len(data.Store.Parties) {
		t.Errorf("focused view has %d dimension groups, want 1..%d (one per party at most)", groups, len(data.Store.Parties))
	}
	// No other dimension may appear in a focused view.
	for _, other := range []string{"Demokratisk åpenhet", "Norsk handlefrihet"} {
		if strings.Contains(body, other) && !strings.Contains(body, ">Barns muligheter<") {
			t.Errorf("focused view on Barns muligheter also shows %q", other)
		}
	}
}

// TestUnknownThemeFallsBackAndSaysSo: a theme the programmes do not cover
// (Finansiering av fellesskapet has statements but no programme points) must
// render the full list and name the miss, never render an empty page.
func TestUnknownThemeFallsBackAndSaysSo(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/partier?tema=finansiering-av-fellesskapet", nil))
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Fant ikke det temaet") {
		t.Error("an unknown theme is not reported")
	}
	if !strings.Contains(body, "Norsk handlefrihet") {
		t.Error("an unknown theme did not fall back to the full list")
	}
}

// TestOmPublishesEveryCorrection: /om's corrections log must show every
// entry in the dataset, newest first, and must not send a reader to a file
// in the (private) repository. A log that renders one entry and silently
// drops the rest is the failure mode this exists to catch: an empty or
// truncated log reads as "nothing to report", which is a stronger claim
// than "we have nothing to report".
func TestOmPublishesEveryCorrection(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/om", nil))
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	body := w.Body.String()

	corrections := data.Store.Corrections
	if len(corrections) == 0 {
		t.Fatal("dataset has no corrections; this test would pass vacuously")
	}

	if got := strings.Count(body, `class="correction correction-`); got != len(corrections) {
		t.Errorf("/om renders %d correction entries, want %d", got, len(corrections))
	}
	for _, c := range corrections {
		if !strings.Contains(body, html.EscapeString(c.Title)) {
			t.Errorf("%s: title is not rendered on /om: %q", c.ID, c.Title)
		}
		if !strings.Contains(body, dateNO(c.Date)) {
			t.Errorf("%s: date %s is not rendered on /om", c.ID, dateNO(c.Date))
		}
	}

	// Newest first: the newest entry's title must appear before the oldest.
	if len(corrections) > 1 {
		newest := strings.Index(body, html.EscapeString(corrections[0].Title))
		oldest := strings.Index(body, html.EscapeString(corrections[len(corrections)-1].Title))
		if newest < 0 || oldest < 0 || newest > oldest {
			t.Errorf("corrections are not rendered newest first (%d vs %d)", newest, oldest)
		}
	}

	// The log must be reachable in the page, and must not point at a file a
	// reader cannot open while the repository is private.
	if !strings.Contains(body, `id="rettelseslogg"`) {
		t.Error("the corrections log has no #rettelseslogg anchor")
	}
	if strings.Contains(body, "docs/roadmap.md") {
		t.Error("/om still points readers at docs/roadmap.md, which is inside the private repository")
	}
}
