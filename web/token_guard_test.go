package web

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// TestEveryColourComesFromAToken makes the brand rule a tested property of
// the shipped CSS, not just a comment. The rule, written in style.css's
// :root during the 2026-10-02 "Civic Ledger" pass, says red and green stay
// out entirely because they read as a verdict.
//
// It drifted anyway. The 2027 budget page (2026-10-12) coloured its change
// column green and red, and nothing noticed, because a rule with no gate is
// a comment and comments do not fail. This is that gate: every colour is
// either a token in :root, or a documented exception.
//
// The one exception is #000 inside a mask gradient, where the value is an
// alpha channel rather than a colour the reader ever sees.
func TestEveryColourComesFromAToken(t *testing.T) {
	hex := regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b`)
	rootBlock := regexp.MustCompile(`(?s):root\s*\{.*?\}`)

	tokens := 0
	files := 0
	err := fs.WalkDir(Static, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".css") {
			return err
		}
		b, err := fs.ReadFile(Static, path)
		if err != nil {
			return err
		}
		files++
		body := string(b)
		tokens += len(hex.FindAllString(strings.Join(rootBlock.FindAllString(body, -1), "\n"), -1))

		// Skip every :root block; those are where colours are defined.
		// 2026-10-23: the dark theme added a second :root block inside
		// @media (prefers-color-scheme: dark). Reading only the first one
		// would have made the gate fail on any legitimate second theme,
		// which is not the rule it enforces — the rule is "every colour
		// is a token", not "there is exactly one block of tokens".
		rooted := rootBlock.ReplaceAllStringFunc(body, func(block string) string {
			return strings.Repeat("\n", strings.Count(block, "\n"))
		})
		for _, line := range strings.Split(rooted, "\n") {
			if !hex.MatchString(line) {
				continue
			}
			// A mask gradient's #000 is an alpha channel, not a colour.
			if strings.Contains(line, "mask") {
				continue
			}
			t.Errorf("%s: colour not from a token: %s", path, strings.TrimSpace(line))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Non-vacuity: the walk must actually have read the stylesheets, and the
	// token block must actually define colours, or this test proves nothing.
	if files == 0 {
		t.Fatal("no stylesheets were read, so the check proved nothing")
	}
	if tokens < 10 {
		t.Fatalf("only %d colour tokens found in :root, expected at least 10", tokens)
	}
}

// TestEverySizeComesFromAToken: the same rule for type and space. Before the
// 2026-10-13 scale pass the stylesheets carried 25 font sizes and 24 spacing
// values, and most of the difference between them was accidental: 0.8, 0.82
// and 0.85 meant one thing written three ways. Every one is now a step in
// :root, and this keeps a new raw value from creeping back in.
//
// Declarations that take no length (margin: 0 auto, gap: 0) are fine; the
// check only looks for a rem value.
func TestEverySizeComesFromAToken(t *testing.T) {
	rootBlock := regexp.MustCompile(`(?s):root\s*\{.*?\}`)
	sizeDecl := regexp.MustCompile(`(?m)(?:^|;|\{)\s*(font-size|margin|padding|gap)(?:-[a-z]+)?\s*:\s*([^;{}]+)`)
	rem := regexp.MustCompile(`[0-9.]+rem`)

	checked := 0
	files := 0
	err := fs.WalkDir(Static, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".css") {
			return err
		}
		b, err := fs.ReadFile(Static, path)
		if err != nil {
			return err
		}
		files++
		body := rootBlock.ReplaceAllString(string(b), "")
		for _, m := range sizeDecl.FindAllStringSubmatch(body, -1) {
			if !rem.MatchString(m[2]) {
				continue
			}
			checked++
			t.Errorf("%s: %s uses a raw rem value: %s", path, m[1], strings.TrimSpace(m[2]))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files == 0 {
		t.Fatal("no stylesheets were read, so the check proved nothing")
	}
	if checked != 0 {
		t.Fatalf("%d declarations still use a raw length", checked)
	}
}

// TestDarkThemeDefinesEveryToken keeps the second token block honest. A
// dark theme that forgets a token silently inherits the light value for
// that one property (a white card on a dark page, an accent that vanishes
// into the surface), and nothing else in the suite would notice. So: the
// dark block must define exactly the same custom-property names as the
// light one — no missing token, and no token invented only for dark.
func TestDarkThemeDefinesEveryToken(t *testing.T) {
	b, err := fs.ReadFile(Static, "static/css/style.css")
	if err != nil {
		t.Fatalf("read style.css: %v", err)
	}
	rootBlocks := regexp.MustCompile(`(?s):root\s*\{(.*?)\}`).FindAllStringSubmatch(string(b), -1)
	if len(rootBlocks) != 2 {
		t.Fatalf("found %d :root blocks, want 2 (light and prefers-color-scheme dark)", len(rootBlocks))
	}
	// A token is a COLOUR token when its light value is a hex literal.
	// Type, space, family and width tokens are theme-independent on
	// purpose, so the dark block must not repeat them; it must repeat
	// every colour, exactly.
	tokenValue := regexp.MustCompile(`(?m)^\s*(--[a-z0-9-]+)\s*:\s*([^;]+);`)
	hexValue := regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b`)
	values := func(block string) map[string]string {
		out := map[string]string{}
		for _, m := range tokenValue.FindAllStringSubmatch(block, -1) {
			out[m[1]] = strings.TrimSpace(m[2])
		}
		return out
	}
	light, dark := values(rootBlocks[0][1]), values(rootBlocks[1][1])
	colourTokens := 0
	for name, value := range light {
		if !hexValue.MatchString(value) {
			continue
		}
		colourTokens++
		dv, ok := dark[name]
		if !ok {
			t.Errorf("dark theme does not define %s, so it silently inherits the light colour %s", name, value)
			continue
		}
		if !hexValue.MatchString(dv) {
			t.Errorf("dark theme defines %s as %q, which is not a hex colour", name, dv)
		}
	}
	if colourTokens < 8 {
		t.Fatalf("only %d colour tokens found in the light theme; the check would prove nothing", colourTokens)
	}
	for name := range dark {
		if _, ok := light[name]; !ok {
			t.Errorf("dark theme defines %s, which the light theme does not; the two themes have drifted", name)
		}
	}
	t.Logf("light %d colour tokens, dark %d tokens", colourTokens, len(dark))
}
