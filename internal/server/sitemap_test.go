package server

import (
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestSitemapCoversEveryNavPage ties the sitemap to the navigation. The
// hardcoded list it replaced had silently lost eight pages, so a page could be
// reachable by link and still invisible to a crawler.
func TestSitemapCoversEveryNavPage(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/sitemap.xml", nil))
	body := w.Body.String()

	for _, l := range navLinks {
		if l.Href == "/resultat" {
			continue
		}
		if !strings.Contains(body, "<loc>"+siteBase+l.Href+"</loc>") {
			t.Errorf("sitemap is missing %s", l.Href)
		}
	}
	for _, p := range []string{"/personvern", "/programmer", "/personer"} {
		if !strings.Contains(body, "<loc>"+siteBase+p+"</loc>") {
			t.Errorf("sitemap is missing %s", p)
		}
	}
	if strings.Contains(body, "/resultat") {
		t.Error("sitemap lists /resultat, which renders empty to a crawler")
	}
}
