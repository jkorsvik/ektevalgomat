package web

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// TestScriptsLoadTheirDependenciesFirst keeps the page from dying on a
// missing script.
//
// quiz.js and analyse.js read the visitor's saved choices through
// EkteStore (store.js) and the scoring contract through EkteScoring
// (scoring.js). Both are plain globals set by their own <script>, so a page
// that loads quiz.js without store.js does not degrade: quiz.js throws
// "EkteStore is not defined" on its first line that touches it, nothing
// renders, and nothing on the server notices. Go renders the page happily.
//
// Order matters as much as presence, because a script runs the moment it
// loads. So for every template that loads quiz.js or analyse.js this checks
// that each dependency appears earlier in the same file, and that the
// container the quiz writes its storage notes into is there.
func TestScriptsLoadTheirDependenciesFirst(t *testing.T) {
	src := regexp.MustCompile(`<script src="\{\{asset "/static/js/([a-z-]+)\.js"\}\}"></script>`)
	needs := map[string][]string{
		"quiz":      {"scoring", "store", "avveining"},
		"avveining": {"scoring"},
		"analyse":   {"scoring", "store"},
	}

	pages, scriptsSeen := 0, 0
	err := fs.WalkDir(Templates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return err
		}
		b, err := fs.ReadFile(Templates, path)
		if err != nil {
			return err
		}
		body := string(b)
		var order []string
		for _, m := range src.FindAllStringSubmatch(body, -1) {
			order = append(order, m[1])
		}
		scriptsSeen += len(order)
		position := map[string]int{}
		for i, name := range order {
			if _, dup := position[name]; !dup {
				position[name] = i
			}
		}
		for consumer, deps := range needs {
			at, uses := position[consumer]
			if !uses {
				continue
			}
			if consumer == "quiz" {
				pages++
				// A live region must already be in the page when its text
				// changes, so the container carries role="status" itself.
				if !strings.Contains(body, `id="storage-note"`) || !strings.Contains(body, `role="status"`) {
					t.Errorf("%s loads quiz.js but has no role=\"status\" #storage-note, so a failed save would be silent to a screen reader", path)
				}
			}
			for _, dep := range deps {
				depAt, ok := position[dep]
				switch {
				case !ok:
					t.Errorf("%s loads %s.js but never loads %s.js: the page would die with %q undefined", path, consumer, dep, dep)
				case depAt > at:
					t.Errorf("%s loads %s.js before %s.js: it runs before its dependency exists", path, consumer, dep)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Non-vacuous: the front page and the result page both run the quiz.
	if pages < 2 {
		t.Fatalf("found %d pages loading quiz.js, want at least 2 (/ and /resultat); "+
			"the pattern no longer matches the templates and this test proved nothing", pages)
	}
	t.Logf("checked %d pages that run the quiz, %d script tags", pages, scriptsSeen)
}
