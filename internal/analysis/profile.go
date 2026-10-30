package analysis

import (
	"math"
	"strconv"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// DimensionRow is one party's record in one dimension, with the time span the
// record actually covers.
//
// The year columns are the honest part. A "position over time" line needs two
// dated observations in the same cell, and almost no cell has that: 496 of 518
// coded cells rest on a single item (measured 2026-10-28; it was 510 of 518
// before the utsagn-55 batch deepened six cells, so re-measure, do not quote). So this does not draw a trend. It reports
// how far back the party's evidence in the dimension reaches and how many cells
// could support a trend at all, which is the difference between "we know what
// they did in 2024" and "we know what they did, and when".
type DimensionRow struct {
	Dimension   string
	Statements  int // statements filed under the dimension
	Coded       int // statements where this party has a coded position
	Mean        float64
	OldestYear  int
	NewestYear  int
	SingleDate  int // coded statements whose evidence carries one distinct date
	MultiDate   int // coded statements whose evidence carries more than one
	WithoutYear int // coded statements with no parseable date at all
}

// YearSpan is how far back the evidence behind a row reaches.
func (r DimensionRow) HasSpan() bool { return r.OldestYear > 0 && r.NewestYear > 0 }

// Span is the number of years between the oldest and newest evidence.
func (r DimensionRow) Span() int {
	if !r.HasSpan() {
		return 0
	}
	return r.NewestYear - r.OldestYear
}

// MeanLabel renders the mean on the same -2..+2 scale the rest of the site uses.
// ok is false when nothing was coded, so a caller never prints "0.0" for
// "nothing known".
//
// The separator is a comma and the sign a typographic minus, because this string
// goes straight into a Norwegian page. strconv's dot form shipped 108 cells reading
// "-0.2" on /kart while every other figure on the site read "-0,2", which is the
// kind of inconsistency a reader reads as two different scales.
func (r DimensionRow) MeanLabel() (string, bool) {
	if r.Coded == 0 {
		return "", false
	}
	neg := r.Mean < 0
	v := math.Abs(r.Mean)
	// Round, not truncate, at one decimal: a mean of 1.96 is 2.0, and printing 1.9
	// would understate it on a scale whose ends are exactly what a reader looks at.
	hundredths := int64(math.Round(v * 10))
	whole := hundredths / 10
	tenths := hundredths % 10
	out := strconv.FormatInt(whole, 10) + "," + strconv.FormatInt(tenths, 10)
	if neg {
		return "\u2212" + out, true
	}
	return out, true
}

// yearOf pulls the year out of an evidence date. Evidence dates are ISO in
// practice, but a few free-form ones exist in the corpus ("2025 (partiprogram)"),
// so the parse is deliberately partial and a failure is counted, not hidden.
func yearOf(date string) (int, bool) {
	date = strings.TrimSpace(date)
	if len(date) < 4 {
		return 0, false
	}
	y, err := strconv.Atoi(date[:4])
	if err != nil || y < 1900 || y > 2100 {
		return 0, false
	}
	return y, true
}

// DimensionProfile is one party's row per dimension, ordered the way the
// statements are filed.
func DimensionProfile(ds *data.Dataset, party string) []DimensionRow {
	// dimension -> statement ids, in first-seen order.
	var order []string
	stmtsByDim := map[string][]string{}
	for _, s := range ds.Statements {
		if _, seen := stmtsByDim[s.Dimension]; !seen {
			order = append(order, s.Dimension)
		}
		stmtsByDim[s.Dimension] = append(stmtsByDim[s.Dimension], s.ID)
	}
	dimOf := map[string]string{}
	for _, s := range ds.Statements {
		dimOf[s.ID] = s.Dimension
	}

	// cell -> coded? positions, and the distinct dates behind it.
	type cell struct {
		coded bool
		sum   int
		dates map[string]bool
	}
	cells := map[string]*cell{}
	for _, e := range ds.Evidence {
		if e.PartyID != party || e.StatementID == "" {
			continue
		}
		c := cells[e.StatementID]
		if c == nil {
			c = &cell{dates: map[string]bool{}}
			cells[e.StatementID] = c
		}
		if e.Date != "" {
			c.dates[e.Date] = true
		}
		if e.Position != nil {
			c.coded = true
			c.sum += *e.Position
		}
	}

	rows := make([]DimensionRow, 0, len(order))
	for _, dim := range order {
		row := DimensionRow{Dimension: dim, Statements: len(stmtsByDim[dim])}
		total := 0
		for _, sid := range stmtsByDim[dim] {
			c := cells[sid]
			if c == nil || !c.coded {
				continue
			}
			row.Coded++
			total += c.sum
			// The span is built from dates that actually parse. A cell whose only
			// date is free-form ("2025 (partiprogram)" parses, "ukjent" does not)
			// cannot support a trend, so it counts as undated rather than as one
			// observation: counting the string would claim a precision the source
			// does not have.
			parsed := 0
			for d := range c.dates {
				y, ok := yearOf(d)
				if !ok {
					continue
				}
				parsed++
				if row.OldestYear == 0 || y < row.OldestYear {
					row.OldestYear = y
				}
				if y > row.NewestYear {
					row.NewestYear = y
				}
			}
			switch {
			case parsed == 0:
				row.WithoutYear++
			case parsed == 1:
				row.SingleDate++
			default:
				row.MultiDate++
			}
		}
		if row.Coded > 0 {
			row.Mean = float64(total) / float64(row.Coded)
		}
		rows = append(rows, row)
	}
	return rows
}

// PartyTotals sums a profile: coded statements, and how many could support a
// trend because they carry two dated observations.
func PartyTotals(rows []DimensionRow) (coded, multiDate, singleDate int) {
	for _, r := range rows {
		coded += r.Coded
		multiDate += r.MultiDate
		singleDate += r.SingleDate
	}
	return coded, multiDate, singleDate
}

// YearsCovered returns the first and last year any dimension's evidence reaches.
func YearsCovered(rows []DimensionRow) (int, int) {
	oldest, newest := 0, 0
	for _, r := range rows {
		if !r.HasSpan() {
			continue
		}
		if oldest == 0 || r.OldestYear < oldest {
			oldest = r.OldestYear
		}
		if r.NewestYear > newest {
			newest = r.NewestYear
		}
	}
	return oldest, newest
}

// PartyIDs returns the dataset's party ids in the order the site uses. The
// order is the data's own, not alphabetical: the site's convention is the
// parties.yaml order and a view that re-sorts would disagree with every other
// page for no gain.
func PartyIDs(ds *data.Dataset) []string {
	out := make([]string, 0, len(ds.Parties))
	for _, p := range ds.Parties {
		out = append(out, p.ID)
	}
	return out
}

// MultiDateStatements returns the statement ids where one party has more than one
// distinct evidence date. It is the per-cell version of DimensionRow.MultiDate, and
// it exists so a view can mark the cells that carry a dated second observation
// instead of only reporting a count per party.
//
// Distinct dates, not rows: two evidence items from the same day are one
// observation, and counting them as two would manufacture a trend where there is a
// duplicate.
func MultiDateStatements(ds *data.Dataset, party string) map[string]bool {
	dates := map[string]map[string]bool{}
	for _, e := range ds.Evidence {
		if e.PartyID != party || e.Position == nil {
			continue
		}
		d := strings.TrimSpace(e.Date)
		if d == "" {
			continue
		}
		if dates[e.StatementID] == nil {
			dates[e.StatementID] = map[string]bool{}
		}
		dates[e.StatementID][d] = true
	}
	out := map[string]bool{}
	for id, set := range dates {
		if len(set) > 1 {
			out[id] = true
		}
	}
	return out
}
