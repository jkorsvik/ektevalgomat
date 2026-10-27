package server

import (
	"strconv"

	"github.com/jkorsvik/ektevalgomat/internal/analysis"
	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The map view: one table of dimension by party, and the time span behind each
// party's record. Everything is derived from the dataset at request time.
//
// The cell is deliberately "mean (coded/statements)" rather than a bare number.
// A mean over 1 of 7 statements is not the same fact as a mean over 7 of 7, and
// a page that shows only the mean invites a reader to compare two numbers that
// are not comparable.
type kartCell struct {
	Party      string
	Short      string
	Mean       string
	Coded      int
	Statements int
	HasData    bool
	MultiDate  int
	// Dated marks a cell that holds at least one statement with two distinct
	// evidence dates. It is the per-cell answer to "can this cell say anything
	// about a year", and it is derived, never stored.
	Dated bool
}

type kartRow struct {
	Dimension string
	Cells     []kartCell
}

type kartParty struct {
	ID         string
	Short      string
	Name       string
	Coded      int
	MultiDate  int
	SingleDate int
	Oldest     int
	Newest     int
}

type kartView struct {
	Meta         pageMeta
	SnapshotDate string
	Dimensions   []string
	Parties      []kartParty
	Rows         []kartRow
	TotalCoded   int
	TotalMulti   int
}

func buildKartView(ds *data.Dataset, snapshotDate string) kartView {
	ids := analysis.PartyIDs(ds)
	byID := make(map[string]data.Party, len(ds.Parties))
	for _, p := range ds.Parties {
		byID[p.ID] = p
	}

	profiles := make(map[string][]analysis.DimensionRow, len(ids))
	for _, id := range ids {
		profiles[id] = analysis.DimensionProfile(ds, id)
	}

	// Dimension order comes from the data, so a new dimension appears without
	// anyone editing this file.
	var dims []string
	seen := map[string]bool{}
	for _, s := range ds.Statements {
		if !seen[s.Dimension] {
			seen[s.Dimension] = true
			dims = append(dims, s.Dimension)
		}
	}

	view := kartView{
		Meta: meta("Kartet · hva hvert parti har gjort, per sakområde",
			"Hvert partis kodede posisjon per sakområde, med antall utsagn bak hvert tall og hvor langt tilbake kildene rekker.",
			"/kart"),
		SnapshotDate: snapshotDate,
		Dimensions:   dims,
	}

	for _, id := range ids {
		p := byID[id]
		rows := profiles[id]
		coded, multi, single := analysis.PartyTotals(rows)
		oldest, newest := analysis.YearsCovered(rows)
		view.Parties = append(view.Parties, kartParty{
			ID: id, Short: p.Short, Name: p.Name,
			Coded: coded, MultiDate: multi, SingleDate: single,
			Oldest: oldest, Newest: newest,
		})
		view.TotalCoded += coded
		view.TotalMulti += multi
	}

	// Per-party, per-statement multi-date facts, and the statements in each
	// dimension, so a cell can be marked rather than only counted.
	multiByParty := map[string]map[string]bool{}
	for _, id := range ids {
		multiByParty[id] = analysis.MultiDateStatements(ds, id)
	}
	stmtsByDim := map[string][]string{}
	for _, st := range ds.Statements {
		stmtsByDim[st.Dimension] = append(stmtsByDim[st.Dimension], st.ID)
	}

	for _, dim := range dims {
		row := kartRow{Dimension: dim}
		for _, id := range ids {
			cell := kartCell{Party: id, Short: byID[id].Short}
			for _, r := range profiles[id] {
				if r.Dimension != dim {
					continue
				}
				cell.Coded, cell.Statements, cell.MultiDate = r.Coded, r.Statements, r.MultiDate
				for _, stmtID := range stmtsByDim[dim] {
					if multiByParty[id][stmtID] {
						cell.Dated = true
					}
				}
				if label, ok := r.MeanLabel(); ok {
					cell.Mean, cell.HasData = label, true
				}
				break
			}
			row.Cells = append(row.Cells, cell)
		}
		view.Rows = append(view.Rows, row)
	}
	return view
}

// pctOf renders a share as a whole percentage, for the "how much of your record
// is a single observation" line. It is here rather than in the template because
// the template must not do arithmetic on data.
func pctOf(part, whole int) string {
	if whole == 0 {
		return "–"
	}
	return strconv.Itoa(part*100/whole) + " %"
}
