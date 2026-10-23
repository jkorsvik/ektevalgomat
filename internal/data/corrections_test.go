package data

import (
	"io/fs"
	"path"
	"regexp"
	"strings"
	"testing"
	"time"
)

// The corrections log is a publication, not a scratchpad: /om prints every
// entry to a reader. These tests are the gate on the file set itself, so a
// half-written entry cannot reach the page just because the YAML parses.

func TestCorrectionsLogIsNotEmpty(t *testing.T) {
	corrections := Store.Corrections
	if len(corrections) == 0 {
		t.Fatal("corrections log is empty; /om would claim this project never corrected itself")
	}
	// Print the resolved list: a gate that passes over zero rows is worse
	// than no gate, and this line is what shows the rows were real.
	for _, c := range corrections {
		t.Logf("%s  %-8s  %s", c.Date, c.Kind, c.ID)
	}
}

func TestEveryCorrectionIsComplete(t *testing.T) {
	statementNumber := regexp.MustCompile(`^[0-9]{2}$`)
	for _, c := range Store.Corrections {
		if c.ID == "" {
			t.Error("correction with empty id")
			continue
		}
		if _, err := time.Parse("2006-01-02", c.Date); err != nil {
			t.Errorf("%s: date %q is not YYYY-MM-DD", c.ID, c.Date)
		}
		if !ValidCorrectionKind(c.Kind) {
			t.Errorf("%s: kind %q is not rettelse or nesten", c.ID, c.Kind)
		}
		for _, field := range []struct{ name, value string }{
			{"title", c.Title},
			{"what", c.What},
			{"reason", c.Reason},
			{"effect", c.Effect},
		} {
			if strings.TrimSpace(field.value) == "" {
				t.Errorf("%s: %s is empty; every entry must say what changed, why, and what it meant", c.ID, field.name)
			}
		}
		for _, n := range c.Statements {
			if !statementNumber.MatchString(n) {
				t.Errorf("%s: statement %q is not a two-digit statement number", c.ID, n)
			}
		}
	}
}

func TestCorrectionsAreNewestFirst(t *testing.T) {
	corrections := Store.Corrections
	if len(corrections) < 2 {
		t.Skipf("only %d correction(s); nothing to order", len(corrections))
	}
	for i := 1; i < len(corrections); i++ {
		prev, cur := corrections[i-1], corrections[i]
		if prev.Date < cur.Date {
			t.Errorf("out of order at %d: %s (%s) comes before %s (%s); newest must be first",
				i, prev.ID, prev.Date, cur.ID, cur.Date)
		}
		if prev.Date == cur.Date && prev.ID > cur.ID {
			t.Errorf("same-day entries out of order at %d: %s before %s; the roadmap's own chronological order is the tiebreak",
				i, prev.ID, cur.ID)
		}
	}
}

func TestEveryCorrectionFileIsLoaded(t *testing.T) {
	entries, err := fs.ReadDir(contentFS, path.Join(contentRoot, "corrections"))
	if err != nil {
		t.Fatalf("reading corrections dir: %v", err)
	}
	files := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		files++
	}
	if files != len(Store.Corrections) {
		t.Fatalf("%d correction file(s) on disk but %d loaded; one is not reaching /om",
			files, len(Store.Corrections))
	}
	if files == 0 {
		t.Fatal("no correction files found; the directory path in the test and the loader have drifted apart")
	}
}

func TestCorrectionIDsMatchTheirFilenames(t *testing.T) {
	// The id is what the page and the tests refer to; the filename is what
	// a human edits. A copy-pasted entry keeps the old stem, so this is the
	// check that catches the entry whose id says it is a different one.
	byStem := map[string]Correction{}
	entries, err := fs.ReadDir(contentFS, path.Join(contentRoot, "corrections"))
	if err != nil {
		t.Fatalf("reading corrections dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		byStem[strings.TrimSuffix(e.Name(), ".yaml")] = Correction{}
	}
	for _, c := range Store.Corrections {
		if _, ok := byStem[c.ID]; !ok {
			t.Errorf("%s: id does not match any filename in content/corrections/", c.ID)
		}
	}
}
