package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The sitemap (2026-10-13). The site has 77 public pages and had no sitemap,
// so a crawler had to find them by following links. Every page is reachable
// by link, so nothing was hidden, but the statement pages alone are 60 URLs
// that only the front page links to.
//
// Not listed: /resultat, which shows nothing until the reader has answered
// in their own browser, and /healthz, which is for the deployment. Listing a
// page that renders empty to a crawler is a claim the page cannot keep.
//
// robots.txt is deliberately not served from here. Cloudflare serves one
// with a copyright content-signals reservation, and replacing it would drop
// that reservation. The Sitemap line belongs there instead.
func sitemap(w http.ResponseWriter, r *http.Request, ds *data.Dataset) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	add := func(path string) {
		b.WriteString("  <url><loc>" + siteBase + path + "</loc></url>\n")
	}
	for _, p := range []string{"/", "/partier", "/budsjett", "/metode", "/om", "/personvern"} {
		add(p)
	}
	for _, s := range ds.Statements {
		add("/utsagn/" + strconv.Itoa(s.Number))
	}
	for _, p := range ds.Parties {
		add("/parti/" + p.ID)
	}
	b.WriteString("</urlset>\n")

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	if _, err := w.Write([]byte(b.String())); err != nil {
		return
	}
}
