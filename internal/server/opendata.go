package server

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// Open data (2026-10-23). Until now the dataset was reachable only by
// reading Go or cloning the repo, which is a strange place to keep the one
// thing the project asks a reader to trust. The pages already show every
// claim with its source; these files are the same data in a shape a
// spreadsheet or a script can take, so a reader can check the whole set
// instead of one card at a time.
//
// Generated from data.Store on every request, never written out at build
// time: a checked-in export is a second copy of the dataset that can be
// stale while still looking authoritative, which is the exact failure this
// project exists to avoid.
//
// The one thing that must not drift is the column list. evidenceColumns
// below produces the CSV header, the CSV rows, and the table on /data, and a
// test asserts it matches the JSON keys in this order. Adding a field to the
// export without describing it therefore fails, rather than shipping a file
// nobody can read.

// openDataSchemaVersion is bumped when a field is renamed or removed.
// Adding a field does not need a bump: every field is named in the JSON and
// in the CSV header, so a reader that selects by name is unaffected.
const openDataSchemaVersion = 1

// licenceNote is what the site says elsewhere about the licence, in the
// place a data consumer looks for it.
const licenceNote = "AGPL-3.0"

// envelope is the shape of both JSON exports. Items, not a bare array, so
// there is somewhere to say which snapshot the file came from; a consumer
// that stores the file can tell when it is out of date without diffing.
type envelope[T any] struct {
	SchemaVersion int    `json:"schema_version"`
	SnapshotDate  string `json:"snapshot_date"`
	Dataset       string `json:"dataset"`
	Licence       string `json:"licence"`
	Count         int    `json:"count"`
	Items         []T    `json:"items"`
}

// statementExport is a statement as published. URL is included because the
// statement's own page is where its evidence lives; a consumer holding only
// this file can point a reader at the page instead of at a bare number.
type statementExport struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	// Dimension is the dimension's Norwegian name, which is also the key
	// the loader uses: "Demokratisk åpenhet". There is no separate slug.
	Dimension       string        `json:"dimension"`
	Text            string        `json:"text"`
	ShortName       string        `json:"short_name"`
	PrimaryQuestion string        `json:"primary_question"`
	CaseTag         string        `json:"case_tag"`
	PracticalStakes *stakesExport `json:"practical_stakes,omitempty"`
	URL             string        `json:"url"`
}

type stakesExport struct {
	Enig  string `json:"enig"`
	Uenig string `json:"uenig"`
}

// evidenceExport is one claim about one party position, with its source.
//
// Position is a pointer on purpose: a party position is -2..2, and "not
// coded" is a third state, not zero. In JSON it marshals as null; in the CSV
// it is an empty cell, and the /data page says so, because a spreadsheet
// that reads "" as 0 would invent a position the site does not claim.
type evidenceExport struct {
	EvidenceID            string   `json:"evidence_id"`
	StatementID           string   `json:"statement_id"`
	SecondaryStatementIDs []string `json:"secondary_statement_ids"`
	PartyID               string   `json:"party_id"`
	Actor                 string   `json:"actor"`
	Date                  string   `json:"date"`
	EvidenceType          string   `json:"evidence_type"`
	EvidenceClass         string   `json:"evidence_class"`
	EvidensstyrkeCap      string   `json:"evidensstyrke_cap"`
	ScanSummary           string   `json:"scan_summary"`
	ClaimAction           string   `json:"claim_action"`
	VerbatimExcerpt       string   `json:"verbatim_excerpt"`
	Status                string   `json:"status"`
	Position              *int     `json:"position"`
	SourceTier            string   `json:"source_tier"`
	SourceName            string   `json:"source_name"`
	SourceURL             string   `json:"source_url"`
	ClaimConfidence       string   `json:"claim_confidence"`
	AttributionConfidence string   `json:"attribution_confidence"`
	ConsistencyFlag       string   `json:"consistency_flag"`
	ContextNote           string   `json:"context_note"`
	RetrievedAt           string   `json:"retrieved_at"`
}

func statementExportOf(s data.Statement) statementExport {
	out := statementExport{
		ID:              s.ID,
		Number:          s.Number,
		Dimension:       s.Dimension,
		Text:            s.Text,
		ShortName:       s.ShortName,
		PrimaryQuestion: s.PrimaryQuestion,
		CaseTag:         s.CaseTag,
		URL:             siteBase + "/utsagn/" + strconv.Itoa(s.Number),
	}
	if s.PracticalStakes != nil {
		out.PracticalStakes = &stakesExport{Enig: s.PracticalStakes.Enig, Uenig: s.PracticalStakes.Uenig}
	}
	return out
}

func evidenceExportOf(e data.Evidence) evidenceExport {
	return evidenceExport{
		EvidenceID:            e.ID,
		StatementID:           e.StatementID,
		SecondaryStatementIDs: e.SecondaryStatementIDs,
		PartyID:               e.PartyID,
		Actor:                 e.Actor,
		Date:                  e.Date,
		EvidenceType:          e.EvidenceType,
		EvidenceClass:         e.EvidenceClass,
		EvidensstyrkeCap:      e.EvidensstyrkeCap,
		ScanSummary:           e.ScanSummary,
		ClaimAction:           e.ClaimAction,
		VerbatimExcerpt:       e.VerbatimExcerpt,
		Status:                e.Status,
		Position:              e.Position,
		SourceTier:            e.SourceTier,
		SourceName:            e.SourceName,
		SourceURL:             e.SourceURL,
		ClaimConfidence:       e.ClaimConfidence,
		AttributionConfidence: e.AttributionConfidence,
		ConsistencyFlag:       e.ConsistencyFlag,
		ContextNote:           e.ContextNote,
		RetrievedAt:           e.RetrievedAt,
	}
}

// evidenceColumn is one column of the CSV, and one row of the table on
// /data. Value renders the cell, so the header and the cells are produced by
// the same list and cannot fall out of step.
type evidenceColumn struct {
	Name  string
	Doc   string
	Value func(evidenceExport) string
}

// evidenceColumns is the whole published shape of an evidence item, in
// order. The names match the JSON keys exactly; TestEvidenceColumnsMatchJSON
// holds that.
var evidenceColumns = []evidenceColumn{
	{"evidence_id", "Unik id for påstanden. Samme som filnavnet i repoet.", func(e evidenceExport) string { return e.EvidenceID }},
	{"statement_id", "Utsagnet påstanden gjelder, som id (utsagn-01).", func(e evidenceExport) string { return e.StatementID }},
	{"secondary_statement_ids", "Andre utsagn kilden også er lagt til grunn for, skilt med ;. Tomt = ingen.", func(e evidenceExport) string { return strings.Join(e.SecondaryStatementIDs, ";") }},
	{"party_id", "Partiet posisjonen gjelder (ap, frp, h, ...). Tomt = institusjonell kilde, ikke et partisyn.", func(e evidenceExport) string { return e.PartyID }},
	{"actor", "Hvem som handlet eller uttalte seg.", func(e evidenceExport) string { return e.Actor }},
	{"date", "Dato for handlingen eller uttalelsen, ikke for nedlastingen (ISO 8601).", func(e evidenceExport) string { return e.Date }},
	{"evidence_type", "Dokumenttypen: votering, innstilling, representantforslag, budsjettvedtak og så videre.", func(e evidenceExport) string { return e.EvidenceType }},
	{"evidence_class", "Hva slags dokument dette er for partiet: eget forslag, egen begrunnelse, stemmegivning.", func(e evidenceExport) string { return e.EvidenceClass }},
	{"evidensstyrke_cap", "Eventuelt tak på hvor sterkt kilden kan telle, med begrunnelse. Tomt = ikke satt.", func(e evidenceExport) string { return e.EvidensstyrkeCap }},
	{"scan_summary", "Én nøktern setning om hva som skjedde. Dette er teksten som vises i kortet.", func(e evidenceExport) string { return e.ScanSummary }},
	{"claim_action", "Påstanden om hva partiet gjorde eller mente, som den står på siden.", func(e evidenceExport) string { return e.ClaimAction }},
	{"verbatim_excerpt", "Ordrett sitat fra kilden. Tomt = det er ikke lagret noe sitat for denne påstanden.", func(e evidenceExport) string { return e.VerbatimExcerpt }},
	{"status", "Hva saken er: fremmet, vedtatt, forkastet, uttalt og så videre.", func(e evidenceExport) string { return e.Status }},
	{"position", "Partiets posisjon -2 til 2. Tomt = ikke kodet, som er en egen tilstand og ikke det samme som 0.", func(e evidenceExport) string {
		if e.Position == nil {
			return ""
		}
		return strconv.Itoa(*e.Position)
	}},
	{"source_tier", "Kildegrad A til E. A = primær og autoritativ, E = OSINT-lead som ennå ikke er verifisert.", func(e evidenceExport) string { return e.SourceTier }},
	{"source_name", "Kilden slik den vises på siden, for eksempel Innst. 135 S (2025–2026).", func(e evidenceExport) string { return e.SourceName }},
	{"source_url", "Lenken til kilden. Prosjektet lagrer ikke en kopi av dokumentet, så en flyttet side gjør påstanden usjekkbar.", func(e evidenceExport) string { return e.SourceURL }},
	{"claim_confidence", "Hvor sikkert påstanden er gjenfunnet i kilden.", func(e evidenceExport) string { return e.ClaimConfidence }},
	{"attribution_confidence", "Hvor sikkert kilden knytter handlingen til dette partiet. Tomt = ikke vurdert.", func(e evidenceExport) string { return e.AttributionConfidence }},
	{"consistency_flag", "Satt når to kilder om samme sak ikke stemmer overens. Tomt = ingen konflikt registrert.", func(e evidenceExport) string { return e.ConsistencyFlag }},
	{"context_note", "Merknad om hva kilden ikke sier, der det er lett å lese for mye inn i den.", func(e evidenceExport) string { return e.ContextNote }},
	{"retrieved_at", "Datoen kilden ble hentet og lest (ISO 8601).", func(e evidenceExport) string { return e.RetrievedAt }},
}

// writeJSON writes v as an indented JSON body. Indented rather than compact
// because the first reader of this file is a person who opened the URL.
func writeJSON(w http.ResponseWriter, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// The dataset changes only on deploy, so a short shared cache is safe
	// and keeps a script from pulling it on every run.
	w.Header().Set("Cache-Control", "public, max-age=600")
	_, _ = w.Write(append(b, '\n'))
}

func serveStatementsJSON(w http.ResponseWriter, ds *data.Dataset) {
	items := make([]statementExport, 0, len(ds.Statements))
	for _, s := range ds.Statements {
		items = append(items, statementExportOf(s))
	}
	writeJSON(w, envelope[statementExport]{
		SchemaVersion: openDataSchemaVersion,
		SnapshotDate:  SnapshotDate,
		Dataset:       "utsagn",
		Licence:       licenceNote,
		Count:         len(items),
		Items:         items,
	})
}

func serveEvidenceJSON(w http.ResponseWriter, ds *data.Dataset) {
	items := make([]evidenceExport, 0, len(ds.Evidence))
	for _, e := range ds.Evidence {
		items = append(items, evidenceExportOf(e))
	}
	writeJSON(w, envelope[evidenceExport]{
		SchemaVersion: openDataSchemaVersion,
		SnapshotDate:  SnapshotDate,
		Dataset:       "evidens",
		Licence:       licenceNote,
		Count:         len(items),
		Items:         items,
	})
}

// serveEvidenceCSV writes the same rows as serveEvidenceJSON, flat.
//
// No byte-order mark. A BOM makes Excel show æøå correctly on a double
// click, but it also turns the first header into "\ufeffevidence_id", which
// breaks a reader's join in Python or R with no error worth the name. A
// machine-correct file plus one line of instructions on /data is the better
// trade: the instructions can be read, a silent parse failure cannot.
func serveEvidenceCSV(w http.ResponseWriter, ds *data.Dataset) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="ektevalgomat-evidens.csv"`)
	w.Header().Set("Cache-Control", "public, max-age=600")

	cw := csv.NewWriter(w)
	header := make([]string, len(evidenceColumns))
	for i, c := range evidenceColumns {
		header[i] = c.Name
	}
	if err := cw.Write(header); err != nil {
		return
	}
	for _, e := range ds.Evidence {
		row := evidenceExportOf(e)
		cells := make([]string, len(evidenceColumns))
		for i, c := range evidenceColumns {
			cells[i] = c.Value(row)
		}
		if err := cw.Write(cells); err != nil {
			return
		}
	}
	cw.Flush()
}
