package data

import (
	"regexp"
	"testing"
)

// The statements in this corpus are written as «do X, not Y». X alone is
// usually what every party already claims — gradual expansion, more money, a
// better service — so a row that affirms only X says nothing about the choice
// the statement actually poses. The clause that decides the cell is the
// rejection of Y, and a source that is merely silent about Y cannot be coded on
// it. See tools/autopilot/LOOP-TRACKS.md for the case that forced this field:
// a Rødt row on tannhelse that was promoted on the instrument and withdrawn the
// same minute it was reviewed against this rule.
//
// The field is opt-in per statement because the choosing is a judgement about
// meaning, and this test only enforces that the judgement was made and is one
// of the declared values. It cannot check that the value is right.
var contrastiveClause = regexp.MustCompile(`,\s*(ikke|uten)\s`)

var codesOnValues = map[string]bool{
	"avvist_alternativ": true,
	"virkemiddel":       true,
}

func TestEveryContrastiveStatementDeclaresItsClause(t *testing.T) {
	declared := 0
	for _, s := range Store.Statements {
		hasContrast := contrastiveClause.MatchString(s.Text)
		switch {
		case hasContrast && s.CodesOn == "":
			t.Errorf("%s poses a contrastive clause but declares no codes_on; "+
				"say which clause a row must address", s.ID)
		case hasContrast && !codesOnValues[s.CodesOn]:
			t.Errorf("%s: codes_on %q is not one of avvist_alternativ, virkemiddel",
				s.ID, s.CodesOn)
		case !hasContrast && s.CodesOn != "":
			t.Errorf("%s declares codes_on %q but poses a single clause; "+
				"the field would be decorative", s.ID, s.CodesOn)
		}
		if s.CodesOn != "" {
			declared++
		}
	}
	// A gate that finds nothing is not green, it is blind: if the yaml field
	// stopped loading, every statement would read as undeclared and this test
	// would fail above rather than pass here. The floor makes that explicit.
	if declared < 5 {
		t.Fatalf("only %d statements declare codes_on; the field has probably "+
			"stopped loading rather than the corpus having changed", declared)
	}
}
