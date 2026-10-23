package server

import "net/http"

// withSecurityHeaders: no cookies, no third-party scripts, no inline
// event handlers — the quiz's own JS lives in /static/js and is loaded
// by <script src>, so a strict CSP costs nothing here. 'unsafe-inline'
// is granted only for the one small JSON data block index.html embeds
// (see server.go) — that block is our own generated data, never user
// input, so this does not open an XSS path.
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self'; img-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer-when-downgrade")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		// Cloudflare terminates TLS in front of us; HSTS is set at the edge.
		next.ServeHTTP(w, r)
	})
}
