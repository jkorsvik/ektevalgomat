package server

import (
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// /budsjett's party section is the only place the party budgets appear, so these
// tests hold what the page claims: every area has a row, the joint column is not
// a party, a missing column stays missing instead of becoming a zero, and the
// printed figure is the one the dataset holds.

func TestPartyBudgetSectionCoversEveryArea(t *testing.T) {
	s := buildPartyBudgetSection(&data.Dataset{})
	if len(s.Rows) == 0 {
		t.Fatal("no rows: the check would pass on nothing")
	}
	if got, want := len(s.Rows), len(data.PartyBudgets()); got != want {
		t.Errorf("rows = %d, want one per table (%d)", got, want)
	}
	seen := map[int]bool{}
	for _, r := range s.Rows {
		if seen[r.Area] {
			t.Errorf("rammeområde %d appears twice", r.Area)
		}
		seen[r.Area] = true
		if len(r.Cells) != len(s.Cols) {
			t.Errorf("rammeområde %d has %d cells for %d columns: the row and the header "+
				"would not line up", r.Area, len(r.Cells), len(s.Cols))
		}
		if r.Posts == 0 {
			t.Errorf("rammeområde %d shows zero bevilgninger", r.Area)
		}
	}
	if s.Posts == 0 || s.Areas == 0 {
		t.Errorf("counts are empty: %d areas, %d posts", s.Areas, s.Posts)
	}
	t.Logf("%d rows, %d columns, %d areas, %d posts, session %s",
		len(s.Rows), len(s.Cols), s.Areas, s.Posts, s.Session)
}

func TestPartyBudgetSectionKeepsTheJointColumnApart(t *testing.T) {
	s := buildPartyBudgetSection(&data.Dataset{})
	if len(s.Cols) == 0 {
		t.Fatal("no columns")
	}
	if !s.Cols[0].Joint {
		t.Errorf("first column is %q, want the joint forlik: a forlik read as a party "+
			"would report several parties' agreement as one party's money", s.Cols[0].Party)
	}
	parties := 0
	for _, c := range s.Cols {
		if c.Party == "forlik" && !c.Joint {
			t.Error("the forlik column is not marked joint")
		}
		if !c.Joint {
			if !knownPartiesForServer[c.Party] {
				t.Errorf("unknown party column %q", c.Party)
			}
			parties++
		}
	}
	if parties < 3 {
		t.Errorf("%d party columns, want at least 3", parties)
	}
}

var knownPartiesForServer = map[string]bool{
	"ap": true, "sv": true, "sp": true, "r": true, "mdg": true,
	"frp": true, "h": true, "krf": true, "v": true,
}

// A party with no column in an area proposed nothing there. Rendering that as
// "0,0" would state a proposal the party never made, so the cell must be absent.
func TestPartyBudgetSectionLeavesMissingColumnsEmpty(t *testing.T) {
	s := buildPartyBudgetSection(&data.Dataset{})
	missing := 0
	for _, r := range s.Rows {
		for i, c := range r.Cells {
			if c.Kind != "fig" {
				missing++
				if c.Change != "" {
					t.Errorf("rammeområde %d column %s is not a figure but carries %q",
						r.Area, s.Cols[i].Party, c.Change)
				}
				if c.Bucket != 0 {
					t.Errorf("rammeområde %d column %s is not a figure but has bar bucket %d",
						r.Area, s.Cols[i].Party, c.Bucket)
				}
			}
		}
	}
	if missing == 0 {
		t.Fatal("no missing cell anywhere: coverage would be even, which the data says " +
			"it is not, so this test would prove nothing")
	}
	t.Logf("%d of %d cells are absent by design", missing, len(s.Rows)*len(s.Cols))
}

func TestPartyBudgetLedeStatesTheCoverageLimit(t *testing.T) {
	s := buildPartyBudgetSection(&data.Dataset{})
	lede := partyBudgetLede(s)
	for _, want := range []string{"avvik fra regjeringens forslag", "Forliket", "ikke at det foreslo null", "summeres det ikke"} {
		if !strings.Contains(lede, want) {
			t.Errorf("lede does not say %q:\n%s", want, lede)
		}
	}
}

// The printed figure must be the dataset's figure, formatted once. Recomputed here
// from the raw table so a formatting change cannot drift from the data.
func TestPartyBudgetCellMatchesTheDataset(t *testing.T) {
	tables := data.PartyBudgets()
	s := buildPartyBudgetSection(&data.Dataset{})
	checked := 0
	for i, tb := range tables {
		netto := tb.Netto()
		key := tb.PartyColumns()[0]
		m, ok := netto[key]
		if !ok {
			continue
		}
		// find the column index for that party
		party := tb.Columns[key].Party
		for ci, c := range s.Cols {
			if c.Party != party {
				continue
			}
			got := s.Rows[i].Cells[ci]
			if got.Kind != "fig" {
				t.Fatalf("rammeområde %d: %s has a column in the table but no cell on the page", tb.Area, party)
			}
			if want := mrd(m.Change); got.Change != want {
				t.Errorf("rammeområde %d %s = %q, want %q", tb.Area, party, got.Change, want)
			}
			checked++
			break
		}
	}
	if checked == 0 {
		t.Fatal("no cell was compared")
	}
	t.Logf("%d cells recomputed from the table", checked)
}

// The drill-down must be a real bevilgning list: every post identifiable by chapter
// and post number. The chapter is printed once on a heading row in the source, so
// without carrying it forward the table said "/1" and nothing else.
func TestPartyBudgetDrillDownNamesEveryPost(t *testing.T) {
	s := buildPartyBudgetSection(&data.Dataset{})
	var area, posts int
	for _, r := range s.Rows {
		if r.Posts > 5 {
			area = r.Area
			posts = r.Posts
			break
		}
	}
	if area == 0 && posts == 0 {
		t.Fatal("no area with posts to drill into")
	}
	s = selectArea(s, area)
	if s.SelectedArea != area {
		t.Fatalf("SelectedArea = %d, want %d", s.SelectedArea, area)
	}
	if len(s.PostRows) == 0 {
		t.Fatal("the drill-down is empty")
	}
	bare := 0
	for _, r := range s.PostRows {
		if r.Kap == "" {
			bare++
		}
		if r.Post == "" {
			t.Errorf("post row %q has no post number", r.Formal)
		}
		if len(r.Cells) != len(s.Cols) {
			t.Errorf("post row %q has %d cells for %d columns", r.Formal, len(r.Cells), len(s.Cols))
		}
	}
	if bare > 0 {
		t.Errorf("%d of %d posts carry no chapter", bare, len(s.PostRows))
	}
	t.Logf("area %d: %d posts, all with a chapter and a post number", area, len(s.PostRows))
}

// The bar is the magnitude, so its bucket must be zero only for a zero deviation and
// must never exceed five. A bucket that drifted would change what the picture says
// without changing any number on the page.
func TestPartyBudgetBucketsAreBoundedAndZeroOnlyAtZero(t *testing.T) {
	if got := bucketOf(0, 100); got != 0 {
		t.Errorf("bucketOf(0) = %d, want 0: a zero deviation has no bar to draw", got)
	}
	if got := bucketOf(-100, 100); got != 5 {
		t.Errorf("bucketOf(-100, 100) = %d, want 5: the row's largest deviation is the full bar", got)
	}
	if got := bucketOf(1, 1000); got != 1 {
		t.Errorf("bucketOf(1, 1000) = %d, want 1: the smallest non-zero deviation still shows", got)
	}
	prev := 0
	for _, v := range []int64{1, 100, 300, 700, 1000} {
		b := bucketOf(v, 1000)
		if b < prev {
			t.Errorf("bucket for %d is %d, below the smaller deviation's %d", v, b, prev)
		}
		if b < 1 || b > 5 {
			t.Errorf("bucket for %d is %d, outside 1..5", v, b)
		}
		prev = b
	}
	// Sign must not change the size of the bar: -700 and +700 are equally large.
	if a, b := bucketOf(700, 1000), bucketOf(-700, 1000); a != b {
		t.Errorf("bucketOf(700) = %d and bucketOf(-700) = %d: the bar shows magnitude, not sign", a, b)
	}
}
