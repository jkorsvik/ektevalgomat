package web

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestStaticJSNeverCallsNetwork makes the site's central trust claim
// ("svarene dine lagres kun i nettleseren din — ingenting sendes til noen
// server") a tested property of the shipped JS, not just copy on the page.
// A design review of the live site pointed out this gap: the HTML has no
// visible third-party script, but HTML alone can't prove quiz.js never
// makes a network call — only reading (or, better, mechanically scanning)
// the actual script can.
//
// To intentionally allow a network call in future (e.g. an opt-in
// analytics feature), add a comment on the same source line naming why:
// `fetch(...)  // network-allowlist: <reason>` — the regex below skips any
// match whose line contains that marker.
func TestStaticJSNeverCallsNetwork(t *testing.T) {
	forbidden := regexp.MustCompile(`\b(fetch|XMLHttpRequest|sendBeacon|WebSocket|EventSource)\s*\(|new\s+(XMLHttpRequest|WebSocket|EventSource)\b`)
	allowlistMarker := "network-allowlist:"

	err := fs.WalkDir(Static, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".js" {
			return nil
		}
		b, err := fs.ReadFile(Static, path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(b), "\n") {
			if forbidden.MatchString(line) && !strings.Contains(line, allowlistMarker) {
				t.Errorf("%s:%d: found a network-call API with no allowlist marker: %q\n"+
					"this file ships on a page that promises \"ingenting sendes til noen "+
					"server\" — if this call is intentional, add `// %s <reason>` on this "+
					"line; otherwise remove it", path, i+1, strings.TrimSpace(line), allowlistMarker)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking embedded static assets: %v", err)
	}
}
