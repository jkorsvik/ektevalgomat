package server

import (
	"encoding/json"
	"sort"

	"github.com/jkorsvik/ektevalgomat/internal/data"
	"github.com/jkorsvik/ektevalgomat/internal/scoring"
)

// SnapshotDate is the cutoff the pages advertise ("Øyeblikksbilde:
// hendelser til og med <date>"). It was 2026-09-23 for the first version;
// it advanced to 2026-10-12 when the 2027 statsbudsjett (presented
// 2026-10-07) and the party programmes were added — an event dated after
// the cutoff would make the badge false, and TestNoEvidenceAfterSnapshotDate
// exists to enforce exactly that. Advancing the cutoff is the sanctioned way
// to add a newer observation; history is added, never rewritten in place.
const SnapshotDate = "2026-10-12"

// quizPayload is the full dataset the client-side quiz needs, serialized
// once at startup and embedded directly into index.html — no separate
// fetch, no API surface to version, no framework. The user's own answers
// never leave the browser (see web/static/js/quiz.js): only this
// read-only payload crosses the server boundary.
type quizPayload struct {
	SnapshotDate string          `json:"snapshot_date"`
	Statements   []quizStatement `json:"statements"`
	Parties      []quizParty     `json:"parties"`
	// Cells[statementID][partyID] — nil PartyPosition means "not coded
	// yet"; the client must render that as "ikke vurdert", never as 0.
	Cells map[string]map[string]scoring.Cell `json:"cells"`
}

type quizStatement struct {
	ID              string `json:"id"`
	Number          int    `json:"number"`
	Text            string `json:"text"`
	ShortName       string `json:"short_name"`
	Dimension       string `json:"dimension"`
	PrimaryQuestion string `json:"primary_question"`
	// PracticalStakesEnig/Uenig are flattened from data.Statement's
	// optional PracticalStakes (nil for most statements) so quiz.js
	// never has to null-check a nested object — omitted from the JSON
	// entirely when there's nothing to show.
	PracticalStakesEnig  string `json:"practical_stakes_enig,omitempty"`
	PracticalStakesUenig string `json:"practical_stakes_uenig,omitempty"`
}

type quizParty struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Short string `json:"short"`
}

func buildQuizPayload(ds *data.Dataset) quizPayload {
	p := quizPayload{
		SnapshotDate: SnapshotDate,
		Cells:        make(map[string]map[string]scoring.Cell, len(ds.Statements)),
	}
	for _, s := range ds.Statements {
		qs := quizStatement{
			ID: s.ID, Number: s.Number, Text: s.Text, ShortName: s.ShortName,
			Dimension: s.Dimension, PrimaryQuestion: s.PrimaryQuestion,
		}
		if s.PracticalStakes != nil {
			qs.PracticalStakesEnig = s.PracticalStakes.Enig
			qs.PracticalStakesUenig = s.PracticalStakes.Uenig
		}
		p.Statements = append(p.Statements, qs)
		cells := make(map[string]scoring.Cell, len(ds.Parties))
		for _, party := range ds.Parties {
			cells[party.ID] = scoring.ComputeCell(ds, s.ID, party.ID)
		}
		p.Cells[s.ID] = cells
	}
	for _, party := range ds.Parties {
		p.Parties = append(p.Parties, quizParty{ID: party.ID, Name: party.Name, Short: party.Short})
	}
	return p
}

func (p quizPayload) MarshalIndent() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// statementPage is the server-rendered evidence-chain view for one
// statement: /utsagn/{number}. This is the memo's core deliverable — "en
// reproduserbar politisk hendelseslogg der enhver bruker kan åpne
// utsagnet og se" the full chain — rendered directly, not hidden behind
// client-side JS, so it's linkable and readable with no script at all.
type statementPage struct {
	Meta                  pageMeta
	Statement             data.Statement
	SnapshotDate          string
	TotalParties          int
	InstitutionalEvidence []evidenceView
	// DocumentedRows and UndocumentedRows split PartyRows in two so the
	// template can show documented parties up front and collapse the
	// undocumented ones behind a single disclosure, instead of nine
	// equally-weighted cards where most of them just repeat "ikke
	// vurdert" three times (2026-09-23 design review).
	DocumentedRows   []partyEvidenceRow
	UndocumentedRows []partyEvidenceRow
	// Chain is the evidence chain for this statement: which steps of
	// claim-to-outcome the site actually has a source for. Every step is
	// present, including the empty ones, because an empty step is a
	// finding — see chain.go.
	Chain       []chainStage
	ChainAny    bool
	ChainAbsent int
	// Theme is what each party says in this statement's own theme (its
	// dimension), with the points that name this statement marked. It
	// exists because the precise join is sparse on purpose: 59 of 316
	// points name a statement. The honest answer to "hva sier partiet
	// her" is therefore the theme, shown as the theme — not a promise
	// claim the point does not support. Empty for a statement whose
	// dimension has no programme points at all.
	// Bearing holds the linked points on their own: a point that names
	// THIS statement is the only one that may make a promise claim about
	// it, so it is shown open, above the theme. Theme holds everything
	// else — the same theme, no link — behind a disclosure, so the page
	// leads with the specific and keeps the general available.
	Bearing    []bearingPointView
	Theme      []themePartyView
	ThemeCount int
}

// bearingPointView is a programme point that names this exact statement.
type bearingPointView struct {
	Party data.Party
	Topic string
	Quote string
	Where string
}

// themePointView is one programme point that shares the statement's theme
// without naming it.
type themePointView struct {
	Topic string
	Quote string
	Where string
}

// themePartyView is one party's points in the statement's dimension.
type themePartyView struct {
	Party  data.Party
	Points []themePointView
}

// AnyPartyDocumented reports whether at least one party has any
// evidence at all (coded or not). When this is false, the template shows
// one compact "0 av N partier dokumentert" card instead of nine
// near-identical empty party cards.
func (p statementPage) AnyPartyDocumented() bool {
	return len(p.DocumentedRows) > 0
}

// CodedCount is how many parties have a coded position. It can be lower
// than len(DocumentedRows): a party can have a real source that still
// does not code a stance on this exact statement (utsagn-45, where the
// Innst. 464 S context is tier A but position: null). The page headline
// counts this, not rows with evidence, so "9 av 9" never appears next to
// two "Ikke vurdert" rows.
func (p statementPage) CodedCount() int {
	n := 0
	for _, r := range p.DocumentedRows {
		if r.Cell.PartyPosition != nil {
			n++
		}
	}
	return n
}

// SourcedOnlyCount is parties with evidence but no coded position.
func (p statementPage) SourcedOnlyCount() int {
	return len(p.DocumentedRows) - p.CodedCount()
}

type evidenceView struct {
	data.Evidence
	SourceTierLabel string
}

type partyEvidenceRow struct {
	Party    data.Party
	Cell     scoring.Cell
	Evidence []evidenceView
}

func buildStatementPage(ds *data.Dataset, statement data.Statement) statementPage {
	all := ds.EvidenceForStatement(statement.ID)

	page := statementPage{Meta: statementMeta(statement), Statement: statement, SnapshotDate: SnapshotDate, TotalParties: len(ds.Parties)}
	page.Chain = buildChain(ds, statement.ID)
	page.ChainAny = chainHasAny(page.Chain)
	for _, s := range page.Chain {
		if s.Count == 0 {
			page.ChainAbsent++
		}
	}

	for _, e := range all {
		if !e.IsPartyEvidence() {
			page.InstitutionalEvidence = append(page.InstitutionalEvidence, evidenceView{
				Evidence: e, SourceTierLabel: sourceTierLabel(e.SourceTier),
			})
		}
	}

	parties := make([]data.Party, len(ds.Parties))
	copy(parties, ds.Parties)
	sort.Slice(parties, func(i, j int) bool { return parties[i].Name < parties[j].Name })

	for _, party := range parties {
		var items []evidenceView
		for _, e := range all {
			if e.PartyID == party.ID {
				items = append(items, evidenceView{Evidence: e, SourceTierLabel: sourceTierLabel(e.SourceTier)})
			}
		}
		cell := scoring.ComputeCell(ds, statement.ID, party.ID)
		row := partyEvidenceRow{Party: party, Cell: cell, Evidence: items}
		if cell.HasEvidence() {
			page.DocumentedRows = append(page.DocumentedRows, row)
		} else {
			page.UndocumentedRows = append(page.UndocumentedRows, row)
		}
	}

	// Bearing comes from the links themselves, NOT from the dimension
	// scan: seven links cross dimensions on purpose, and building this
	// list from the theme would silently drop exactly those (the
	// dimension scan cannot see a point that lives in another one). Eye
	// on it: a test that only checks utsagn-01 cannot catch this, because
	// there the two sets coincide.
	bearingKeys := map[string]bool{}
	for _, ref := range ds.ProgramPointsForStatement(statement.ID) {
		party, ok := ds.Party(ref.PartyID)
		if !ok {
			continue
		}
		page.Bearing = append(page.Bearing, bearingPointView{
			Party: party, Topic: ref.Point.Topic, Quote: ref.Point.Quote, Where: ref.Point.Where,
		})
		bearingKeys[ref.PartyID+"\x00"+ref.Point.Topic] = true
	}

	for _, party := range parties {
		pts := ds.ProgramPointsForDimension(party.ID, statement.Dimension)
		if len(pts) == 0 {
			continue
		}
		row := themePartyView{Party: party}
		for _, pt := range pts {
			if bearingKeys[party.ID+"\x00"+pt.Topic] {
				continue // already shown open above; not repeated below
			}
			row.Points = append(row.Points, themePointView{
				Topic: pt.Topic, Quote: pt.Quote, Where: pt.Where,
			})
			page.ThemeCount++
		}
		if len(row.Points) > 0 {
			page.Theme = append(page.Theme, row)
		}
	}

	return page
}

func sourceTierLabel(tier string) string {
	labels := map[string]string{
		"A": "A — primær og autoritativ (lov, offisielt voteringsresultat, regnskap, registerdata)",
		"B": "B — primær institusjonell analyse (SSB, FFI, NSM, PST, NVE, Riksrevisjonen)",
		"C": "C — aktørens egen publisering (program, pressemelding)",
		"D": "D — verifisert sekundærkilde (etablert presse/forskning)",
		"E": "E — OSINT-lead, ennå ikke verifisert",
	}
	if l, ok := labels[tier]; ok {
		return l
	}
	return tier
}
