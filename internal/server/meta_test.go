package server

import (
	"log/slog"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// pagePaths is every page a crawler or a reader can reach, built from the
// data rather than written down, so a new statement or party is covered
// automatically.
func pagePaths(ds *data.Dataset) []string {
	paths := []string{"/", "/partier", "/budsjett", "/metode", "/om", "/personvern", "/resultat"}
	for _, s := range ds.Statements {
		paths = append(paths, "/utsagn/"+strconv.Itoa(s.Number))
	}
	for _, p := range ds.Parties {
		paths = append(paths, "/parti/"+p.ID)
	}
	return paths
}

// TestEveryPageHasShareableMetadata: before 2026-10-13 only two pages of
// seventy-odd carried a description, none carried a canonical link and none
// carried an og tag, so a shared link produced no preview card. This pins
// all four on every page.
func TestEveryPageHasShareableMetadata(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	paths := pagePaths(data.Store)
	if len(paths) < 70 {
		t.Fatalf("only %d pages, expected at least 70", len(paths))
	}
	seen := map[string]string{}
	for _, p := range paths {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", p, nil))
		if w.Code != 200 {
			t.Errorf("%s: status %d", p, w.Code)
			continue
		}
		body := w.Body.String()
		for _, want := range []string{`<title>`, `name="description"`, `rel="canonical"`, `og:title`, `og:description`} {
			if !strings.Contains(body, want) {
				t.Errorf("%s: missing %s", p, want)
			}
		}
		m := regexp.MustCompile(`rel="canonical" href="([^"]+)"`).FindStringSubmatch(body)
		if m == nil {
			continue
		}
		if strings.HasPrefix(m[1], "https://ektevalgomat.no") == false {
			t.Errorf("%s: canonical is not absolute: %q", p, m[1])
		}
		if prev, ok := seen[m[1]]; ok {
			t.Errorf("%s and %s share the canonical %q", prev, p, m[1])
		}
		seen[m[1]] = p
	}
	if len(seen) < 70 {
		t.Fatalf("only %d distinct canonicals, expected at least 70", len(seen))
	}
}

// TestDescriptionsAreReal: a description of a few characters passes the tag
// check while saying nothing, so this asserts length, and that no page
// describes itself with a different page's words.
func TestDescriptionsAreReal(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	byDesc := map[string]string{}
	checked := 0
	for _, p := range pagePaths(data.Store) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", p, nil))
		m := regexp.MustCompile(`name="description" content="([^"]*)"`).FindStringSubmatch(w.Body.String())
		if m == nil {
			t.Errorf("%s: no description", p)
			continue
		}
		d := m[1]
		if len([]rune(d)) < 40 {
			t.Errorf("%s: description is only %d characters: %q", p, len([]rune(d)), d)
		}
		if len([]rune(d)) > 200 {
			t.Errorf("%s: description is %d characters, too long to be shown", p, len([]rune(d)))
		}
		if prev, ok := byDesc[d]; ok && p != prev {
			t.Errorf("%s and %s share a description", prev, p)
		}
		byDesc[d] = p
		checked++
	}
	if checked < 70 {
		t.Fatalf("only checked %d descriptions", checked)
	}
	// Statement pages must describe the statement, not the site.
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/utsagn/60", nil))
	if !strings.Contains(w.Body.String(), "nødnett") {
		t.Error("utsagn/60's description does not mention the statement's subject")
	}
}

// TestSitemapListsEveryPublicPage: the sitemap must hold every page a crawler
// should index, and must not list the ones that render empty without a
// reader's own answers.
func TestSitemapListsEveryPublicPage(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/sitemap.xml", nil))
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "xml") {
		t.Errorf("content type is %q, want xml", ct)
	}
	body := w.Body.String()
	if !strings.HasPrefix(body, `<?xml version="1.0"`) || !strings.Contains(body, "</urlset>") {
		t.Fatal("the sitemap is not well-formed XML")
	}
	locs := regexp.MustCompile(`<loc>([^<]+)</loc>`).FindAllStringSubmatch(body, -1)
	if len(locs) < 70 {
		t.Fatalf("only %d URLs, expected at least 70", len(locs))
	}
	have := map[string]bool{}
	for _, m := range locs {
		have[m[1]] = true
	}
	for _, p := range pagePaths(data.Store) {
		if p == "/resultat" {
			continue // shows nothing until the reader answers, so not listed
		}
		if !have[siteBase+p] {
			t.Errorf("the sitemap does not list %s", p)
		}
	}
	if have[siteBase+"/resultat"] {
		t.Error("the sitemap lists /resultat, which renders empty to a crawler")
	}
	if have[siteBase+"/sitemap.xml"] {
		t.Error("the sitemap lists itself")
	}
}
