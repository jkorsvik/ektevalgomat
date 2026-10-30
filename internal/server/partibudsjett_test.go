package server

import (
	"fmt"
	"math"
	"strconv"
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

// The comparable total is the only cross-area figure on the page, so it must never
// reach beyond the areas where every party has a column. If it counted an area one
// party is missing from, it would compare coverage and call it money.
func TestPartyBudgetCommonSumCoversOnlySharedAreas(t *testing.T) {
	s := buildPartyBudgetSection(&data.Dataset{})
	common := s.Common.Areas
	if len(common) == 0 {
		t.Fatal("no common areas: the comparable sum has nothing to stand on")
	}
	// The binding constraint is the party with the fewest columns, so the shared set
	// can never be larger than the smallest coverage.
	smallest := s.Areas
	for _, c := range s.Coverage {
		if c.Areas < smallest {
			smallest = c.Areas
		}
	}
	if len(common) > smallest {
		t.Errorf("%d shared areas, but the thinnest party covers only %d", len(common), smallest)
	}

	// Every shared area must carry every party column.
	for _, area := range common {
		var row partyBudgetRow
		found := false
		for _, r := range s.Rows {
			if r.Area == area {
				row, found = r, true
				break
			}
		}
		if !found {
			t.Fatalf("shared area %d has no row", area)
		}
		for i, col := range s.Cols {
			if col.Joint {
				continue
			}
			if row.Cells[i].Kind == "absent" {
				t.Errorf("area %d is counted as shared, but %s has no column there", area, col.Label)
			}
		}
	}

	// And each total must equal the cells it claims to sum, read back from the text
	// the reader sees. This is deliberately a second, independent path: comparing
	// the total against rowChange() would use the same function that built it, and
	// that is exactly how a dropped sign stayed invisible. mrd() prints U+2212, so
	// the parser must normalise it; a parser expecting ASCII "-" reads every
	// negative as zero and the "sum" becomes a sum of increases only.
	parse := func(txt string) (int64, bool) {
		norm := strings.NewReplacer("\u2212", "-", "\u2013", "-", "\u2014", "-", "\u00a0", "").Replace(txt)
		norm = strings.TrimSpace(strings.ReplaceAll(norm, ",", "."))
		var f float64
		if _, err := fmt.Sscanf(norm, "%g", &f); err != nil {
			return 0, false
		}
		return int64(math.Round(f * 1e9)), true
	}
	for i, col := range s.Cols {
		if col.Joint {
			continue
		}
		var want, posOnly, biggestNegative int64
		negatives := 0
		for _, r := range s.Rows {
			if !containsInt(common, r.Area) {
				continue
			}
			if i >= len(r.Cells) || r.Cells[i].Kind != "fig" {
				continue
			}
			v, ok := parse(r.Cells[i].Change)
			if !ok {
				t.Fatalf("%s: the cell %q does not parse, so the total cannot be checked", col.Label, r.Cells[i].Change)
			}
			want += v
			if v > 0 {
				posOnly += v
			} else if v < 0 {
				negatives++
				if -v > biggestNegative {
					biggestNegative = -v
				}
			}
		}
		var got int64
		found := false
		for _, cc := range s.Common.Cols {
			if cc.Label == col.Label {
				v, ok := parse(cc.Total)
				if !ok {
					t.Fatalf("%s: the printed total %q does not parse", col.Label, cc.Total)
				}
				got, found = v, true
			}
		}
		if !found {
			t.Fatalf("%s has no printed total", col.Label)
		}
		// Each cell truncates to one decimal, so a text-based sum can drift by up to
		// 0.1 mrd per cell against the exact total.
		tolerance := int64(len(common)+2) * 100_000_000
		if diff := got - want; diff > tolerance || diff < -tolerance {
			t.Errorf("%s: total %s, but its shared cells sum to %s", col.Label, mrd(got), mrd(want))
		}
		// The direct guard against the dropped sign: if a negative cell exists, the
		// total must sit at least that far below the sum of the increases.
		if negatives > 0 {
			gap := posOnly - got
			if gap < biggestNegative/2 {
				t.Errorf("%s: has %d negative cells (largest %s) but its total %s sits only %s below the sum of increases %s: negatives were left out",
					col.Label, negatives, mrd(biggestNegative), mrd(got), mrd(gap), mrd(posOnly))
			}
		}
	}
	var as []string
	for _, a := range common {
		as = append(as, strconv.Itoa(a))
	}
	t.Logf("%d shared areas of %d: %s", len(common), s.Areas, strings.Join(as, ", "))
}

// The ordering must be a real ordering: sorting by spread has to put the largest
// disagreement first, and the default must leave the administrative order alone.
func TestPartyBudgetSortPutsTheLargestSpreadFirst(t *testing.T) {
	s := buildPartyBudgetSection(&data.Dataset{})
	first := s.Rows[0].Area

	bySpread := sortRows(s, "spredning")
	if bySpread.Sort != "spredning" {
		t.Errorf("Sort = %q, want spredning", bySpread.Sort)
	}
	for i := 1; i < len(bySpread.Rows); i++ {
		if bySpread.Rows[i-1].SpreadMrd < bySpread.Rows[i].SpreadMrd {
			t.Fatalf("row %d spreads %.1f after row %d's %.1f",
				bySpread.Rows[i-1].Area, bySpread.Rows[i-1].SpreadMrd,
				bySpread.Rows[i].Area, bySpread.Rows[i].SpreadMrd)
		}
	}
	if bySpread.Rows[0].SpreadMrd <= 0 {
		t.Error("the largest spread is not positive")
	}

	bySize := sortRows(s, "storrelse")
	for i := 1; i < len(bySize.Rows); i++ {
		if bySize.Rows[i-1].PeakMrd < bySize.Rows[i].PeakMrd {
			t.Fatalf("storrelse is out of order at %d", i)
		}
	}

	plain := sortRows(s, "")
	if plain.Rows[0].Area != first {
		t.Errorf("the default order moved area %d off the top", first)
	}
	for i, r := range plain.Rows {
		if r.Area != i {
			t.Fatalf("the default order is not the administrative one: row %d is area %d", i, r.Area)
		}
	}
	t.Logf("størst uenighet: område %d (%.1f), størst beløp: område %d (%.1f)",
		bySpread.Rows[0].Area, bySpread.Rows[0].SpreadMrd, bySize.Rows[0].Area, bySize.Rows[0].PeakMrd)
}
