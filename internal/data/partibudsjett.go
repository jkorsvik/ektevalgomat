package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

//go:embed budget/partier.json
var partyBudgetJSON []byte

// PartyBudgetColumn is one column of a committee's party-budget table: the
// government's own proposal, the budsjettforlik several parties stand behind
// jointly, or a single party's alternative budget.
//
// The column map is read from the document's header text, never from position,
// because the columns are not in a fixed order and the forlik's member list
// differs per document. A forlik is a group and is kept apart from the party
// columns, so a joint agreement is never reported as one party's money.
type PartyBudgetColumn struct {
	Kind    string   `json:"kind"` // "government", "forlik" or "party"
	Party   string   `json:"party,omitempty"`
	Parties []string `json:"parties,omitempty"`
}

// Money is an amount in kroner with that column's deviation from the government's
// proposal. The source prints these in tusen kroner; the generator multiplies by
// 1000, so every figure here is kroner and the unit is never guessed at the page.
type Money struct {
	Kr     int64 `json:"kr"`
	Change int64 `json:"change"`
}

// PartyBudgetPost is one bevilgning: a kapittel, a post and what each column
// would spend on it.
type PartyBudgetPost struct {
	Kap     string           `json:"kap"`
	Post    string           `json:"post"`
	Formal  string           `json:"formal"`
	Amounts map[string]Money `json:"amounts"`
}

// PartyBudgetTable is one rammeområde's party-budget table.
type PartyBudgetTable struct {
	Document    string                       `json:"document"`
	Number      int                          `json:"number"`
	URL         string                       `json:"url"`
	Area        int                          `json:"rammeomrade"`
	Unit        string                       `json:"unit"`
	Columns     map[string]PartyBudgetColumn `json:"columns"`
	Sums        map[string]map[string]Money  `json:"sums"`
	Posts       map[string][]PartyBudgetPost `json:"posts"`
	CheckFailed []string                     `json:"check_failures"`
}

type partyBudgetFile struct {
	Session string             `json:"session"`
	Tables  []PartyBudgetTable `json:"tables"`
}

var partyBudgetDoc = mustLoadPartyBudgets()

func mustLoadPartyBudgets() partyBudgetFile {
	var f partyBudgetFile
	if err := json.Unmarshal(partyBudgetJSON, &f); err != nil {
		// A malformed generated file is a build defect: the generator is in this
		// repo and its output is committed, so this can only happen if someone
		// edited the artefact by hand.
		panic(fmt.Sprintf("internal/data/budget/partier.json is not valid: %v", err))
	}
	return f
}

// PartyBudgets is every committee's party-budget table, ordered by rammeområde so
// the page reads like the budget itself.
func PartyBudgets() []PartyBudgetTable {
	out := make([]PartyBudgetTable, len(partyBudgetDoc.Tables))
	copy(out, partyBudgetDoc.Tables)
	sort.Slice(out, func(i, j int) bool { return out[i].Area < out[j].Area })
	return out
}

// PartyBudgetSession is the storting session the tables come from.
func PartyBudgetSession() string { return partyBudgetDoc.Session }

// ColKeys returns the table's column keys in their document order, so a caller
// never iterates a map and gets a random column order.
func (t PartyBudgetTable) ColKeys() []string {
	keys := make([]string, 0, len(t.Columns))
	for k := range t.Columns {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return numLess(keys[i], keys[j]) })
	return keys
}

func numLess(a, b string) bool {
	if len(a) != len(b) {
		return len(a) < len(b)
	}
	return a < b
}

// PartyColumns returns the single parties' column keys in document order. A party
// that proposed no alternative in this area has no column, and its absence is
// absence of a proposal, not a zero.
func (t PartyBudgetTable) PartyColumns() []string {
	var out []string
	for _, k := range t.ColKeys() {
		if t.Columns[k].Kind == "party" {
			out = append(out, k)
		}
	}
	return out
}

// ForlikColumn returns the joint column's key, or "" when the table has none.
func (t PartyBudgetTable) ForlikColumn() string {
	for _, k := range t.ColKeys() {
		if t.Columns[k].Kind == "forlik" {
			return k
		}
	}
	return ""
}

// Netto returns each column's net figure for the area: the sum row the document
// itself prints, falling back to utgifter when a table has no netto row.
func (t PartyBudgetTable) Netto() map[string]Money {
	if line, ok := t.Sums["netto"]; ok && len(line) > 0 {
		return line
	}
	return t.Sums["utgifter"]
}

// PostCount is how many bevilgninger the table carries.
func (t PartyBudgetTable) PostCount() int {
	n := 0
	for _, rows := range t.Posts {
		n += len(rows)
	}
	return n
}

// PartyAreas counts, per party, in how many rammeområder it put forward an
// alternative. Coverage differs per party, which is why no total across areas is
// published: a sum over different sets of areas compares different bundles.
func PartyAreas() map[string]int {
	out := map[string]int{}
	for _, t := range partyBudgetDoc.Tables {
		for _, k := range t.PartyColumns() {
			out[t.Columns[k].Party]++
		}
	}
	return out
}

// PartyBudgetAreas is how many rammeområder the tables cover.
func PartyBudgetAreas() int {
	seen := map[int]bool{}
	for _, t := range partyBudgetDoc.Tables {
		seen[t.Area] = true
	}
	return len(seen)
}

// CheckPartyBudgetColumns re-adds each column's per-post deviations and compares
// them with the sum row the document prints. The generator refuses to write a
// table that fails, so this is a second, independent reading of the same guarantee
// in a different language: if the two ever disagree, one of them is wrong and the
// build says so instead of the page silently showing a figure nobody can trace.
func CheckPartyBudgetColumns(t PartyBudgetTable) []string {
	var bad []string
	for _, kind := range []string{"utgifter", "inntekter"} {
		rows := t.Posts[kind]
		sums := t.Sums[kind]
		if len(rows) == 0 || len(sums) == 0 {
			continue
		}
		for _, k := range t.ColKeys() {
			want, ok := sums[k]
			if !ok {
				continue
			}
			var got int64
			for _, r := range rows {
				if m, ok := r.Amounts[k]; ok {
					got += m.Change
				}
			}
			if got != want.Change {
				name := t.Columns[k].Party
				if name == "" {
					name = t.Columns[k].Kind
				}
				bad = append(bad, fmt.Sprintf("%s rammeområde %d column %s (%s): rows add to %d, sum row says %d",
					kind, t.Area, k, name, got, want.Change))
			}
		}
	}
	return bad
}
