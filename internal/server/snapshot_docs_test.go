package server

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The roadmap describes the snapshot badge in prose, and that sentence drifted
// once already: it still said 2026-09-23 while the constant had moved to
// 2026-10-12, and nothing read it. TestNoEvidenceAfterSnapshotDate enforces the
// cutoff on the data; this enforces it on the description of the cutoff, which is
// what a reader actually sees.
//
// The claim is parsed out of the document and compared with the constant, never
// against a stored copy of itself. If the sentence disappears, this fails too: a
// rewrite must consciously restate the date rather than drop it and leave the test
// with nothing to check.
func TestRoadmapStatesTheRealSnapshotDate(t *testing.T) {
	roadmap, err := os.ReadFile("../../docs/roadmap.md")
	if err != nil {
		t.Fatalf("read roadmap: %v", err)
	}
	// Collapse whitespace: the sentence wraps across lines, and a line break inside
	// a date is not a change of meaning.
	flat := strings.Join(strings.Fields(string(roadmap)), " ")

	claim := regexp.MustCompile(`hendelser til og med (\d{4}-\d{2}-\d{2})`)
	found := claim.FindAllStringSubmatch(flat, -1)
	if len(found) == 0 {
		t.Fatal("the roadmap no longer states the snapshot date; restate it or this test checks nothing")
	}
	for _, m := range found {
		if m[1] != SnapshotDate {
			t.Errorf("roadmap says the badge reads %q, the constant is %q", m[1], SnapshotDate)
		}
	}

	// The badge itself must use the same date, so the page and the description of
	// the page cannot part company.
	if !strings.Contains(flat, SnapshotDate) {
		t.Errorf("roadmap never mentions %q", SnapshotDate)
	}
}
