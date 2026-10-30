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
			if !c.Has {
				missing++
				if c.Change != "" {
					t.Errorf("rammeområde %d column %s is missing but carries %q",
						r.Area, s.Cols[i].Party, c.Change)
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
			if !got.Has {
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
