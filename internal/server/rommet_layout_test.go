package server

import (
	"strings"
	"testing"
)

func TestRommetPutsTheTaskBeforeTheResultsInLinearDocumentFlow(t *testing.T) {
	body := getRommet(t)
	for _, want := range []string{
		`<main class="rom-comparison">`,
		`<section class="rom-finder"`,
		`<div class="rom-positions"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the linear comparison layout is missing %q", want)
		}
	}
	formAt := strings.Index(body, `<form class="rom-controls"`)
	resultsAt := strings.Index(body, `<section class="rom-results"`)
	if formAt < 0 || resultsAt < 0 || formAt > resultsAt {
		t.Errorf("the task must precede the results: form=%d results=%d", formAt, resultsAt)
	}
	for _, unwanted := range []string{`class="rom-axes"`, `class="rom-canvas-wrap"`, `class="rom-spectrum-wrap"`} {
		if strings.Contains(body, unwanted) {
			t.Errorf("an absolute or canvas-specific layout survived: %q", unwanted)
		}
	}
}
