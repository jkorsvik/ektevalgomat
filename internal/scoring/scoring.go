// Package scoring computes the three separate indicators the source memo
// insists on for every statement×party cell: saksmatch, konsistens, and
// evidensstyrke. None of these are ever combined into one score — that
// collapse is exactly what the memo says a real valgomat must not do,
// because a party can be highly consistent in a policy the user disagrees
// with, or match the user's values in its program while never having
// acted on them.
package scoring

import (
	"sort"
	"strconv"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// Cell is the fully computed result for one statement×party pair, ready
// to render or serialize to the quiz's client-side JSON.
type Cell struct {
	StatementID string `json:"statement_id"`
	PartyID     string `json:"party_id"`

	// PartyPosition is the NEWEST coded row's position, not the mean of the
	// rows. It is nil when no evidence has coded a stance yet — the quiz must
	// show "ikke vurdert", never silently treat this as neutral (0), which would
	// manufacture agreement or disagreement nobody has documented.
	//
	// Why newest and not the mean: a cell holds one row per document, and the
	// rows are not independent observations of one fact. Averaging them let a
	// weaker document drag a stronger one (a +1 vote pulled a +2 own instrument
	// down to "1 (for)") and made a documented reversal print "0 (nøytral/delt)",
	// which is true of neither date. The newest row answers the question a reader
	// asks of a cell — where does the party stand now — and the older rows are
	// the history, which PositionText shows.
	PartyPosition *int `json:"party_position"`

	// PositionMin and PositionMax bound the coded positions behind
	// PartyPosition, and PositionText shows them as a range whenever they
	// differ. A same-side difference counts: +2 and +1 are both "for", but the
	// cell resting on both must not silently print the stronger one alone. Ties
	// on date are broken by evidence id, so the choice is deterministic.
	PositionMin int `json:"position_min"`
	PositionMax int `json:"position_max"`

	// PositionYears holds the distinct years the coded positions are dated to,
	// ascending. A cell with two of them can say something about change over
	// time, which is what the consistency label needs and never used to read.
	PositionYears []int `json:"position_years"`

	Evidensstyrke      string `json:"evidensstyrke"`       // Ingen evidens / Ingen kodet posisjon / Svak / Moderat / Sterk
	EvidensstyrkeCount int    `json:"evidensstyrke_count"` // how many evidence items are listed for this cell
	EvidensstyrkeBasis int    `json:"evidensstyrke_basis"` // how many of them the label is computed from, see strengthBasis
	Konsistens         string `json:"konsistens"`          // human-readable, see konsistensLabel
}

// noCodedPosition is the strength label for a cell that lists sources
// but codes no position from them. There is nothing to rate, so it is
// not a point on the Svak/Moderat/Sterk scale.
const noCodedPosition = "Ingen kodet posisjon"

// HasEvidence reports whether any evidence at all has been collected for
// this cell. Templates use this to decide between a single collapsed
// "ikke vurdert ennå" line and the full position/documentation/history
// breakdown — never both, and never three rows that all separately say
// "no evidence yet" (2026-09-23 design review).
func (c Cell) HasEvidence() bool {
	return c.EvidensstyrkeCount > 0
}

// SaksmatchPercent scores the user's answer (-2..+2) against this cell's
// coded party position on the same -2..+2 scale. Returns (percent, true)
// when a position exists, or (0, false) when it doesn't — callers must
// check the bool and display "ikke vurdert" rather than a fabricated 0%.
func (c Cell) SaksmatchPercent(userAnswer int) (int, bool) {
	if c.PartyPosition == nil {
		return 0, false
	}
	diff := userAnswer - *c.PartyPosition
	if diff < 0 {
		diff = -diff
	}
	// Max possible distance on a -2..+2 scale is 4.
	pct := 100 - (diff*100)/4
	return pct, true
}

// ComputeCell builds the Cell for one statement×party pair from every
// piece of party-specific evidence attached to that statement (secondary
// tags included). Institutional evidence (no party_id) never reaches
// here — it's context on the statement page, not a party score.
func ComputeCell(ds *data.Dataset, statementID, partyID string) Cell {
	items := ds.EvidenceForCell(statementID, partyID)

	cell := Cell{StatementID: statementID, PartyID: partyID}

	// basis holds exactly the coded items: the position and the
	// strength label both come from it, so a cell has a position if and
	// only if it has a strength basis.
	var codedPositions []int
	basis := strengthBasis(items)
	for _, e := range basis {
		codedPositions = append(codedPositions, *e.Position)
	}

	if len(codedPositions) > 0 {
		lo, hi := codedPositions[0], codedPositions[0]
		for _, p := range codedPositions {
			if p < lo {
				lo = p
			}
			if p > hi {
				hi = p
			}
		}
		cell.PositionMin, cell.PositionMax = lo, hi

		// Newest dated row wins. An undated row sorts below every dated one, so a
		// dated document always outranks a row whose date the source did not give;
		// the id breaks a tie on date, which keeps the choice deterministic
		// without inventing an order the data does not have.
		newest := basis[0]
		for _, e := range basis[1:] {
			if e.Date > newest.Date || (e.Date == newest.Date && e.ID > newest.ID) {
				newest = e
			}
		}
		pos := *newest.Position
		cell.PartyPosition = &pos
	}

	// Distinct years only, ascending: a repeated date is one observation.
	yearSeen := map[int]bool{}
	for _, e := range basis {
		if y, ok := data.YearOf(e.Date); ok && !yearSeen[y] {
			yearSeen[y] = true
			cell.PositionYears = append(cell.PositionYears, y)
		}
	}
	sort.Ints(cell.PositionYears)

	cell.EvidensstyrkeCount = len(items)
	cell.EvidensstyrkeBasis = len(basis)
	cell.Evidensstyrke = evidensstyrkeLabel(len(items), basis)
	cell.Konsistens = konsistensLabel(items, cell.PositionYears, cell.PositionDisagrees())

	return cell
}

// strengthBasis returns the items the strength label is computed from:
// the items with a coded position, and nothing else. An item with no
// position (context, or a vote left uncoded on purpose) is still listed
// and counted on the page, but it must not make the coded position look
// stronger, and it must not break a cap that every coded item shares
// (2026-10-02 trust audit: one uncoded tier-A item lifted six
// programme-only cells on utsagn-45 to "Sterk"). When no item is coded
// there is no position to document, so the basis is empty and the cell
// says so (see noCodedPosition), rather than rating the strength of a
// position nobody has coded (Ap on utsagn-57, Rødt and Sp on utsagn-45).
func strengthBasis(items []data.Evidence) []data.Evidence {
	var coded []data.Evidence
	for _, e := range items {
		if e.Position != nil {
			coded = append(coded, e)
		}
	}
	return coded
}

// evidensstyrkeLabel owns every strength label. listed is how many items
// the cell shows, basis the coded ones among them (see strengthBasis).
// Nothing listed: no evidence. Listed but none coded: no position to
// rate, so noCodedPosition, not a point on the scale. Otherwise it
// buckets by the best source tier in basis. A single
// tier-A item (an official vote, a law, a register entry) is already
// strong; anything else needs either volume or accepts a lower label —
// this is a coarse, honest proxy for "how much and how authoritative is
// what we've actually collected", not a statistical confidence interval.
//
// If every item feeding this cell carries the same evidensstyrke_cap
// (see data.Evidence.EvidensstyrkeCap), the label is clamped down to
// that cap even when the best tier is A — a hand-verified exception for the
// narrow set of cells whose only evidence is a party's own merknad with
// no vote resolving the specific question asked, not a general downgrade
// of partiets_begrunnelse (which is real, tier-A evidence in the
// overwhelming majority of cases; see 2026-09-26 review).
func evidensstyrkeLabel(listed int, basis []data.Evidence) string {
	if listed == 0 {
		return "Ingen evidens innhentet ennå"
	}
	if len(basis) == 0 {
		return noCodedPosition
	}

	bestTier := basis[0].SourceTier // alphabetically earliest = highest authority
	for _, e := range basis[1:] {
		if e.SourceTier < bestTier {
			bestTier = e.SourceTier
		}
	}

	label := ""
	switch {
	case bestTier == "A":
		label = "Sterk"
	case bestTier == "B":
		label = "Moderat"
	default:
		label = "Svak"
	}

	if cap, ok := uniformCap(basis); ok {
		label = clampToCap(label, cap)
	}
	return label
}

// uniformCap reports the evidensstyrke_cap shared by every item in
// items, and false if any item has no cap or the caps disagree — a cap
// only applies when the ENTIRE cell's evidence was hand-reviewed and
// found wanting, never when just one of several corroborating items
// happens to carry one.
func uniformCap(items []data.Evidence) (string, bool) {
	cap := items[0].EvidensstyrkeCap
	if cap == "" {
		return "", false
	}
	for _, e := range items[1:] {
		if e.EvidensstyrkeCap != cap {
			return "", false
		}
	}
	return cap, true
}

// clampToCap never raises a label, only lowers it to the cap when the
// computed label was stronger.
func clampToCap(label, cap string) string {
	rank := map[string]int{"Svak": 0, "Moderat": 1, "Sterk": 2}
	capLabel := map[string]string{"svak": "Svak", "moderat": "Moderat"}
	capRank := map[string]int{"svak": 0, "moderat": 1}
	r, labelKnown := rank[label]
	cr, capKnown := capRank[cap]
	if labelKnown && capKnown && cr < r {
		return capLabel[cap]
	}
	return label
}

// konsistensLabel reflects each item's own consistency_flag when the
// author of that evidence set one. Phase 1's real dataset is almost all
// single-snapshot observations (one vote, one law, one proposal) with no
// repeated action over time to compare — so for nearly every cell this
// honestly reports "ikke nok data" rather than computing a score the
// underlying data can't support yet. Deliberately short (2026-09-23
// design review: don't repeat "ingen evidens innhentet ennå" three times
// on one row) — templates decide whether to show this field at all;
// see Cell.HasEvidence.
func konsistensLabel(items []data.Evidence, years []int, disagrees bool) string {
	if len(items) == 0 {
		return "Ingen evidens innhentet ennå"
	}
	// An author-set flag outranks the dates: it records a judgement about the
	// sources that arithmetic cannot make.
	flags := map[string]bool{}
	for _, e := range items {
		if e.ConsistencyFlag != "" {
			flags[e.ConsistencyFlag] = true
		}
	}
	if len(flags) == 1 {
		for f := range flags {
			return humanConsistencyFlag(f)
		}
	}
	// Two dated observations of the same cell can answer the question the flag
	// used to stand in for. This is the case the label used to miss: it read no
	// dates at all, so a cell with a documented change of position still said
	// "ikke nok data".
	if len(years) >= 2 {
		span := strconv.Itoa(years[0]) + " og " + strconv.Itoa(years[len(years)-1])
		if disagrees {
			return "Endret posisjon mellom " + span
		}
		return "Samme posisjon i " + span
	}
	return "Ikke nok data til å vurdere utvikling over tid"
}

// PositionText renders this cell's position the way a reader must see it.
//
// It exists because PartyPosition alone can lie about a cell that holds
// observations on both sides: +2 and -2 average to 0, and 0 on this scale means
// "nøytral/delt", which is false about both. When the spread straddles zero this
// returns the two ends and the year span they come from, so the reader sees a
// change of position instead of a fabricated neutrality.
func (c Cell) PositionText() string {
	if c.PartyPosition == nil {
		return PositionLabel(nil)
	}
	if c.PositionMin != c.PositionMax {
		return PositionLabel(&c.PositionMin) + " til " + PositionLabel(&c.PositionMax) +
			c.yearSpanSuffix()
	}
	return PositionLabel(c.PartyPosition)
}

// PositionDisagrees reports whether the cell holds coded positions on both sides
// of zero. A mean over such a cell is not a summary of a stance, it is an
// artefact of arithmetic.
func (c Cell) PositionDisagrees() bool {
	return c.PartyPosition != nil && c.PositionMin < 0 && c.PositionMax > 0
}

// yearSpanSuffix renders " (2022 og 2025)" when the cell is dated to more than one
// year, and nothing when it is not. It never guesses a span from a single year.
func (c Cell) yearSpanSuffix() string {
	if len(c.PositionYears) < 2 {
		return ""
	}
	parts := make([]string, 0, len(c.PositionYears))
	for _, y := range c.PositionYears {
		parts = append(parts, strconv.Itoa(y))
	}
	return " (" + strings.Join(parts, " og ") + ")"
}

// PositionLabel renders a coded -2..+2 position (or nil) as Norwegian text
// for templates. Kept as a plain function taking *int (not a Cell method)
// so it also works directly on data.Evidence.Position in the evidence-chain
// view, not just on a computed Cell.
func PositionLabel(p *int) string {
	if p == nil {
		return "Ikke vurdert"
	}
	labels := map[int]string{
		-2: "−2 (sterkt mot)",
		-1: "−1 (mot)",
		0:  "0 (nøytral/delt)",
		1:  "+1 (for)",
		2:  "+2 (sterkt for)",
	}
	if l, ok := labels[*p]; ok {
		return l
	}
	return "Ikke vurdert"
}

// PositionWord renders the same coded position as the single dominant
// word a general reader should see first — "FOR", "MOT", "STERKT FOR" —
// with the exact -2..+2 number kept as secondary detail (PositionLabel or
// the raw value), never the other way around. 2026-09-23 design review:
// "+1 ... ikke bør være den mest fremtredende tingen. For folk flest er
// FOR mer forståelig."
func PositionWord(p *int) string {
	if p == nil {
		return "Ikke vurdert"
	}
	words := map[int]string{
		-2: "STERKT MOT",
		-1: "MOT",
		0:  "NØYTRALT / DELT",
		1:  "FOR",
		2:  "STERKT FOR",
	}
	if w, ok := words[*p]; ok {
		return w
	}
	return "Ikke vurdert"
}

func humanConsistencyFlag(flag string) string {
	labels := map[string]string{
		"consistent":            "Konsistent: uttalt prinsipp og handling peker samme vei",
		"tension":               "Spenning: handlingen skaper en reell, men forklarbar spenning",
		"contradiction":         "Motsigelse: to sammenlignbare handlinger peker direkte motsatt",
		"changed_position":      "Endret standpunkt (eksplisitt)",
		"coalition_constraint":  "Regjeringssamarbeid: program og kompromiss kan skilles dokumentert",
		"insufficient_evidence": "Utilstrekkelig grunnlag for konklusjon",
		"not_comparable":        "Sakene ligner, men premissene er for forskjellige",
	}
	if l, ok := labels[flag]; ok {
		return l
	}
	return flag
}
