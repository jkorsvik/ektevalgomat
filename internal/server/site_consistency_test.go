package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// The site had no page-level identity: all thirteen pages carried
// <h1 class="site-title">Ekte valgomat</h1>, so the first-level heading named the site
// on every page and no page named itself. /rommet then added its own h1 and had two.
// These guards keep the structure fixed.
//
// readerPages is every page a reader is meant to reach from the navigation. Utility
// and detail pages are deliberately absent: they are reached from a statement or a
// party, not from the menu.
func readerPages() []string {
	return []string{"/", "/resultat", "/partier", "/kart", "/stemmer", "/tidslinje",
		"/rommet", "/budsjett", "/penger", "/eierskap", "/makt", "/data", "/metode", "/om"}
}

func newTestServer() http.Handler {
	return New(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func bodyOf(t *testing.T, h http.Handler, path string) string {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	if w.Code != 200 {
		t.Fatalf("%s: status %d", path, w.Code)
	}
	return w.Body.String()
}

// TestEveryPageHasExactlyOneHeadingOfItsOwn: one h1, and it is not the site name.
func TestEveryPageHasExactlyOneHeadingOfItsOwn(t *testing.T) {
	h := newTestServer()
	pages := append(readerPages(), "/personvern", "/bidra")
	checked := 0
	for _, path := range pages {
		body := bodyOf(t, h, path)
		n := strings.Count(body, "<h1")
		if n != 1 {
			t.Errorf("%s: %d <h1> elements, want exactly 1", path, n)
			continue
		}
		checked++
		if strings.Contains(body, `<h1 class="site-title"`) {
			t.Errorf("%s: the site name is the h1; the page must name itself", path)
		}
		if !strings.Contains(body, `<p class="site-title">`) {
			t.Errorf("%s: the brand is not rendered as a paragraph link", path)
		}
		// The heading must be a real page name, not an empty element.
		m := regexp.MustCompile(`<h1[^>]*>([^<]*)</h1>`).FindStringSubmatch(body)
		if m == nil || len(strings.TrimSpace(m[1])) < 3 {
			t.Errorf("%s: the h1 is empty", path)
		}
	}
	if checked < 15 {
		t.Fatalf("only %d pages checked: this guard walked nothing", checked)
	}
	t.Logf("%d pages, each with one h1 of its own", checked)
}

// TestTheNavReachesEveryReaderPage: a page that exists but is not in the navigation is
// a page nobody finds. Before this, eight pages were linked only from a footer line.
func TestTheNavReachesEveryReaderPage(t *testing.T) {
	h := newTestServer()
	// The nav is rendered from navLinks, so check the definition against the routes a
	// reader is meant to reach, and then check it actually reaches the HTML.
	inNav := map[string]bool{}
	for _, l := range navLinks {
		inNav[l.Href] = true
	}
	for _, path := range readerPages() {
		if !inNav[path] {
			t.Errorf("%s is a reader page but not in navLinks", path)
		}
	}
	body := bodyOf(t, h, "/kart")
	m := regexp.MustCompile(`(?s)<nav[^>]*>(.*?)</nav>`).FindStringSubmatch(body)
	if m == nil {
		t.Fatal("no <nav> in the header")
	}
	reached := 0
	for _, path := range readerPages() {
		if strings.Contains(m[1], `href="`+path+`"`) {
			reached++
		}
	}
	if reached != len(readerPages()) {
		t.Errorf("the nav reaches %d of %d reader pages", reached, len(readerPages()))
	}
	// And it marks where the reader is, or it is only a list of links.
	if !strings.Contains(body, `aria-current="page"`) {
		t.Error("the nav does not mark the current page")
	}
	if !strings.Contains(body, `<a href="/kart" aria-current="page">`) {
		t.Error("the current page is not the one marked")
	}
	t.Logf("nav reaches all %d reader pages and marks the current one", reached)
}

// TestNoPagePrintsADotDecimal: Norwegian text uses a comma. /kart shipped 108 cells
// reading "-0.2" while every other figure on the site read "-0,2", which a reader
// reads as two different scales.
func TestNoPagePrintsADotDecimal(t *testing.T) {
	h := newTestServer()
	dot := regexp.MustCompile(`>[-−]?[0-9]+\.[0-9]+<`)
	pages := append(readerPages(), "/personvern", "/bidra", "/utsagn/52")
	scanned, cells := 0, 0
	for _, path := range pages {
		body := bodyOf(t, h, path)
		scanned++
		if found := dot.FindAllString(body, -1); len(found) > 0 {
			cells += len(found)
			t.Errorf("%s: %d cells print a dot decimal, e.g. %q", path, len(found), found[0])
		}
	}
	if scanned < 15 {
		t.Fatalf("only %d pages scanned: this guard proved nothing", scanned)
	}
	t.Logf("%d pages scanned, %d dot decimals", scanned, cells)
}
