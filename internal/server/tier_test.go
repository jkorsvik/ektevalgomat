package server

import (
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// TestHasStrongTier pins the predicate that decides two published figures:
// how many cells /metode says rest on a programme alone, and whether /om may
// say the positions on utsagn 45 "kommer fra partiprogrammene". Both read
// hasStrongTier, so a wrong line here moves a number and a sentence at once.
//
// The case that matters is the uncoded A item. An evidence row with no
// position is context or a vote left uncoded on purpose; it is listed on the
// page but it decides nothing, and counting it as strong would hide exactly
// the cells the method page exists to disclose (2026-10-02 trust audit: one
// uncoded tier-A item lifted six programme-only cells to "Sterk").
func TestHasStrongTier(t *testing.T) {
	pos := 1
	cases := []struct {
		name  string
		items []data.Evidence
		want  bool
	}{
		{"no evidence at all", nil, false},
		{"empty tier with a position", []data.Evidence{{Position: &pos}}, false},
		{"programme only (C)", []data.Evidence{{Position: &pos, SourceTier: "C"}}, false},
		{"secondary account (D)", []data.Evidence{{Position: &pos, SourceTier: "D"}}, false},
		{"A source with no coded position", []data.Evidence{{SourceTier: "A"}}, false},
		{"B source with no coded position", []data.Evidence{{SourceTier: "B"}}, false},
		{"A source with a position", []data.Evidence{{Position: &pos, SourceTier: "A"}}, true},
		{"B source with a position", []data.Evidence{{Position: &pos, SourceTier: "B"}}, true},
		{
			"uncoded A next to a coded C",
			[]data.Evidence{{SourceTier: "A"}, {Position: &pos, SourceTier: "C"}},
			false,
		},
		{
			"coded A after a coded C",
			[]data.Evidence{{Position: &pos, SourceTier: "C"}, {Position: &pos, SourceTier: "A"}},
			true,
		},
	}
	fired := 0
	for _, c := range cases {
		got := hasStrongTier(c.items)
		if got != c.want {
			t.Errorf("%s: hasStrongTier = %v, want %v", c.name, got, c.want)
			fired++
		}
	}
	if fired == 0 {
		t.Logf("all %d tier cases behaved as the predicate claims", len(cases))
	}
}
