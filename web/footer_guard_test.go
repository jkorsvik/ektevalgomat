package web

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// TestEveryFooterLinksTheDataPage. The dataset is only useful to a reader who
// can find it, and the footer is where a reader looks for licence, source and
// downloads. When /data was added it was reachable from two pages, and the
// footers had drifted into three different shapes: five templates with the
// "Kildekoden er åpen" line, two with a one-sentence line, and one page with
// no footer at all. This holds them together: if a page shows a footer, that
// footer offers the data.
//
// /data itself is exempt: a page linking to itself in its own footer is noise.
func TestEveryFooterLinksTheDataPage(t *testing.T) {
	footer := regexp.MustCompile(`(?s)<footer.*?</footer>`)
	checked := 0

	err := fs.WalkDir(Templates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return err
		}
		if path == "templates/data.html.tmpl" {
			return nil
		}
		b, err := fs.ReadFile(Templates, path)
		if err != nil {
			return err
		}
		m := footer.Find(b)
		if m == nil {
			return nil // no footer: nothing to hold to this rule
		}
		checked++
		if !strings.Contains(string(m), `href="/data"`) {
			t.Errorf("%s: footer does not link /data", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// A guard that walks the wrong directory passes by finding nothing.
	if checked == 0 {
		t.Fatal("checked no footers: the walk found no templates")
	}
	t.Logf("checked %d footers", checked)
}

// TestDataPageIsLinkedFromTheNavsNeighbours is deliberately absent: the header
// nav holds six items and already wraps to two rows at 390px, so a seventh is
// a layout decision, not a test's.
