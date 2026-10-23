package web

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// TestDecorativeArrowsStayOutOfAccessibleNames keeps the link arrows out of
// what assistive tech reads.
//
// Every call-to-action on this site ends in an arrow: "Meld interesse →",
// "Personvern →", "Grunnlag ↓", "Start valgomaten ↓". They are decoration —
// the words already say what the link does — but written as plain text
// inside the <a> they become part of the link's accessible name, so a
// screen reader announced "Personvern, pil hoyre". Reported by two separate
// reviews on 2026-10-13 as a P3, and the kind of thing that silently
// regresses the next time someone adds a link.
//
// The rule: inside web/templates, an arrow glyph must sit in a span marked
// aria-hidden. Arrow glyphs in editorial DATA (budget/2027.yaml uses
// "159,11 → 0" in prose) are not templates and not links, so they are out of
// scope here; those need a wording fix, not markup.
func TestDecorativeArrowsStayOutOfAccessibleNames(t *testing.T) {
	arrow := regexp.MustCompile(`[→←↓↑]`)
	wrapped := regexp.MustCompile(`aria-hidden="true">[→←↓↑]`)
	// A CSS or comment line is not rendered markup.
	notMarkup := regexp.MustCompile(`(?m)^\s*(/\*|\*|//|\.|#)`)

	files, total, hidden := 0, 0, 0
	err := fs.WalkDir(Templates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return err
		}
		b, err := fs.ReadFile(Templates, path)
		if err != nil {
			return err
		}
		files++
		for _, line := range strings.Split(string(b), "\n") {
			if notMarkup.MatchString(line) {
				continue
			}
			// Ignore template comments, which quote glyphs in prose.
			if i := strings.Index(line, "{{/*"); i >= 0 {
				line = line[:i]
			}
			for _, loc := range arrow.FindAllStringIndex(line, -1) {
				total++
				// Allow any characters between the marker and the glyph so
				// a class attribute or a space does not defeat the check,
				// but require them on the same line and inside the span.
				before := line[:loc[0]]
				if wrapped.MatchString(before) || strings.HasSuffix(before, `aria-hidden="true">`) {
					hidden++
					continue
				}
				t.Errorf("%s: arrow %q is not aria-hidden, so it lands in the accessible name: %s",
					path, line[loc[0]:loc[1]], strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Non-vacuous: there are many arrows on this site, and the check is
	// worthless if it finds none (a changed file extension, a moved dir).
	if total < 8 {
		t.Fatalf("found only %d arrow glyphs across %d templates; the check is not reading the markup", total, files)
	}
	if hidden != total {
		t.Errorf("%d of %d arrow glyphs are aria-hidden; want all", hidden, total)
	}
	t.Logf("checked %d arrow glyphs across %d templates, all %d aria-hidden", total, files, hidden)
}
