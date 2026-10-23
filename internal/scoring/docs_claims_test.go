package scoring

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The README and roadmap open with hard numbers a visitor reads before
// anything else: how many statements, how many evidence files, how many
// statements are fully coded. Every one of those is derivable from the
// dataset, and every one of them was wrong when this test was written:
// the README said 613 evidence files against 615, 53 fully coded
// statements against 54, and the roadmap still said 58 statements and
// 608 files, two statements behind. Prose drifts silently because
// nothing reads it.
//
// So nothing here compares a number against a stored copy of itself.
// Each claim is parsed out of the document and compared against the
// dataset, which means the document cannot disagree with the thing it
// describes. A claim that disappears from the document fails too: a
// rewrite must consciously re-state these numbers, not drop them and
// leave this test with nothing to check.
func TestDocsStateTheRealCounts(t *testing.T) {
	ds := data.Store
	statements := len(ds.Statements)
	evidenceFiles := len(ds.Evidence)

	full := 0
	for _, s := range ds.Statements {
		coded := 0
		for _, p := range ds.Parties {
			if ComputeCell(ds, s.ID, p.ID).PartyPosition != nil {
				coded++
			}
		}
		if coded == len(ds.Parties) {
			full++
		}
	}

	quoted := 0
	for _, e := range ds.Evidence {
		if e.VerbatimExcerpt != "" {
			quoted++
		}
	}

	type claim struct {
		file    string
		pattern string
		want    []int
	}
	claims := []claim{
		{"README.md", `Current state \(\d{4}-\d{2}-\d{2}\): (\d+) statements, (\d+) evidence\s+files`,
			[]int{statements, evidenceFiles}},
		{"README.md", `(\d+) of (\d+) statements have a coded position`,
			[]int{full, statements}},
		{"README.md", `(\d+) of\s+the (\d+) evidence rows carry such an excerpt`,
			[]int{quoted, evidenceFiles}},
		{"docs/roadmap.md", `\*\*(\d+) statements, (\d+) evidence files\.\*\*`,
			[]int{statements, evidenceFiles}},
		{"docs/roadmap.md", `\*\*(\d+) of (\d+) statements have a coded position`,
			[]int{full, statements}},
	}

	root := filepath.Join("..", "..")
	checkedNumbers := 0
	for _, c := range claims {
		path := filepath.Join(root, c.file)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s: %v (the Dockerfile must copy this file into the build stage)", c.file, err)
		}
		re := regexp.MustCompile(c.pattern)
		m := re.FindStringSubmatch(string(b))
		if m == nil {
			t.Errorf("%s: the claim %q is gone from the document; re-state the numbers rather than dropping them",
				c.file, c.pattern)
			continue
		}
		for i, want := range c.want {
			got, err := strconv.Atoi(m[i+1])
			if err != nil {
				t.Errorf("%s: %q in %q is not a number", c.file, m[i+1], c.pattern)
				continue
			}
			checkedNumbers++
			if got != want {
				t.Errorf("%s: says %d where the dataset has %d (%s)", c.file, got, want, c.pattern)
			}
		}
	}
	// A gate that parsed nothing must not report success.
	if checkedNumbers == 0 {
		t.Fatal("checked no numbers: every pattern failed to match, which is not a pass")
	}
	t.Logf("checked %d documented numbers against the dataset (%d statements, %d evidence, %d fully coded, %d quoted)",
		checkedNumbers, statements, evidenceFiles, full, quoted)
}
