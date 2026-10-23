package web

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// A template is not JSX, and the two look close enough to mix up: writing
// {" "} to mean a non-breaking space, or {expr} to mean an interpolation,
// passes every build and every HTML assertion while rendering the braces to
// the reader. One of them reached production on /data and was only caught by
// reading the live page as text, which is not a gate.
//
// So the gate is here: no artifact from another templating syntax may sit in
// a shipped template.
func TestTemplatesCarryNoForeignTemplateSyntax(t *testing.T) {
	// Each pattern is a construct that would render literally to a reader.
	forbidden := []struct {
		name    string
		pattern *regexp.Regexp
		why     string
	}{
		{"jsx-space", regexp.MustCompile(`\{" "\}`), `renders the braces; use a plain space`},
		// Matched below with a look at the preceding byte: RE2 has no
		// lookbehind, and the tail of {{end}} looks exactly like {end}.
		{"jsx-expression", regexp.MustCompile(`\{[a-zA-Z_][a-zA-Z0-9_.]*\}`), `JSX interpolation; Go templates use {{.Field}}`},
		{"mustache-unescaped", regexp.MustCompile(`\{&[a-zA-Z]`), `Mustache syntax; not a Go template`},
	}

	checked := 0
	err := fs.WalkDir(Templates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return nil
		}
		b, err := fs.ReadFile(Templates, path)
		if err != nil {
			return err
		}
		checked++
		for _, f := range forbidden {
			for _, m := range f.pattern.FindAllIndex(b, -1) {
				// A Go template action ends in }} and a JS template literal
				// opens with ${: both contain {word}, and neither is the bug.
				if m[0] > 0 {
					prev := b[m[0]-1]
					if prev == '{' || prev == '$' {
						continue
					}
				}
				line := 1 + strings.Count(string(b[:m[0]]), "\n")
				t.Errorf("%s:%d: %s (%s): %q", path, line, f.name, f.why, b[m[0]:m[1]])
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking templates: %v", err)
	}
	// One template would be enough for the loop to pass while the estate went
	// unread, which is the failure this guard exists to prevent.
	if checked < 8 {
		t.Fatalf("checked only %d templates: expected the whole set", checked)
	}
}

// The same mistake reached a reader because nothing read the rendered text.
// This asserts on output, not source, for the one artifact already seen.
func TestRenderedPagesShowNoRawBraces(t *testing.T) {
	// Pages are rendered in the server package against the real dataset; this
	// guard is the template half, kept next to the templates themselves.
	err := fs.WalkDir(Templates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return err
		}
		b, _ := fs.ReadFile(Templates, path)
		if strings.Contains(string(b), `{" "}`) {
			t.Errorf("%s still contains the literal {\" \"}", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking templates: %v", err)
	}
}
