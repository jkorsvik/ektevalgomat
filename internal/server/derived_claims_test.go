package server

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// pageText renders a path through the real handler and returns the body with
// whitespace collapsed, so a test can assert on a sentence that the template
// breaks across lines.
func pageText(t *testing.T, path string) string {
	t.Helper()
	w := get(t, path)
	if w.Code != 200 {
		t.Fatalf("%s returned %d", path, w.Code)
	}
	return strings.Join(strings.Fields(w.Body.String()), " ")
}

// TestMetodeStatesProgrammeOnlyCells: the site's headline is that it does not
// compare an answer with a party's own programme. That headline is not true
// for every cell, so the number of cells where it fails is printed on the
// method page and held here to the dataset.
func TestMetodeStatesProgrammeOnlyCells(t *testing.T) {
	cells, coded := programmeOnlyCells(data.Store)
	if coded == 0 {
		t.Fatal("no coded cells; the figure would be vacuous")
	}
	if cells == 0 {
		t.Fatal("no programme-only cells: either the data changed or the counter stopped counting")
	}
	body := pageText(t, "/metode")
	want := fmt.Sprintf("%d av %d celler", cells, coded)
	if !strings.Contains(body, want) {
		t.Errorf("/metode does not state %q, so the page and the dataset disagree", want)
	}
	t.Logf("%d of %d coded cells rest on C/D sources only", cells, coded)
}

// TestFrontPageNamesUncomparableStatements: a question every party lacks a
// position on cannot change a result. The front page has to say which ones
// before the reader answers, not only the result panel afterwards.
func TestFrontPageNamesUncomparableStatements(t *testing.T) {
	noPos := uncomparableStatements(data.Store)
	if len(noPos) == 0 {
		t.Skip("every statement has at least one coded position; nothing to disclose")
	}
	body := pageText(t, "/")
	want := fmt.Sprintf("%d av dem har foreløpig ingen dokumentert partiposisjon", len(noPos))
	if !strings.Contains(body, want) {
		t.Errorf("/ does not state %q", want)
	}
	for _, s := range noPos {
		if !strings.Contains(body, fmt.Sprintf(`href="/utsagn/%d"`, s.Number)) {
			t.Errorf("/ does not link utsagn %d, which it just said cannot count", s.Number)
		}
	}
	om := pageText(t, "/om")
	if !strings.Contains(om, "ingen kodet posisjon for noen av partiene") {
		t.Error("/om does not disclose statements with no coded position anywhere")
	}
}
