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
	// The drill-down: the area the reader asked for, with its bevilgninger.
	SelectedArea int
	SelectedDoc  string
	SelectedURL  string
	PostRows     []partyBudgetPostRow
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
	// Spread is how far apart the parties are in this area, in mrd kroner: the
	// largest deviation minus the smallest. It is a comparison inside one row, so
	// it does not suffer the coverage problem a total across areas would.
	Spread string
}

// partyBudgetPostRow is one bevilgning as the drill-down shows it.
type partyBudgetPostRow struct {
	Kap    string
	Post   string
	Formal string
	Cells  []partyBudgetCell
}

type partyBudgetCell struct {
	// Kind is "fig", "absent" or "blank", and the difference matters: absent means
	// the party had no column in this area at all, blank means it had one but the
	// document printed no total for it. Rendering both as the same dash let a
	// reader read "not on the committee" and "chose not to propose" as one thing.
	Kind   string
	Change string // signed deviation, mrd kroner, comma decimal
	Class  string // "up", "down" or "flat"
	// Bucket is 1..5, the share of this row's largest deviation, so magnitude shows
	// as a bar. The CSP forbids inline styles (style-src 'self'), so the width is a
	// class, never a computed attribute.
	Bucket int
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

// columnKey is the table's column key for one of the page's columns, or "" when the
// party has no column in this area.
func columnKey(tb data.PartyBudgetTable, col partyBudgetCol) string {
	if col.Joint {
		return tb.ForlikColumn()
	}
	for _, k := range tb.PartyColumns() {
		if tb.Columns[k].Party == col.Party {
			return k
		}
	}
	return ""
}

// bucketOf turns a deviation into 1..5 of the row's largest, so magnitude can be a
// class. Zero stays 0: a party that proposed the government's own figure has no bar
// to draw, and a bar of minimum width would read as a small proposal.
func bucketOf(v, peak int64) int {
	if v == 0 || peak <= 0 {
		return 0
	}
	if v < 0 {
		v = -v
	}
	b := int((v*5 + peak - 1) / peak)
	if b < 1 {
		b = 1
	}
	if b > 5 {
		b = 5
	}
	return b
}

// buildPartyBudgetSection turns the loaded tables into the page's shape.
func buildPartyBudgetSection(ds *data.Dataset) partyBudgetSection {
	tables := data.PartyBudgets()
	// -1, not 0: the zero value would highlight rammeområde 0 and open an empty
	// drill-down for a reader who asked for nothing.
	s := partyBudgetSection{Session: data.PartyBudgetSession(), Areas: len(tables), SelectedArea: -1}

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

		// The row's own scale, so a bar means something inside its row. Comparing
		// magnitudes across rows would compare a 60 billion area with a 2 billion one.
		var changes []int64
		for _, col := range s.Cols {
			key := columnKey(tb, col)
			if key == "" {
				continue
			}
			if m, ok := netto[key]; ok {
				changes = append(changes, m.Change)
			}
		}
		peak := int64(0)
		lo, hi := int64(0), int64(0)
		for i, v := range changes {
			if v < 0 {
				v = -v
			}
			if v > peak {
				peak = v
			}
			if i == 0 || changes[i] < lo {
				lo = changes[i]
			}
			if i == 0 || changes[i] > hi {
				hi = changes[i]
			}
		}
		if len(changes) > 0 {
			row.Spread = mrd(hi - lo)
		}

		for _, col := range s.Cols {
			cell := partyBudgetCell{Kind: "absent", Bucket: 0}
			key := columnKey(tb, col)
			if key != "" {
				cell.Kind = "blank"
				if m, ok := netto[key]; ok {
					cell.Kind = "fig"
					cell.Change = mrd(m.Change)
					cell.Class = mrdClass(m.Change)
					cell.Bucket = bucketOf(m.Change, peak)
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

// selectArea fills the drill-down for one rammeområde: its bevilgninger, with each
// column's figure and deviation, so a reader can go from "MDG −51,2 i dette
// området" to the posts that make it up. Without this the table was a summary of a
// figure the page had no way to explain.
func selectArea(s partyBudgetSection, area int) partyBudgetSection {
	s.SelectedArea = -1
	var tb *data.PartyBudgetTable
	tables := data.PartyBudgets()
	for i := range tables {
		if tables[i].Area == area {
			tb = &tables[i]
			break
		}
	}
	if tb == nil {
		return s
	}
	s.SelectedArea = area
	s.SelectedDoc = tb.Document
	s.SelectedURL = tb.URL

	// Per-section peak, so the bars in the post table use the same 1..5 scale as the
	// summary rows and cannot imply a magnitude the summary contradicts.
	peak := int64(0)
	for _, rows := range tb.Posts {
		for _, r := range rows {
			for _, m := range r.Amounts {
				v := m.Change
				if v < 0 {
					v = -v
				}
				if v > peak {
					peak = v
				}
			}
		}
	}
	for _, kind := range []string{"utgifter", "inntekter"} {
		for _, r := range tb.Posts[kind] {
			row := partyBudgetPostRow{Kap: r.Kap, Post: r.Post, Formal: r.Formal}
			for _, col := range s.Cols {
				cell := partyBudgetCell{Kind: "absent"}
				if key := columnKey(*tb, col); key != "" {
					cell.Kind = "blank"
					if m, ok := r.Amounts[key]; ok && m.Change != 0 {
						cell.Kind = "fig"
						cell.Change = mrd(m.Change)
						cell.Class = mrdClass(m.Change)
						cell.Bucket = bucketOf(m.Change, peak)
					}
				}
				row.Cells = append(row.Cells, cell)
			}
			s.PostRows = append(s.PostRows, row)
		}
	}
	return s
}
