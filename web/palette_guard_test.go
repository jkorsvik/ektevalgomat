package web

import (
	"fmt"
	"io/fs"
	"math"
	"regexp"
	"strconv"
	"testing"
)

// The nine party colours live in two files: static/css/rommet.css paints the dot
// beside each party in the table, static/js/rom.js paints the cloud that party owns
// on the canvas. If they drift, the page tells a reader that a colour is one party
// in the table and a different one in the drawing, and nothing else would notice:
// both files are valid, the page renders, and the error is entirely visual.
//
// So the two are parsed and compared. This is the same rule the DOM-id guard applies
// to ids: when a value must mean one thing in two places, a test owns it.
func TestPartyColoursMatchBetweenTableAndCanvas(t *testing.T) {
	css, err := fs.ReadFile(Static, "static/css/rommet.css")
	if err != nil {
		t.Fatalf("rommet.css: %v", err)
	}
	js, err := fs.ReadFile(Static, "static/js/rom.js")
	if err != nil {
		t.Fatalf("rom.js: %v", err)
	}

	fromCSS := map[string]string{}
	// The colours are tokens (style.css's brand rule, gated by token_guard_test.go),
	// so read them where they are defined.
	for _, m := range regexp.MustCompile(`--parti-([a-z]+):\s*#([0-9a-fA-F]{6})`).
		FindAllStringSubmatch(string(css), -1) {
		fromCSS[m[1]] = "#" + m[2]
	}
	if len(fromCSS) == 0 {
		t.Fatal("no party colours found in rommet.css: the guard would pass on nothing")
	}

	// ap: [0.90, 0.10, 0.29] -> #e61a4a, rounded the same way on both sides.
	fromJS := map[string]string{}
	re := regexp.MustCompile(`([a-z]+):\s*\[\s*([0-9.]+),\s*([0-9.]+),\s*([0-9.]+)\s*\]`)
	for _, m := range re.FindAllStringSubmatch(string(js), -1) {
		var rgb [3]float64
		for i := 0; i < 3; i++ {
			v, err := strconv.ParseFloat(m[i+2], 64)
			if err != nil {
				t.Fatalf("%s: %v", m[1], err)
			}
			rgb[i] = v
		}
		fromJS[m[1]] = fmt.Sprintf("#%02x%02x%02x",
			int(math.Round(rgb[0]*255)), int(math.Round(rgb[1]*255)), int(math.Round(rgb[2]*255)))
	}
	if len(fromJS) == 0 {
		t.Fatal("no party colours found in rom.js: the canvas would fall back to grey")
	}

	if len(fromCSS) != len(fromJS) {
		t.Errorf("css defines %d party colours, js defines %d", len(fromCSS), len(fromJS))
	}
	for id, want := range fromCSS {
		got, ok := fromJS[id]
		if !ok {
			t.Errorf("%s has a colour in the table but not on the canvas", id)
			continue
		}
		if got != want {
			t.Errorf("%s is %s in the table and %s on the canvas: the same party would be "+
				"two different colours on one page", id, want, got)
		}
	}
	t.Logf("%d party colours agree between rommet.css and rom.js", len(fromCSS))
}
