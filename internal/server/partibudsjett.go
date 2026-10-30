package server

import (
	"strconv"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The party-budget section of /budsjett: what each party would actually spend,
// rammeområde by rammeområde, read out of the committee innstillinger.
//
// Two things this page must never do. It must not print a total across areas,
// because coverage differs per party (V proposed an alternative in 11 of 24
// areas, FrP in 24) so a total would compare different bundles of spending. And
// it must not read a missing column as zero: a party with no column in an area
// proposed nothing there, and a zero would be a claim it made.

// partyOrder is the canonical reading order of the nine parties, so the columns
// are the same on every row and on every rebuild.
var partyOrder = []string{"ap", "sv", "sp", "r", "mdg", "frp", "h", "krf", "v"}

type partyBudgetSection struct {
	Session  string
	Cols     []partyBudgetCol
	Rows     []partyBudgetRow
	Coverage []partyBudgetCoverage
	Posts    int
	Areas    int
}

type partyBudgetCol struct {
	Party string // party id, or "forlik"
	Label string // what the column header prints
	Joint bool
}

type partyBudgetRow struct {
	Area     int
	Document string
	URL      string
	Posts    int
	Cells    []partyBudgetCell // aligned with Cols, index for index
}

type partyBudgetCell struct {
	Has    bool
	Change string // signed deviation, mrd kroner, comma decimal
	Class  string // "up", "down" or "flat"
}

type partyBudgetCoverage struct {
	Party string
	Label string
	Areas int
	Of    int
}

// mrd formats a kroner amount as signed billions with one decimal: "+33,2".
// One decimal is the honest precision for a figure read off a table printed in
// tusen kroner at this magnitude; more digits would imply a precision the
// grouping does not carry.
func mrd(kr int64) string {
	sign := "+"
	if kr < 0 {
		sign = "−"
		kr = -kr
	}
	whole := kr / 1_000_000_000
	tenths := (kr % 1_000_000_000) / 100_000_000
	return sign + strconv.FormatInt(whole, 10) + "," + strconv.FormatInt(tenths, 10)
}

func mrdClass(kr int64) string {
	switch {
	case kr > 0:
		return "up"
	case kr < 0:
		return "down"
	default:
		return "flat"
	}
}

// buildPartyBudgetSection turns the loaded tables into the page's shape.
func buildPartyBudgetSection(ds *data.Dataset) partyBudgetSection {
	tables := data.PartyBudgets()
	s := partyBudgetSection{Session: data.PartyBudgetSession(), Areas: len(tables)}

	// The columns are the union of what any table names, in canonical order, with
	// the joint forlik column first because it is not a party.
	present := map[string]bool{}
	joint := false
	for _, tb := range tables {
		for _, k := range tb.PartyColumns() {
			present[tb.Columns[k].Party] = true
		}
		if tb.ForlikColumn() != "" {
			joint = true
		}
	}
	if joint {
		s.Cols = append(s.Cols, partyBudgetCol{Party: "forlik", Label: "Forlik", Joint: true})
	}
	for _, p := range partyOrder {
		if present[p] {
			label := p
			if pl, ok := ds.Party(p); ok && pl.Short != "" {
				label = pl.Short
			}
			s.Cols = append(s.Cols, partyBudgetCol{Party: p, Label: label})
		}
	}

	for _, tb := range tables {
		netto := tb.Netto()
		row := partyBudgetRow{
			Area: tb.Area, Document: tb.Document, URL: tb.URL, Posts: tb.PostCount(),
		}
		s.Posts += row.Posts
		for _, col := range s.Cols {
			cell := partyBudgetCell{}
			var key string
			if col.Joint {
				key = tb.ForlikColumn()
			} else {
				for _, k := range tb.PartyColumns() {
					if tb.Columns[k].Party == col.Party {
						key = k
						break
					}
				}
			}
			if key != "" {
				if m, ok := netto[key]; ok {
					cell.Has = true
					cell.Change = mrd(m.Change)
					cell.Class = mrdClass(m.Change)
				}
			}
			row.Cells = append(row.Cells, cell)
		}
		s.Rows = append(s.Rows, row)
	}

	cov := data.PartyAreas()
	for _, p := range partyOrder {
		if n, ok := cov[p]; ok {
			label := p
			if pl, ok := ds.Party(p); ok && pl.Short != "" {
				label = pl.Short
			}
			s.Coverage = append(s.Coverage, partyBudgetCoverage{Party: p, Label: label, Areas: n, Of: s.Areas})
		}
	}
	return s
}

// partyBudgetLede is the sentence the section opens with. It names the deviation
// basis and the coverage limit, because both decide whether two cells may be read
// against each other.
func partyBudgetLede(s partyBudgetSection) string {
	var b strings.Builder
	b.WriteString("Hver kolonne er partiets eget alternative budsjett for rammeområdet, ")
	b.WriteString("vist som avvik fra regjeringens forslag i Prop. 1 S. Forliket er flere ")
	b.WriteString("partier sammen og står først, skilt fra partikolonnene.")
	b.WriteString(" Tomt felt betyr at partiet ikke la fram noe alternativ i dette ")
	b.WriteString("rammeområdet, ikke at det foreslo null. Dekningen er ulik: ")
	for i, c := range s.Coverage {
		if i > 0 {
			if i == len(s.Coverage)-1 {
				b.WriteString(" og ")
			} else {
				b.WriteString(", ")
			}
		}
		b.WriteString(c.Label + " har " + strconv.Itoa(c.Areas) + " av " + strconv.Itoa(c.Of))
	}
	b.WriteString(". Derfor summeres det ikke på tvers av rammeområder.")
	return b.String()
}
