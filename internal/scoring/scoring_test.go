package scoring

import (
	"strings"
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

// A cell that holds positions on both sides of zero must not be presented as its
// mean. +2 and -2 average to 0, and 0 on this scale reads as "nøytral/delt",
// which is false about both observations. This is the defect a harvester found in
// the utsagn-12 FrP cell: a documented change of position displayed as neutrality.
func TestPositionTextShowsTheSpreadInsteadOfAMeanThatLies(t *testing.T) {
	zero := 0
	split := Cell{PartyPosition: &zero, PositionMin: -2, PositionMax: 2, PositionYears: []int{2022, 2025}}
	got := split.PositionText()
	if !strings.Contains(got, "til") {
		t.Errorf("PositionText() = %q, want the two ends of the spread", got)
	}
	for _, want := range []string{"-2", "+2", "2022", "2025"} {
		if !strings.Contains(got, strings.Replace(want, "-", "−", 1)) {
			t.Errorf("PositionText() = %q, missing %q", got, want)
		}
	}
	if !split.PositionDisagrees() {
		t.Error("PositionDisagrees() = false for a cell holding -2 and +2")
	}

	// A cell whose positions agree keeps the plain label: a mean is a fair summary
	// there, and replacing it with a range would overstate a single stance.
	two := 2
	agree := Cell{PartyPosition: &two, PositionMin: 2, PositionMax: 2, PositionYears: []int{2025}}
	if agree.PositionDisagrees() {
		t.Error("PositionDisagrees() = true for a cell holding only +2")
	}
	if got := agree.PositionText(); strings.Contains(got, "til") {
		t.Errorf("PositionText() = %q, want a single label for an agreeing cell", got)
	}

	// No position at all must never read as neutral.
	if got := (Cell{}).PositionText(); got != "Ikke vurdert" {
		t.Errorf("PositionText() on an empty cell = %q, want %q", got, "Ikke vurdert")
	}
}

// The consistency label used to read only consistency_flag and never a date, so a
// cell with two dated observations was labelled "ikke nok data" even when it
// documented a reversal. Each branch is asserted separately.
func TestKonsistensLabelReadsTheDates(t *testing.T) {
	items := []data.Evidence{{Position: ptr(1), Date: "2025-06-05"}}

	if got := konsistensLabel(items, []int{2022, 2025}, true); !strings.Contains(got, "Endret posisjon") ||
		!strings.Contains(got, "2022") || !strings.Contains(got, "2025") {
		t.Errorf("konsistensLabel(two years, disagreeing) = %q, want a change naming both years", got)
	}
	if got := konsistensLabel(items, []int{2022, 2025}, false); !strings.Contains(got, "Samme posisjon") {
		t.Errorf("konsistensLabel(two years, agreeing) = %q, want the same-position label", got)
	}
	if got := konsistensLabel(items, []int{2025}, false); !strings.Contains(got, "Ikke nok data") {
		t.Errorf("konsistensLabel(one year) = %q, want the honest not-enough-data label", got)
	}
	if got := konsistensLabel(nil, nil, false); got != "Ingen evidens innhentet ennå" {
		t.Errorf("konsistensLabel(no items) = %q", got)
	}

	// An author-set flag outranks the dates: it is a judgement, not arithmetic.
	flagged := []data.Evidence{{Position: ptr(1), Date: "2025-06-05", ConsistencyFlag: "consistent"}}
	if got := konsistensLabel(flagged, []int{2022, 2025}, true); strings.Contains(got, "Endret") {
		t.Errorf("konsistensLabel(flagg satt) = %q, want the author's flag to win", got)
	}
}

func ptr(n int) *int { return &n }

// The cell's integer is the newest row, not the mean of the rows. Averaging let a
// weaker document drag a stronger one (a +1 vote pulled a +2 own instrument down to
// "1 (for)") and made a documented reversal print "0 (nøytral/delt)", true of neither
// date. The expectation is derived from the evidence here rather than hardcoded, so a
// cell gaining or losing rows does not silently retire the test — and a vacuous pass
// is impossible, because the test fails if no cell exercises the difference.
func TestCellPositionIsTheNewestRowNotTheMean(t *testing.T) {
	// Group the real coded rows the same way ComputeCell will see them.
	type row struct {
		date, id string
		pos      int
	}
	cells := map[[2]string][]row{}
	for _, e := range data.Store.Evidence {
		if e.Position == nil {
			continue
		}
		k := [2]string{e.StatementID, e.PartyID}
		cells[k] = append(cells[k], row{e.Date, e.ID, *e.Position})
	}

	checked, differing := 0, 0
	for k, rows := range cells {
		if len(rows) < 2 {
			continue
		}
		// Newest by (date, id), which is what ComputeCell does.
		newest := rows[0]
		sum := 0
		lo, hi := rows[0].pos, rows[0].pos
		for _, r := range rows {
			sum += r.pos
			if r.pos < lo {
				lo = r.pos
			}
			if r.pos > hi {
				hi = r.pos
			}
			if r.date > newest.date || (r.date == newest.date && r.id > newest.id) {
				newest = r
			}
		}
		mean := sum / len(rows)
		cell := ComputeCell(data.Store, k[0], k[1])
		if cell.PartyPosition == nil {
			t.Fatalf("%s/%s: PartyPosition is nil though %d rows are coded", k[0], k[1], len(rows))
		}
		if *cell.PartyPosition != newest.pos {
			t.Errorf("%s/%s: PartyPosition = %d, want %d (the newest row, not the mean %d)",
				k[0], k[1], *cell.PartyPosition, newest.pos, mean)
		}
		checked++
		if newest.pos != mean {
			differing++
		}
		// A range must be shown whenever the rows differ at all, same side included.
		if lo != hi && !strings.Contains(cell.PositionText(), "til") {
			t.Errorf("%s/%s: PositionText() = %q, want a range for rows %d..%d",
				k[0], k[1], cell.PositionText(), lo, hi)
		}
		// Only a sign flip is a change of position; a change of strength is not.
		if (lo < 0 && hi > 0) != cell.PositionDisagrees() {
			t.Errorf("%s/%s: PositionDisagrees() = %v for rows %d..%d",
				k[0], k[1], cell.PositionDisagrees(), lo, hi)
		}
	}
	if checked == 0 {
		t.Fatal("no cell holds two coded rows: the assertion would pass on nothing")
	}
	if differing == 0 {
		t.Fatal("no cell's newest row differs from its mean: the test cannot tell the two rules apart")
	}
	t.Logf("%d multi-row cells checked, %d where newest and mean disagree", checked, differing)
}
