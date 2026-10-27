package server

import (
	"strconv"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/analysis"
	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The year view: party x dimension x year, computed at request time.
//
// Why a separate page from /kart: the map answers "what does this party stand for
// in this area", which is one number per cell. It cannot answer "when", because a
// mean has no date. This page keeps the dates, so every number it shows belongs to
// a named year, and an empty cell is the honest statement that no source has been
// collected for that year yet.
//
// The columns are the same in every table on purpose. A per-dimension column set
// would be narrower, but it would hide the thing a reader most needs to see: that
// the record is thin in most years, in every area at once.

type yearCellView struct {
	Label string // "1,5 (2)", or empty when there is no observation
	Has   bool
	N     int
	Tone  string // pos | neg | zero, for styling only
}

type yearPartyRow struct {
	ID      string
	Short   string
	Name    string
	Cells   []yearCellView
	Undated int
	Total   int
}

type yearDimBlock struct {
	Dimension  string
	Statements int
	Rows       []yearPartyRow
}

type yearView struct {
	Meta         pageMeta
	SnapshotDate string
	Years        []int
	Blocks       []yearDimBlock
	TotalCoded   int
	Dated        int
	Undated      int
	Oldest       int
	Newest       int
	Dimensions   int
	Parties      int
}

// yearMeanLabel renders a cell the way the rest of the site does: the mean on the
// -2..+2 scale with one decimal and a Norwegian comma, then the number of
// positions it was computed from. A cell with no observation returns "", never
// "0,0", because "nothing known" and "neutral" must not look alike.
func yearMeanLabel(c analysis.YearCell) string {
	if c.Positions == 0 {
		return ""
	}
	n := " (" + strconv.Itoa(c.Positions) + ")"
	// A cell holding positions on both sides of zero has no mean worth printing:
	// +2 and -2 average to 0, which on this scale reads as "neutral", and that is
	// false about both. The spread is the honest summary.
	if c.Disagrees() {
		return signed(c.Max) + " til " + signed(c.Min) + n
	}
	s := strconv.FormatFloat(c.Mean(), 'f', 1, 64)
	return strings.Replace(s, ".", ",", 1) + n
}

// signed renders a position the way the rest of the site does: an explicit minus
// for a negative, and no plus, so the two ends of a range are never confused.
func signed(v int) string {
	if v < 0 {
		return "−" + strconv.Itoa(-v)
	}
	return strconv.Itoa(v)
}

func cellTone(c analysis.YearCell) string {
	switch {
	case c.Positions == 0:
		return ""
	case c.Disagrees():
		return "split" // neither pos nor neg: the cell took both sides
	case c.Sum > 0:
		return "pos"
	case c.Sum < 0:
		return "neg"
	default:
		return "zero"
	}
}

func buildYearView(snapshotDate string) yearView {
	ds := data.Store
	ids := analysis.PartyIDs(ds)

	byID := map[string]data.Party{}
	for _, p := range ds.Parties {
		byID[p.ID] = p
	}

	rows, years := analysis.YearGrid(ds, ids)

	view := yearView{
		Meta: meta("Parti, sakområde og år · hva de gjorde, og når",
			"Hvert partis kodede posisjon per sakområde og år, med antall utsagn bak hvert tall. Tomme felt betyr at ingen kilde er samlet for det året.",
			"/tidslinje"),
		SnapshotDate: snapshotDate,
		Years:        years,
		Blocks:       make([]yearDimBlock, 0, len(rows)),
		Dimensions:   len(rows),
		Parties:      len(ids),
	}

	for _, dim := range rows {
		block := yearDimBlock{Dimension: dim.Dimension, Statements: dim.Statements}
		for _, py := range dim.Parties {
			p := byID[py.Party]
			row := yearPartyRow{
				ID:      py.Party,
				Short:   p.Short,
				Name:    p.Name,
				Cells:   make([]yearCellView, 0, len(years)),
				Undated: py.Undated,
				Total:   py.Total(),
			}
			for _, y := range years {
				c := py.Years[y]
				label := yearMeanLabel(c)
				row.Cells = append(row.Cells, yearCellView{
					Label: label,
					Has:   label != "",
					N:     c.Positions,
					Tone:  cellTone(c),
				})
				view.TotalCoded += c.Positions
				view.Dated += c.Positions
				if c.Positions > 0 {
					if view.Oldest == 0 || y < view.Oldest {
						view.Oldest = y
					}
					if y > view.Newest {
						view.Newest = y
					}
				}
			}
			view.Undated += py.Undated
			block.Rows = append(block.Rows, row)
		}
		view.Blocks = append(view.Blocks, block)
	}
	view.TotalCoded += view.Undated
	return view
}
