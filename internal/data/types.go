// Package data holds the evidence model behind ektevalgomat: statements,
// parties, and the evidence objects that link them to primary sources.
//
// Deliberately NOT the full 42×9 schema from the source research memo
// (docs/research-memo.md). Phase 1 covers one dimension (Demokratisk
// åpenhet, statements 1-5) and the three cases that memo already had seed
// evidence for. Fields that memo's evidence table describes but phase 1
// has no real data for yet (beneficiaries, cost_bearers, amount_nok,
// foreign_dependency, time_horizon, reviewer_1/reviewer_2, review_status)
// are left out on purpose — an empty struct field is indistinguishable
// from "not collected yet" in a Go zero value, and that ambiguity is
// exactly what this project exists to refuse. Add a field only when a
// real evidence item needs it.
package data

// Statement is one of the value statements the valgomat quiz asks about.
// Phase 1 only populates statements 1-5 (dimension: Demokratisk åpenhet);
// the other 37 from the source memo's 9-dimension model do not exist yet.
type Statement struct {
	ID              string `yaml:"id"`
	Number          int    `yaml:"number"`
	Dimension       string `yaml:"dimension"`
	PrimaryQuestion string `yaml:"primary_question"`
	CaseTag         string `yaml:"case_tag"`
	Text            string `yaml:"text"`
	// ShortName is a 2-5 word label for use in navigation, breadcrumbs,
	// and result overviews — anywhere the full multi-line statement text
	// would be too heavy to serve as a wayfinding element. The full Text
	// is still the only thing ever shown where the user is actually
	// reading or answering the statement (2026-09-23 design review:
	// "Utsagn 45-overskriften er for lang til å fungere som
	// navigasjonselement").
	ShortName string `yaml:"short_name"`
	// PracticalStakes spells out, in plain and strictly mechanical
	// terms, what a "sterkt enig" and a "sterkt uenig" answer would mean
	// if acted on — never an evaluation of which side is right, and
	// never a downstream empirical claim ("dette reduserer korrupsjon")
	// that would itself need its own source. Just the direct, definitional
	// consequence of the statement's own text, spelled out for someone
	// unfamiliar with the specific policy area who otherwise has nothing
	// but the raw statement to go on (2026-09-25, review feedback: "hva
	// sterkt uenig / sterkt enig ville sagt for deg som enkelt
	// menneske... satt på spissen, så det er skikkelig lett å forstå").
	// Present on all 55 statements since 2026-09-27 (see
	// statement_form_test.go, which asserts both fields are non-empty for
	// every statement). Still a pointer/omitempty in the schema rather
	// than a required field, so a future statement added without stakes
	// fails the gate loudly instead of loading with two silently blank
	// strings.
	PracticalStakes *PracticalStakes `yaml:"practical_stakes,omitempty"`
}

// PracticalStakes is the plain-language pair of "if this extreme answer
// won, what would concretely be true" statements attached to a Statement.
type PracticalStakes struct {
	Enig  string `yaml:"enig"`
	Uenig string `yaml:"uenig"`
}

// Party is one of the nine Storting party groups for the 2025-2029 period.
type Party struct {
	ID    string `yaml:"id"`
	Name  string `yaml:"name"`
	Short string `yaml:"short"`
}

// Evidence is a single observation linking a statement to a source. Most
// fields mirror the source memo's central evidence table 1:1 (evidence_id,
// statement_id, party_id, date, evidence_type, claim_action, source_tier,
// position, claim_confidence, retrieved_at); this is the smallest schema
// that still forces every party-scoring claim to name where it came from.
//
// PartyID is empty for institutional or non-party evidence (a Storting
// vote as a body, a state law, an unattributed influence operation) — the
// memo's own schema marks party_id nullable for exactly this reason.
// Position is a pointer so "no coded position yet" (nil) is distinguishable
// from "coded as neutral" (0) — collapsing those two would silently invent
// a stance nobody has actually taken.
type Evidence struct {
	ID                    string   `yaml:"id"`
	StatementID           string   `yaml:"statement_id"`
	SecondaryStatementIDs []string `yaml:"secondary_statement_ids"`
	PartyID               string   `yaml:"party_id"`
	Actor                 string   `yaml:"actor"`
	Date                  string   `yaml:"date"`
	EvidenceType          string   `yaml:"evidence_type"`
	// EvidenceClass distinguishes what KIND of claim this item makes, so
	// the UI can show it visually (never color-only, per WCAG) rather
	// than only in a tooltip. One of: fakta (a documented event/outcome
	// with no attributive judgment attached), partiets_begrunnelse (the
	// evidence IS a party's own proposal/reasoning in their own words —
	// almost always paired with VerbatimExcerpt), analyse (an assessed
	// or inferred judgment, e.g. an attribution finding — AttributionConfidence
	// is set to a real value on these), uklart (the source itself
	// reports the finding as unresolved or contested, e.g.
	// AttributionConfidence == "ukjent"). See loader_test.go for the
	// invariant that keeps this rule from silently drifting.
	EvidenceClass string `yaml:"evidence_class"`
	// EvidensstyrkeCap is an explicit, hand-set override that caps how
	// strong scoring.evidensstyrkeLabel is allowed to call this item's
	// cell, regardless of source_tier. Almost always empty (no cap) —
	// source_tier A already means "official Storting record", and for
	// the overwhelming majority of partiets_begrunnelse items that
	// record IS a resolved vote or a formally filed document, which is
	// real evidence, not spin. Set this only for the narrow case where
	// the position rests SOLELY on a party's own merknad/argument with
	// no vote resolving that specific question — and only when the
	// item's own ContextNote already says so explicitly (never inferred
	// from evidence_class or status alone: both vocabularies mix real
	// vote outcomes and bare arguments too inconsistently to tell them
	// apart automatically, see 2026-09-26 review). One of "" (no cap,
	// default), "moderat", "svak".
	EvidensstyrkeCap string `yaml:"evidensstyrke_cap,omitempty"`
	// ScanSummary is the one-sentence "what happened" shown at the
	// evidence card's first disclosure level (Scan) — hand-written, not
	// derived from ClaimAction, because Norwegian legal references
	// ("Prop.", "Innst.", "St.") make naive sentence-splitting unsafe.
	ScanSummary           string `yaml:"scan_summary"`
	ClaimAction           string `yaml:"claim_action"`
	VerbatimExcerpt       string `yaml:"verbatim_excerpt"`
	Status                string `yaml:"status"`
	Position              *int   `yaml:"position"`
	SourceTier            string `yaml:"source_tier"`
	SourceName            string `yaml:"source_name"`
	SourceURL             string `yaml:"source_url"`
	ClaimConfidence       string `yaml:"claim_confidence"`
	AttributionConfidence string `yaml:"attribution_confidence"`
	ConsistencyFlag       string `yaml:"consistency_flag"`
	ContextNote           string `yaml:"context_note"`
	RetrievedAt           string `yaml:"retrieved_at"`
}

// AppliesToStatement reports whether this evidence is attached to
// statementID as either its primary or one of its secondary tags.
func (e Evidence) AppliesToStatement(statementID string) bool {
	if e.StatementID == statementID {
		return true
	}
	for _, id := range e.SecondaryStatementIDs {
		if id == statementID {
			return true
		}
	}
	return false
}

// IsPartyEvidence reports whether this evidence documents a specific
// party's own action or position, as opposed to institutional/state/actor
// context that doesn't belong to any single party.
func (e Evidence) IsPartyEvidence() bool {
	return e.PartyID != ""
}

// Program is one party's 2025-2029 programme points: what the party says it
// will do, quoted from its own programme document. This is tier C evidence —
// a promise, not a vote. The programme page shows these next to the party's
// coded positions on the statements, so a reader can see, in one place,
// what the party promised and (on the statement pages) what it actually
// did. A programme point never sets a position on its own: that needs a
// vote or a merknad. Added 2026-10-03 with the autopilot, which is what
// fills it.
type Program struct {
	ID          string         `yaml:"id"` // party_id
	SourceName  string         `yaml:"source_name"`
	SourceURL   string         `yaml:"source_url"`
	RetrievedAt string         `yaml:"retrieved_at"`
	Points      []ProgramPoint `yaml:"points"`
}

// ProgramPoint is one verbatim promise from a programme, with where in the
// document it sits so a reader can find it. Quote is the party's own words,
// never a paraphrase; an empty Points list for a dimension means the party's
// programme says nothing on it, which is itself an honest result.
type ProgramPoint struct {
	Dimension string `yaml:"dimension"`
	Topic     string `yaml:"topic"`
	Quote     string `yaml:"quote"`
	Where     string `yaml:"where"`
	// Statements lists the statement ids this programme point bears on.
	// Plural, because one promise often answers more than one question:
	// a point on state ownership of nuclear power bears on both "statlig
	// kontroll over kjernekraft" and "kommersiell kjernekraft nå". It is
	// optional and most points carry none — that is the honest default,
	// not a gap to fill. 64 links exist across 316 points, 59 of them in
	// the domain of a point with no link at all; see
	// docs/programme-statement-join-analysis.md for why a mechanical rule
	// cannot produce more, and why inventing them would be worse than an
	// empty field. A link is set by hand, with a stated reason, and the
	// loader refuses one whose statement sits in another dimension.
	Statements []string `yaml:"statements,omitempty"`
}

// Budget is one statsbudsjett, collected and analysed as a whole: the
// headline figures from the government's own documents, the per-department
// totals, and the findings that bear on the statements. It is the page that
// answers "what did the state actually decide about the money", so a reader
// can hold the budget next to the promises (programs) and the votes
// (evidence). Added 2026-10-12; a finding with an empty StatementID is a
// real result -- it means no statement covers it yet, named rather than
// forced into the nearest one.
type Budget struct {
	ID          string          `yaml:"id"`
	Year        int             `yaml:"year"`
	Title       string          `yaml:"title"`
	Presented   string          `yaml:"presented"`
	SourceName  string          `yaml:"source_name"`
	SourceURL   string          `yaml:"source_url"`
	RetrievedAt string          `yaml:"retrieved_at"`
	Summary     string          `yaml:"summary"`
	KeyFigures  []BudgetFigure  `yaml:"key_figures"`
	Departments []BudgetDept    `yaml:"departments"`
	Findings    []BudgetFinding `yaml:"findings"`
}

// BudgetFigure is one headline number with the caveat that must travel with
// it (the definition, or why it is not comparable to the line above).
type BudgetFigure struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
	Note  string `yaml:"note"`
}

// BudgetDept is one department's utgifter in thousands of kroner, with the
// year-on-year change in percent. Amount is int64 (full kroner precision is
// not meaningful here, but the Gul bok table is in thousands and losing the
// unit silently would be worse).
type BudgetDept struct {
	Name      string  `yaml:"name"`
	Amount    int64   `yaml:"amount"`
	ChangePct float64 `yaml:"change_pct"`
}

// BudgetFinding is one thing the budget decided, tied where possible to the
// statement it bears on. Quote is verbatim from SourceURL when present.
type BudgetFinding struct {
	Topic       string `yaml:"topic"`
	StatementID string `yaml:"statement_id,omitempty"`
	Quote       string `yaml:"quote,omitempty"`
	Note        string `yaml:"note"`
	SourceURL   string `yaml:"source_url"`
	// Observed holds the same post at each stage it exists in, when the
	// finding turns on a change. Optional: a finding with no arithmetic
	// leaves it empty. Added 2026-10-13 after the bredbånd post was found
	// comparing the government's 2026 PROPOSAL (34,1 mill. kroner) against
	// its 2027 proposal, while the Stortinget's SALDERT 2026 was 159,11
	// mill. Both comparisons are true and they answer different questions,
	// so the stage is now a field and never an assumption.
	Observed []BudgetObservation `yaml:"observed,omitempty"`
}

// BudgetObservation is one amount for one budget post, named by the year it
// belongs to and the stage that number is: what the government proposed,
// what the Stortinget settled, or what the accounts show. Validated by the
// loader, so an unknown stage fails the build rather than rendering as an
// unlabelled number a reader cannot place.
type BudgetObservation struct {
	FiscalYear int    `yaml:"fiscal_year"`
	Stage      string `yaml:"stage"`
	Amount     int64  `yaml:"amount"` // kroner
}

// Correction is one entry in the public corrections log (/om). It records
// a mistake this project made, in the project's own words, so a reader can
// check the claim without leaving the site.
//
// This repository is published with a single initial commit, so its git
// history is not a record anyone can consult. The record is these entries
// plus the build log in docs/roadmap.md (see the note on contentFS in
// loader.go). The facts here are taken from that build log's own
// day-by-day entries — hand-written, never invented.
//
// Kind separates the two honest cases. "rettelse" (correction) means
// something a reader could SEE on the site was wrong and changed.
// "nesten" (near miss) means something was wrong before it reached the
// site and was caught. Both belong in the log: a record that only lists
// the mistakes that escaped would be a curated record, not a record.
type Correction struct {
	ID    string `yaml:"id"`
	Date  string `yaml:"date"` // YYYY-MM-DD, the day the correction was made
	Kind  string `yaml:"kind"` // rettelse or nesten
	Title string `yaml:"title"`
	// Statements lists the statement numbers this correction touches, as
	// the reader knows them ("07", "44"). Numbers, not internal ids: the
	// log records history, and utsagn 43 no longer exists as an id even
	// though the correction that removed it is part of that history. The
	// loader therefore validates the SHAPE of each entry, never that the
	// statement still exists today.
	Statements []string `yaml:"statements,omitempty"`
	// What changed, why it was wrong, and what it meant for the result.
	// Three separate fields because collapsing them lets "we fixed it"
	// stand in for "here is what a reader saw that was false".
	What   string `yaml:"what"`
	Reason string `yaml:"reason"`
	Effect string `yaml:"effect"`
}

// validCorrectionKinds are the two kinds an entry may name.
var validCorrectionKinds = map[string]bool{
	"rettelse": true, // a published claim was wrong and changed
	"nesten":   true, // caught before it reached the site
}

// ValidCorrectionKind reports whether kind is one the loader accepts.
func ValidCorrectionKind(kind string) bool { return validCorrectionKinds[kind] }

// validBudgetStages are the stages an observation may name. Deliberately
// narrow: adding one means deciding what it compares against.
var validBudgetStages = map[string]bool{
	"forslag":  true, // regjeringens forslag, lagt fram
	"saldert":  true, // Stortingets vedtatte budsjett
	"regnskap": true, // faktisk regnskap
}

// ValidBudgetStage reports whether stage is one the loader accepts.
func ValidBudgetStage(stage string) bool { return validBudgetStages[stage] }
