package data

import (
	"strconv"
	"testing"
)

// crossDimension is every programme point whose linked statement sits in a
// different dimension from the point itself. The two taxonomies overlap
// without being identical, so this is not an error — but it is also not
// something a careless edit should be able to do quietly. A link not on
// this list must be justified here, with a reason, before it can ship.
var crossDimension = map[string]string{
	"ap[18]":  "datasentrenes styring ligger i Teknologisk kapasitet i programmet, men utsagn-11 regner det som Norsk handlefrihet",
	"ap[19]":  "bredbånd ligger i Teknologisk kapasitet i programmet, men utsagn-61 regner støtten som Universell grunnmur",
	"ap[27]":  "barnehageprisen ligger i Universell grunnmur i programmet, men utsagn-34 er Barns muligheter",
	"krf[5]":  "kjernekraft-eierskap ligger i Forvaltning av fellesskapets ressurser i programmet, men utsagn-06 spør om statlig kontroll, i Norsk handlefrihet",
	"mdg[21]": "åpen kildekode ligger i Teknologisk kapasitet i programmet, men utsagn-23 regner det som Innovasjon / omstillingsevne",
	"mdg[33]": "skolemåltidet ligger i Universell grunnmur i programmet, men utsagn-35 er Barns muligheter",
	"mdg[34]": "moderasjonsordningen ligger i Universell grunnmur i programmet, men utsagn-34 er Barns muligheter",
}

// TestEveryProgramLinkResolvesAndIsPlaced is the gate on the hand-set
// programme-to-statement links. The loader already refuses an unresolvable
// id; this adds what it cannot see: the total count, the cross-dimension
// exception list, and the statement coverage. A link added or removed must
// be a deliberate act, not a side effect of another edit.
func TestEveryProgramLinkResolvesAndIsPlaced(t *testing.T) {
	ds := Store
	const wantLinks = 64
	const wantPoints = 59

	links, points := 0, 0
	byStatement := map[string]int{}
	for _, prog := range ds.Programs {
		for i, pt := range prog.Points {
			if len(pt.Statements) == 0 {
				continue
			}
			points++
			key := prog.ID + "[" + strconv.Itoa(i) + "]"
			pointCrosses := false
			for _, sid := range pt.Statements {
				st, ok := ds.Statement(sid)
				if !ok {
					t.Fatalf("%s: unknown statement %q", key, sid)
				}
				links++
				byStatement[sid]++
				if st.Dimension != pt.Dimension {
					pointCrosses = true
				}
			}
			if _, listed := crossDimension[key]; pointCrosses != listed {
				t.Errorf("%s: cross-dimension is %v, but the exception list says %v — "+
					"add it there with a reason, or fix the link", key, pointCrosses, listed)
			}
		}
	}
	if links != wantLinks {
		t.Errorf("programme links: got %d, want %d — a change here must be deliberate", links, wantLinks)
	}
	if points != wantPoints {
		t.Errorf("points carrying a link: got %d, want %d", points, wantPoints)
	}
	if len(byStatement) != 32 {
		t.Errorf("statements reached by a link: got %d, want 32", len(byStatement))
	}
	// The great majority of points must stay unlinked: the join is sparse
	// on purpose, and a rule that linked everything would be inventing.
	if points*4 > len(ds.Programs)*80 {
		t.Errorf("too many points carry a link (%d of 316)", points)
	}
}
