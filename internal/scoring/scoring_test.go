package scoring

import (
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

func TestComputeCellEmptyIsHonestlyEmpty(t *testing.T) {
	// Use a deliberately empty, isolated dataset rather than a real
	// statement/party pair from data.Store — every real cell gets
	// filled in eventually as evidence-gathering continues, and this
	// test must keep verifying "no evidence" behavior forever, not
	// just until someone documents utsagn-05/ap (which happened
	// 2026-09-24). The cell must say so plainly — never silently
	// score it as if a neutral position had been documented.
	empty := &data.Dataset{}
	cell := ComputeCell(empty, "utsagn-does-not-exist", "party-does-not-exist")
	if cell.PartyPosition != nil {
		t.Errorf("PartyPosition = %v, want nil", *cell.PartyPosition)
	}
	if cell.Evidensstyrke != "Ingen evidens innhentet ennå" {
		t.Errorf("Evidensstyrke = %q, want the explicit empty label", cell.Evidensstyrke)
	}
	if _, ok := cell.SaksmatchPercent(2); ok {
		t.Error("SaksmatchPercent should report ok=false when no position is coded")
	}
}

func TestSaksmatchPercentMath(t *testing.T) {
	two := 2
	cell := Cell{PartyPosition: &two}

	tests := []struct {
		userAnswer int
		wantPct    int
	}{
		{userAnswer: 2, wantPct: 100}, // exact match
		{userAnswer: -2, wantPct: 0},  // maximum disagreement on a -2..+2 scale
		{userAnswer: 0, wantPct: 50},  // halfway
	}
	for _, tt := range tests {
		pct, ok := cell.SaksmatchPercent(tt.userAnswer)
		if !ok {
			t.Fatalf("SaksmatchPercent(%d): ok=false, want true", tt.userAnswer)
		}
		if pct != tt.wantPct {
			t.Errorf("SaksmatchPercent(%d) = %d, want %d", tt.userAnswer, pct, tt.wantPct)
		}
	}
}

func TestPositionLabelNeverPrintsAPointerAddress(t *testing.T) {
	if got := PositionLabel(nil); got != "Ikke vurdert" {
		t.Errorf("PositionLabel(nil) = %q, want %q", got, "Ikke vurdert")
	}
	two := 2
	if got := PositionLabel(&two); got == "" {
		t.Error("PositionLabel(&2) returned empty string")
	}
}

func TestPositionWordIsTheDominantSignal(t *testing.T) {
	one, two, minusOne := 1, 2, -1
	cases := []struct {
		p    *int
		want string
	}{
		{nil, "Ikke vurdert"},
		{&one, "FOR"},
		{&two, "STERKT FOR"},
		{&minusOne, "MOT"},
	}
	for _, tt := range cases {
		if got := PositionWord(tt.p); got != tt.want {
			t.Errorf("PositionWord(%v) = %q, want %q", tt.p, got, tt.want)
		}
	}
}

func TestCellHasEvidenceGatesTheCollapsedEmptyState(t *testing.T) {
	empty := Cell{EvidensstyrkeCount: 0}
	if empty.HasEvidence() {
		t.Error("HasEvidence() = true for a cell with zero evidence items")
	}
	nonEmpty := Cell{EvidensstyrkeCount: 1}
	if !nonEmpty.HasEvidence() {
		t.Error("HasEvidence() = false for a cell with one evidence item")
	}
}

func TestEvidensstyrkeCapClampsAUniformlyCappedCell(t *testing.T) {
	// A cell whose only evidence item was hand-reviewed and found to
	// rest solely on a merknad with no resolving vote (see
	// data.Evidence.EvidensstyrkeCap doc, and the 5 utsagn-06 files it
	// was added for 2026-09-26) must not show "Sterk" just because the
	// source_tier is A.
	items := []data.Evidence{
		{SourceTier: "A", EvidensstyrkeCap: "moderat"},
	}
	if got := evidensstyrkeLabel(len(items), items); got != "Moderat" {
		t.Errorf("evidensstyrkeLabel with uniform moderat cap = %q, want %q", got, "Moderat")
	}
}

func TestEvidensstyrkeCapNeverRaisesTheLabel(t *testing.T) {
	// A cap only lowers; a "svak" cap on a bestTier-B cell must stay
	// "Svak", not jump up to "Moderat".
	items := []data.Evidence{
		{SourceTier: "B", EvidensstyrkeCap: "svak"},
	}
	if got := evidensstyrkeLabel(len(items), items); got != "Svak" {
		t.Errorf("evidensstyrkeLabel with svak cap on bestTier B = %q, want %q", got, "Svak")
	}
}

func TestEvidensstyrkeCapOnlyAppliesWhenEveryItemAgrees(t *testing.T) {
	// A single corroborating item without a cap (e.g. an actual
	// recorded vote for the same cell) must pull the label back up —
	// the cap exists for cells whose ENTIRE evidence was found wanting,
	// not for a hand-review note on just one of several items.
	items := []data.Evidence{
		{SourceTier: "A", EvidensstyrkeCap: "moderat"},
		{SourceTier: "A"},
	}
	if got := evidensstyrkeLabel(len(items), items); got != "Sterk" {
		t.Errorf("evidensstyrkeLabel with a mixed cap = %q, want %q (cap should not apply)", got, "Sterk")
	}
}

func TestEvidensstyrkeNoCapUnaffected(t *testing.T) {
	items := []data.Evidence{{SourceTier: "A"}}
	if got := evidensstyrkeLabel(len(items), items); got != "Sterk" {
		t.Errorf("evidensstyrkeLabel with no cap = %q, want %q", got, "Sterk")
	}
}

func TestHoyreKjernekraftEierskapCellIsClampedByRealData(t *testing.T) {
	// End-to-end regression against the real dataset, not synthetic
	// items: utsagn-06/h's only evidence is the hand-reviewed
	// merknad-only item capped 2026-09-26 (see
	// 27-hoyre-privat-finansiering.yaml) — ComputeCell must actually
	// apply the cap through the real load path, not just in the
	// unit-level tests above.
	cell := ComputeCell(data.Store, "utsagn-06", "h")
	if cell.Evidensstyrke != "Moderat" {
		t.Errorf("utsagn-06/h Evidensstyrke = %q, want %q", cell.Evidensstyrke, "Moderat")
	}
}

func TestStrengthBasisIgnoresUncodedItemsWhenAnyIsCoded(t *testing.T) {
	// A programme (tier C) is the only coded evidence; an uncoded tier-A
	// context item sits next to it. The label must describe the
	// programme, so the basis is the coded item alone.
	one := 1
	items := []data.Evidence{
		{SourceTier: "C", Position: &one},
		{SourceTier: "A"},
	}
	basis := strengthBasis(items)
	if len(basis) != 1 || basis[0].SourceTier != "C" {
		t.Fatalf("strengthBasis = %+v, want only the coded tier-C item", basis)
	}
	if got := evidensstyrkeLabel(len(items), basis); got != "Svak" {
		t.Errorf("label = %q, want %q", got, "Svak")
	}
}

func TestStrengthBasisIsEmptyWhenNoneIsCoded(t *testing.T) {
	items := []data.Evidence{{SourceTier: "A"}, {SourceTier: "C"}}
	basis := strengthBasis(items)
	if len(basis) != 0 {
		t.Errorf("strengthBasis kept %d items, want 0 (no coded item, so no position to rate)", len(basis))
	}
	if got := evidensstyrkeLabel(len(items), basis); got != noCodedPosition {
		t.Errorf("label = %q, want %q", got, noCodedPosition)
	}
}

func TestUncodedCellIsNeverRatedOnTheStrengthScale(t *testing.T) {
	// Invariant over the whole dataset: a cell with no coded position
	// has no position whose documentation could be Svak, Moderat or
	// Sterk. It used to fall back to rating its uncoded context, so Ap
	// on utsagn-57 showed "Ikke vurdert" next to "Sterk".
	uncodedWithSources := 0
	for _, st := range data.Store.Statements {
		for _, p := range data.Store.Parties {
			cell := ComputeCell(data.Store, st.ID, p.ID)
			if cell.PartyPosition != nil {
				continue
			}
			want := "Ingen evidens innhentet ennå"
			if cell.EvidensstyrkeCount > 0 {
				want = noCodedPosition
				uncodedWithSources++
			}
			if cell.Evidensstyrke != want {
				t.Errorf("%s/%s: uncoded cell with %d source(s) has Evidensstyrke %q, want %q",
					st.ID, p.ID, cell.EvidensstyrkeCount, cell.Evidensstyrke, want)
			}
			if cell.EvidensstyrkeBasis != 0 {
				t.Errorf("%s/%s: uncoded cell has EvidensstyrkeBasis %d, want 0", st.ID, p.ID, cell.EvidensstyrkeBasis)
			}
		}
	}
	// The loop must have met the cells this test exists for, or it
	// passes vacuously. Was 3 until 2026-10-13, then 2, as Rødt's and
	// Sp's programmer were coded on utsagn-45 and those cells stopped
	// being uncoded-with-sources cells.
	if uncodedWithSources < 1 {
		t.Errorf("found %d uncoded cells with sources, want at least 1 (utsagn-57/ap)", uncodedWithSources)
	}
}

func TestKnownUncodedCellsSayNoCodedPosition(t *testing.T) {
	for _, c := range []struct{ statement, party string }{
		{"utsagn-57", "ap"},
	} {
		cell := ComputeCell(data.Store, c.statement, c.party)
		if cell.PartyPosition != nil {
			t.Errorf("%s/%s is coded; this test assumes it is not", c.statement, c.party)
		}
		if cell.EvidensstyrkeCount == 0 {
			t.Errorf("%s/%s has no sources; this test assumes it has", c.statement, c.party)
		}
		if cell.Evidensstyrke != noCodedPosition {
			t.Errorf("%s/%s Evidensstyrke = %q, want %q", c.statement, c.party, cell.Evidensstyrke, noCodedPosition)
		}
	}
}

func TestProgrammeOnlyCellsOnUtsagn45AreNotSterk(t *testing.T) {
	// End-to-end against the real dataset: each of these cells is coded
	// only from a party programme (tier C). Innst. 464 S sits next to
	// each as uncoded tier-A context and used to lift all six to
	// "Sterk" (2026-10-02 trust audit).
	for _, p := range []string{"ap", "frp", "h", "krf", "mdg", "sv"} {
		cell := ComputeCell(data.Store, "utsagn-45", p)
		if cell.PartyPosition == nil {
			t.Errorf("utsagn-45/%s has no coded position; this test assumes one", p)
		}
		if cell.Evidensstyrke != "Svak" {
			t.Errorf("utsagn-45/%s Evidensstyrke = %q, want %q", p, cell.Evidensstyrke, "Svak")
		}
	}
}

func TestUncodedVoteKeepsTheMerknadCapOnUtsagn06(t *testing.T) {
	// Ap, Sp and MDG are coded +1 from a capped merknad (files 29-31),
	// and their uncoded, uncapped vote against Rødt's forslag 30-32
	// (files 701-703) is listed next to it. The vote must be counted,
	// but must not break the cap and lift the cell to "Sterk".
	for _, p := range []string{"ap", "sp", "mdg"} {
		cell := ComputeCell(data.Store, "utsagn-06", p)
		if cell.EvidensstyrkeCount != 2 {
			t.Errorf("utsagn-06/%s EvidensstyrkeCount = %d, want 2 (merknad + vote)", p, cell.EvidensstyrkeCount)
		}
		if cell.EvidensstyrkeBasis != 1 {
			t.Errorf("utsagn-06/%s EvidensstyrkeBasis = %d, want 1 (only the merknad is coded)", p, cell.EvidensstyrkeBasis)
		}
		if cell.PartyPosition == nil || *cell.PartyPosition != 1 {
			t.Errorf("utsagn-06/%s PartyPosition changed, want +1", p)
		}
		if cell.Evidensstyrke != "Moderat" {
			t.Errorf("utsagn-06/%s Evidensstyrke = %q, want %q", p, cell.Evidensstyrke, "Moderat")
		}
	}
}
