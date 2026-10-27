package analysis

import (
	"sort"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The year grid is the party x dimension x year view: for one party, in one
// dimension, in one year, what did the record say.
//
// The honest constraint is that most cells were observed once. So this does not
// interpolate, carry a value forward, or hold a position constant between two
// observations. A year with no observation is empty, and the empty cell is the
// finding: nobody has looked here yet.

// YearCell is what the record holds for one party in one dimension in one year.
type YearCell struct {
	Positions int // distinct statements with a coded position in this year
	Sum       int // sum of those positions, on the -2..+2 scale
	Min       int // lowest coded position in the cell
	Max       int // highest coded position in the cell
}

// Disagrees reports whether the cell holds positions on both sides of zero.
//
// This is the one thing a mean cannot say. Two positions, +2 and -2, average to
// 0, and 0 on this scale means "neutral / split", which is false about both: the
// party took a clear side twice, in opposite directions. A caller must show the
// spread instead of the mean, or say that the cell is split.
func (c YearCell) Disagrees() bool { return c.Min < 0 && c.Max > 0 }

// Mean is the average position in the cell, or 0 when nothing was coded.
func (c YearCell) Mean() float64 {
	if c.Positions == 0 {
		return 0
	}
	return float64(c.Sum) / float64(c.Positions)
}

// PartyYears is one party's record in one dimension, sliced by year.
type PartyYears struct {
	Party   string
	Years   map[int]YearCell
	Undated int // coded statements whose evidence carries no parseable year
}

// Total is the number of coded statements behind this row.
func (p PartyYears) Total() int {
	n := p.Undated
	for _, c := range p.Years {
		n += c.Positions
	}
	return n
}

// DimensionYears is one dimension, all parties, sliced by year.
type DimensionYears struct {
	Dimension  string
	Statements int
	Parties    []PartyYears
}

// YearGrid slices the dataset by party, dimension and year.
//
// A statement observed in two different years counts once in each, because the
// observation is real in both: "we know they held this in 2021" and "we know
// they held this in 2024" are two facts. Two rows from the same year count once
// in that year: a duplicate date is one observation, not two, and counting it
// twice would inflate a cell and shift its mean.
//
// years is every year that holds at least one observation anywhere, ascending,
// so a caller renders a fixed column set instead of guessing a range.
func YearGrid(ds *data.Dataset, parties []string) (rows []DimensionYears, years []int) {
	// dimension -> statement ids, in first-seen order, and each statement's dimension.
	var dimOrder []string
	stmtsByDim := map[string][]string{}
	dimOf := map[string]string{}
	for _, s := range ds.Statements {
		if _, seen := stmtsByDim[s.Dimension]; !seen {
			dimOrder = append(dimOrder, s.Dimension)
		}
		stmtsByDim[s.Dimension] = append(stmtsByDim[s.Dimension], s.ID)
		dimOf[s.ID] = s.Dimension
	}

	// (dimension, party) -> year -> cell, plus the undated count. Held as sets of
	// statement ids so a repeated row in the same year cannot count twice.
	type key struct{ dim, party string }
	cells := map[key]map[int]map[string]int{} // year -> statement id set
	sums := map[key]map[int]int{}
	mins := map[key]map[int]int{}
	maxs := map[key]map[int]int{}
	undated := map[key]map[string]bool{}
	yearSet := map[int]bool{}

	wanted := map[string]bool{}
	for _, p := range parties {
		wanted[p] = true
	}

	for _, e := range ds.Evidence {
		if e.Position == nil {
			continue // a source without a coded position is not an observation
		}
		dim, ok := dimOf[e.StatementID]
		if !ok || !wanted[e.PartyID] {
			continue
		}
		k := key{dim, e.PartyID}
		y, ok := yearOf(e.Date)
		if !ok {
			if undated[k] == nil {
				undated[k] = map[string]bool{}
			}
			undated[k][e.StatementID] = true
			continue
		}
		if cells[k] == nil {
			cells[k] = map[int]map[string]int{}
			sums[k] = map[int]int{}
		}
		if cells[k][y] == nil {
			cells[k][y] = map[string]int{}
		}
		if _, seen := cells[k][y][e.StatementID]; seen {
			continue // same statement, same year: one observation
		}
		cells[k][y][e.StatementID] = 1
		sums[k][y] += *e.Position
		if mins[k] == nil {
			mins[k] = map[int]int{}
			maxs[k] = map[int]int{}
		}
		if v, ok := mins[k][y]; !ok || *e.Position < v {
			mins[k][y] = *e.Position
		}
		if v, ok := maxs[k][y]; !ok || *e.Position > v {
			maxs[k][y] = *e.Position
		}
		yearSet[y] = true
	}

	for y := range yearSet {
		years = append(years, y)
	}
	sort.Ints(years)

	for _, dim := range dimOrder {
		row := DimensionYears{Dimension: dim, Statements: len(stmtsByDim[dim])}
		for _, p := range parties {
			k := key{dim, p}
			py := PartyYears{Party: p, Years: map[int]YearCell{}, Undated: len(undated[k])}
			for y, set := range cells[k] {
				py.Years[y] = YearCell{
					Positions: len(set),
					Sum:       sums[k][y],
					Min:       mins[k][y],
					Max:       maxs[k][y],
				}
			}
			if py.Total() == 0 {
				continue // nothing known here: an absent row beats an empty one
			}
			row.Parties = append(row.Parties, py)
		}
		if len(row.Parties) > 0 {
			rows = append(rows, row)
		}
	}
	return rows, years
}
