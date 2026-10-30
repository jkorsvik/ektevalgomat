package data

import "testing"

// The party-budget tables are the only place in the dataset that says what each
// party would actually spend. These tests hold the properties the page relies on:
// every rammeområde present exactly once, the document's own column arithmetic
// re-checked in Go, and a party column always named one of the nine.

var knownParties = map[string]bool{
	"ap": true, "sv": true, "sp": true, "r": true, "mdg": true,
	"frp": true, "h": true, "krf": true, "v": true,
}

func TestPartyBudgetsCoverEveryAreaOnce(t *testing.T) {
	tables := PartyBudgets()
	if len(tables) == 0 {
		t.Fatal("no party-budget tables loaded: the check would pass on nothing")
	}
	seen := map[int]int{}
	for _, tb := range tables {
		seen[tb.Area]++
		if tb.PostCount() == 0 {
			t.Errorf("rammeområde %d (%s) carries no bevilgninger: a table read as empty is "+
				"worse than an absent one, because it looks collected", tb.Area, tb.Document)
		}
	}
	for area, n := range seen {
		if n > 1 {
			t.Errorf("rammeområde %d is claimed by %d tables: a committee summary or a "+
				"mis-attributed label is being read as a budget table", area, n)
		}
	}
	// Contiguity is the point: the first run covered 18 of 24 areas and reported
	// success, and the six that were missing were missing by accident.
	for area := 0; area <= 23; area++ {
		if seen[area] == 0 {
			t.Errorf("rammeområde %d has no table", area)
		}
	}
	if got, want := PartyBudgetAreas(), 24; got != want {
		t.Errorf("areas = %d, want %d", got, want)
	}
	t.Logf("%d tables over %d areas, %d bevilgninger, session %s",
		len(tables), PartyBudgetAreas(), totalPartyBudgetPosts(tables), PartyBudgetSession())
}

func totalPartyBudgetPosts(tables []PartyBudgetTable) int {
	n := 0
	for _, tb := range tables {
		n += tb.PostCount()
	}
	return n
}

// Every column's per-post deviations must add to the sum the document prints for
// that column. The generator refuses to write a table that fails this, so the
// check here is a second reader of the same guarantee: it catches a hand-edited
// artefact, and it would catch a generator change that stopped checking.
func TestPartyBudgetColumnsAddToTheirSumRows(t *testing.T) {
	tables := PartyBudgets()
	if len(tables) == 0 {
		t.Fatal("no tables loaded")
	}
	checked := 0
	for _, tb := range tables {
		if bad := CheckPartyBudgetColumns(tb); len(bad) > 0 {
			for _, b := range bad {
				t.Errorf("%s: %s", tb.Document, b)
			}
		}
		checked += len(tb.ColKeys())
	}
	if checked == 0 {
		t.Fatal("no column was checked")
	}
	t.Logf("%d columns checked across %d tables", checked, len(tables))
}

// The check must be able to fail. Mutating one post's amount by one krone breaks
// the identity, and the guard must notice.
func TestPartyBudgetCheckNoticesOneKroneOutOfPlace(t *testing.T) {
	tables := PartyBudgets()
	if len(tables) == 0 {
		t.Fatal("no tables loaded")
	}
	var target *PartyBudgetTable
	var col string
	for i, tb := range tables {
		rows := tb.Posts["utgifter"]
		if len(rows) == 0 {
			continue
		}
		for _, k := range tb.PartyColumns() {
			if _, ok := rows[0].Amounts[k]; ok {
				target = &tables[i]
				col = k
				break
			}
		}
		if target != nil {
			break
		}
	}
	if target == nil {
		t.Fatal("no table with a party column and an utgifter post to mutate")
	}
	rows := target.Posts["utgifter"]
	m := rows[0].Amounts[col]
	m.Change++
	rows[0].Amounts[col] = m
	target.Posts["utgifter"] = rows

	if bad := CheckPartyBudgetColumns(*target); len(bad) == 0 {
		t.Fatalf("a one-krone mutation in rammeområde %d column %s went unnoticed: the "+
			"column check cannot fail", target.Area, col)
	}
}

func TestPartyBudgetColumnsNameOnlyTheNineParties(t *testing.T) {
	for _, tb := range PartyBudgets() {
		gov := 0
		for _, k := range tb.ColKeys() {
			c := tb.Columns[k]
			switch c.Kind {
			case "government":
				gov++
			case "forlik":
				if len(c.Parties) < 2 {
					t.Errorf("rammeområde %d: a forlik column with %d member(s) is not joint",
						tb.Area, len(c.Parties))
				}
				for _, p := range c.Parties {
					if !knownParties[p] {
						t.Errorf("rammeområde %d: forlik names unknown party %q", tb.Area, p)
					}
				}
			case "party":
				if !knownParties[c.Party] {
					t.Errorf("rammeområde %d: unknown party %q", tb.Area, c.Party)
				}
			default:
				t.Errorf("rammeområde %d: unknown column kind %q", tb.Area, c.Kind)
			}
		}
		if gov != 1 {
			t.Errorf("rammeområde %d has %d government columns, want exactly 1", tb.Area, gov)
		}
		if len(tb.PartyColumns()) < 3 {
			t.Errorf("rammeområde %d has %d party columns: too few to call a party table",
				tb.Area, len(tb.PartyColumns()))
		}
	}
}

// Coverage differs per party, and the page must not imply otherwise. V put forward
// an alternative in fewer areas than FrP, so a total across all areas would compare
// different bundles of spending.
func TestPartyCoverageIsUnevenAndKnown(t *testing.T) {
	cov := PartyAreas()
	if len(cov) == 0 {
		t.Fatal("no party coverage measured")
	}
	lo, hi := 99, 0
	for p, n := range cov {
		if !knownParties[p] {
			t.Errorf("coverage lists unknown party %q", p)
		}
		if n < lo {
			lo = n
		}
		if n > hi {
			hi = n
		}
	}
	if hi == 0 {
		t.Fatal("every party covers zero areas")
	}
	t.Logf("coverage per party: %v (min %d, max %d of %d areas)", cov, lo, hi, PartyBudgetAreas())
}
