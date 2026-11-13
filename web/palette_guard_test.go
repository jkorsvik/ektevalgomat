package web

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// Party colour is data, not decoration. Every party chip needs exactly one
// colour token and one dot class that uses it. The old version of this guard
// compared the CSS with the WebGL canvas; the canvas is gone, so the guard now
// protects the actual reader-facing contract instead of a retired consumer.
func TestEveryPartyHasOneColourTokenAndDotClass(t *testing.T) {
	css, err := fs.ReadFile(Static, "static/css/rommet.css")
	if err != nil {
		t.Fatalf("rommet.css: %v", err)
	}
	text := string(css)
	ids := []string{"ap", "frp", "h", "krf", "mdg", "rodt", "sp", "sv", "v"}
	for _, id := range ids {
		token := regexp.MustCompile(`--parti-` + id + `:\s*#[0-9a-fA-F]{6}`)
		if got := len(token.FindAllString(text, -1)); got != 1 {
			t.Errorf("%s has %d colour tokens, want exactly 1", id, got)
		}
		dot := ".rom-dot-" + id + " { background: var(--parti-" + id + "); }"
		if !strings.Contains(text, dot) {
			t.Errorf("party dot does not use its own token: %q", dot)
		}
	}
	allTokens := regexp.MustCompile(`--parti-[a-z]+:\s*#[0-9a-fA-F]{6}`).FindAllString(text, -1)
	if len(allTokens) != len(ids) {
		t.Errorf("found %d party colour tokens, want %d; an unnamed party could fall back to grey", len(allTokens), len(ids))
	}
}
