package server

import (
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The programme pages (2026-10-03): /partier lists all nine parties' 2025-2029
// programme points; /parti/{id} is one party. They show what each party *says*
// it will do, quoted from its own programme — tier C, a promise. The statement
// pages show what it *did* from votes and merknader — tier A. Read together,
// a reader can see promise and action in one place, which is the whole point
// of the site. Built and filled by the autopilot (tools/autopilot/), so a
// party with no programme file yet renders as "not read yet", not as empty.

type programPage struct {
	Meta      pageMeta
	Party     data.Party
	Program   data.Program
	HasPoints bool
	Groups    []programGroup
}

type programGroup struct {
	Dimension string
	Points    []data.ProgramPoint
}

type programsIndex struct {
	Meta    pageMeta
	Parties []programPage
	Any     bool
	// The index exists so a reader can navigate, not only scroll. /parti/{id}
	// had been live since 2026-10-03 while /partier never linked to it, so
	// nine working pages were reachable only by typing the URL (found
	// 2026-10-13). Same for reading one theme across parties.
	PartyLinks []partyLink
	Dimensions []dimensionLink
	// Tema is the dimension currently in focus, or empty for the whole list.
	Tema       string
	TemaName   string
	Focused    bool
	TemaMissed bool
}

type partyLink struct {
	ID    string
	Name  string
	Short string
	Count int
}

type dimensionLink struct {
	Slug  string
	Name  string
	Count int // programme points across all parties on this dimension
}

// groupPoints keeps a stable dimension order (the order the statements are
// numbered in) so two parties' pages read the same way down the page.
func groupPoints(ds *data.Dataset, p data.Program) []programGroup {
	order := make([]string, 0, 9)
	for _, s := range ds.Statements {
		found := false
		for _, d := range order {
			if d == s.Dimension {
				found = true
				break
			}
		}
		if !found {
			order = append(order, s.Dimension)
		}
	}
	var groups []programGroup
	for _, dim := range order {
		var pts []data.ProgramPoint
		for _, pt := range p.Points {
			if pt.Dimension == dim {
				pts = append(pts, pt)
			}
		}
		if len(pts) > 0 {
			groups = append(groups, programGroup{Dimension: dim, Points: pts})
		}
	}
	return groups
}

func buildProgramPages(ds *data.Dataset) []programPage {
	pages := make([]programPage, 0, len(ds.Parties))
	for _, party := range ds.Parties {
		pp := programPage{Meta: partyMeta(party, 0), Party: party}
		if prog, ok := ds.Program(party.ID); ok {
			pp.Program = prog
			pp.Groups = groupPoints(ds, prog)
			pp.HasPoints = len(prog.Points) > 0
			pp.Meta = partyMeta(party, len(prog.Points))
		}
		pages = append(pages, pp)
	}
	return pages
}

// dimensionSlug makes a URL-safe key from a dimension name, so a theme can be
// linked and bookmarked without a router per theme. Norwegian letters are
// spelled out rather than dropped: "Barns muligheter" -> "barns-muligheter",
// and "Forvaltning av fellesskapets ressurser / energi" keeps every word.
func dimensionSlug(name string) string {
	repl := strings.NewReplacer("æ", "ae", "ø", "o", "å", "a", "Æ", "ae", "Ø", "o", "Å", "a")
	s := repl.Replace(strings.ToLower(name))
	var b strings.Builder
	dash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		default:
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// buildProgramIndex builds the navigation and, when tema is a known slug,
// narrows every party to that one dimension so the reader can compare them
// side by side. An unknown tema is reported, never silently ignored.
func buildProgramIndex(ds *data.Dataset, pages []programPage, tema string) programsIndex {
	idx := programsIndex{Parties: pages}
	idx.Meta = meta("Partiprogrammene · Ekte valgomat",
		"Alle ni partiers stortingsvalgprogram 2025–2029, samlet tema for tema og sitert med partiets egne ord og side.",
		"/partier")
	byDim := map[string]*dimensionLink{}
	for i, s := range ds.Statements {
		_ = i
		if _, ok := byDim[s.Dimension]; !ok {
			byDim[s.Dimension] = &dimensionLink{Slug: dimensionSlug(s.Dimension), Name: s.Dimension}
		}
	}
	for _, p := range pages {
		if p.HasPoints {
			idx.Any = true
		}
		idx.PartyLinks = append(idx.PartyLinks, partyLink{
			ID: p.Party.ID, Name: p.Party.Name, Short: p.Party.Short, Count: len(p.Program.Points),
		})
		for _, g := range p.Groups {
			if d, ok := byDim[g.Dimension]; ok {
				d.Count += len(g.Points)
			}
		}
	}
	// Only themes the programmes actually cover. The site has more dimensions
	// than the programmes do (Finansiering av fellesskapet, Tilgang ... and
	// the addendum bucket have no programme points), and a theme link that
	// lands on nine empty parties is worse than no link. The front page
	// already shows the whole vision, including empty dimensions.
	for _, dim := range dimensionOrder {
		if d, ok := byDim[dim]; ok && d.Count > 0 {
			idx.Dimensions = append(idx.Dimensions, *d)
		}
	}
	if tema == "" {
		return idx
	}
	want := ""
	for _, d := range idx.Dimensions {
		if d.Slug == tema {
			want = d.Name
			break
		}
	}
	if want == "" {
		idx.TemaMissed = true
		return idx
	}
	idx.Focused = true
	idx.Tema, idx.TemaName = tema, want
	for i := range idx.Parties {
		var kept []programGroup
		for _, g := range idx.Parties[i].Groups {
			if g.Dimension == want {
				kept = append(kept, g)
			}
		}
		idx.Parties[i].Groups = kept
	}
	return idx
}
