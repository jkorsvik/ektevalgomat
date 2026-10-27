// Package server wires ektevalgomat's data, scoring, and templates into
// an http.Handler. Stdlib net/http only — no router dependency, no ORM,
// no session store. The only state is the read-only Dataset loaded once
// at startup (internal/data); nothing here is mutable at runtime.
package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jkorsvik/ektevalgomat/internal/data"
	"github.com/jkorsvik/ektevalgomat/internal/scoring"
	"github.com/jkorsvik/ektevalgomat/web"
)

func currentYear() int { return time.Now().Year() }

// assetVersion cache-busts static assets on every content change. Without
// it, Cloudflare's edge can serve a stale-cached stylesheet against fresh
// HTML.
var assetVersion = computeAssetVersion()

func computeAssetVersion() string {
	h := sha256.New()
	err := fs.WalkDir(web.Static, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		f, err := web.Static.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(h, f)
		return err
	})
	if err != nil {
		panic("computeAssetVersion: " + err.Error())
	}
	return hex.EncodeToString(h.Sum(nil))[:10]
}

func withVersion(p string) string { return p + "?v=" + assetVersion }

// cacheStatic lets browsers keep an asset for a year when its URL can never
// serve different bytes: anything requested with ?v=<content hash>, and the
// font files (a new font must get a new file name). Directory listings end
// in "/" and can change, so they never qualify. Error responses lose the
// header anyway: FileServer deletes Cache-Control on errors (Go 1.23+,
// pinned by TestStaticCacheHeaders). Everything else keeps the default
// revalidation.
func cacheStatic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isFile := !strings.HasSuffix(r.URL.Path, "/")
		if isFile && (r.URL.Query().Has("v") || strings.HasPrefix(r.URL.Path, "/static/fonts/")) {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		h.ServeHTTP(w, r)
	})
}

// addendumDimension marks statements that exist outside the source
// memo's 1-42 numbering (see docs/roadmap.md for why they exist and how
// they got there) — they must never be counted or displayed as if they
// were one of the memo's actual 9 dimensions.
const addendumDimension = "Tillegg (utenfor de 42)"

// plannedDimensionTotals is the memo's own dimension → PLANNED-statement
// count (docs/roadmap.md's "Correction 2", reproduced from
// research-memo.md) plus, for the two post-memo dimensions the owner
// approved, the count committed when each was added. It records how many
// statements the original vision PLANNED for each dimension.
//
// It is a PLAN. It is never a count of statements that exist, it must
// never be a dimension's Total in a "built / total" row, and it must
// never be summed into a claim about the live dataset. It went stale the
// moment the dataset grew past it: "Norsk handlefrihet" planned 6 but has
// 7, "Universell grunnmur" 6 → 7, "Robusthet / adaptabilitet" 4 → 5, and
// "Finansiering av fellesskapet" 5 → 7. Used as a denominator, a
// planned-low number shows a dimension at more than 100% built — the
// exact overclaim this project's correction history exists to prevent.
// Existing counts come from ds.Statements (dimensionBuiltCounts), never
// from here; the memo plan is shown only where it is labelled as a plan
// (plannedMemoTotal).
var plannedDimensionTotals = map[string]int{
	"Demokratisk åpenhet":                             5,
	"Norsk handlefrihet":                              6,
	"Forvaltning av fellesskapets ressurser / energi": 7,
	"Teknologisk kapasitet":                           3,
	"Innovasjon / omstillingsevne":                    3,
	"Universell grunnmur":                             6,
	"Barns muligheter":                                6,
	"Robusthet / adaptabilitet":                       4,
	"Generasjonsansvar":                               2,
	"Finansiering av fellesskapet":                    5,
	"Tilgang, integrering og rettsstat":               6,
}

// dimensionBuiltCounts derives, from the validated dataset, how many
// statements exist in each dimension — the addendum excluded. This is the
// single source of truth for a dimension's Total: a hand-maintained count
// drifts the moment content is added, which is exactly how the front page
// once showed a dimension at more than 100% built.
func dimensionBuiltCounts(ds *data.Dataset) map[string]int {
	counts := map[string]int{}
	for _, s := range ds.Statements {
		if s.Dimension == addendumDimension {
			continue
		}
		counts[s.Dimension]++
	}
	return counts
}

// memoDimensions is the original research memo's nine dimensions, in the
// memo's own order (docs/research-memo.md's "Metode og undersøkelsesroadmap"
// table).
var memoDimensions = []string{
	"Demokratisk åpenhet",
	"Norsk handlefrihet",
	"Forvaltning av fellesskapets ressurser / energi",
	"Teknologisk kapasitet",
	"Innovasjon / omstillingsevne",
	"Universell grunnmur",
	"Barns muligheter",
	"Robusthet / adaptabilitet",
	"Generasjonsansvar",
}

// postMemoDimensions are the two dimensions the owner approved beyond the
// memo's original nine — neither is in research-memo.md and neither
// extends any of its 42 planned statement slots. Listed in owner-assigned
// number order (10, 11), which is why "Finansiering av fellesskapet" (10)
// precedes "Tilgang, integrering og rettsstat" (11) even though it was
// coded into this repo later.
var postMemoDimensions = []string{
	"Finansiering av fellesskapet",
	"Tilgang, integrering og rettsstat",
}

// dimensionOrder is every dimension the site shows: the memo's nine
// (memoDimensions), then the owner-approved additions (postMemoDimensions).
// A fixed slice, not discovered from ds.Statements, so allDimensionsSummary
// can show a first-time visitor the whole vision, including a dimension
// with no statement built yet.
var dimensionOrder = append(append([]string{}, memoDimensions...), postMemoDimensions...)

// plannedMemoTotal is the number of statements the original memo planned
// across its nine dimensions (docs/roadmap.md's "Correction 2"). It is a
// PLAN, and is presented as one, never as a count of existing statements —
// see plannedDimensionTotals. Derived from the map rather than written
// down a second time, so it cannot drift from it.
func plannedMemoTotal() int {
	n := 0
	for _, d := range memoDimensions {
		n += plannedDimensionTotals[d]
	}
	return n
}

// dimensionQuestions is the memo's own "Primært spørsmål" per dimension
// (same table as dimensionOrder) for the first nine rows, quoted from
// the source memo rather than invented front-page marketing copy; plus
// each post-memo dimension's own stated question appended after it,
// quoted from the decision that created it (see docs/roadmap.md)
// rather than paraphrased.
var dimensionQuestions = map[string]string{
	"Demokratisk åpenhet":                             "Kan offentligheten se hvem som påvirker hvem, for hvem og med hvilke penger?",
	"Norsk handlefrihet":                              "Beholder Norge reell beslutnings- og operasjonsevne?",
	"Forvaltning av fellesskapets ressurser / energi": "Hvem kontrollerer ressursene, risikoen og verdiskapingen?",
	"Teknologisk kapasitet":                           "Kan kritiske digitale funksjoner og kompetanse opprettholdes under press?",
	"Innovasjon / omstillingsevne":                    "Bygger staten og markedet ny kapasitet eller bare kjøper ferdige løsninger?",
	"Universell grunnmur":                             "Fungerer samfunnets grunnleggende tjenester uansett bosted og inntekt?",
	"Barns muligheter":                                "I hvilken grad kompenserer samfunnet for forhold barn ikke har valgt?",
	"Robusthet / adaptabilitet":                       "Tåler systemet bortfall, sjokk og feilantakelser?",
	"Generasjonsansvar":                               "Synliggjøres konsekvensene 10-30 år frem?",
	"Finansiering av fellesskapet":                    "Hvordan betaler vi for fellesgodene, og hvem bærer kostnaden nå og om 30 år?",
	"Tilgang, integrering og rettsstat":               "Hvem får tilgang til fellesgodene, på hvilke vilkår, og hvordan håndhever samfunnet loven?",
}

// dimensionCoverage is one row of dimensionSummary's output: a dimension
// name, how many of its statements exist in the validated dataset (Built),
// and the total the row is measured against (Total). Both are derived from
// ds.Statements, so Built can never exceed Total and a dimension can never
// read as "more than 100% built".
type dimensionCoverage struct {
	Name  string
	Built int
	Total int
}

// dimensionSummary returns per-dimension coverage in first-appearance
// order (statement-number order). Each row's Total is the number of
// statements that actually exist in the dimension, derived from the
// dataset — not a hand-maintained plan — so Built equals Total and the
// row can never show more built than exist.
func dimensionSummary(ds *data.Dataset) []dimensionCoverage {
	order := []string{}
	seen := map[string]bool{}
	counts := dimensionBuiltCounts(ds)
	for _, s := range ds.Statements {
		if s.Dimension == addendumDimension {
			continue
		}
		if !seen[s.Dimension] {
			seen[s.Dimension] = true
			order = append(order, s.Dimension)
		}
	}
	rows := make([]dimensionCoverage, 0, len(order))
	for _, name := range order {
		rows = append(rows, dimensionCoverage{Name: name, Built: counts[name], Total: counts[name]})
	}
	return rows
}

// cellCoverage is what one statement's cells add up to: how many of the
// parties have a coded position, out of how many, which source tiers those
// positions rest on ("A", "A/C", ...), and whether every one of them is
// programme-only.
type cellCoverage struct {
	Coded, Total  int
	Tiers         string
	ProgrammeOnly bool
}

// hasStrongTier reports whether any evidence item that carries a coded
// position rests on an A or B source. An item with no position is context or
// an uncoded vote: it is listed on the page but it does not decide anything,
// so it must not make a programme-only cell look stronger (see
// scoring.strengthBasis, which draws the same line).
func hasStrongTier(items []data.Evidence) bool {
	for _, e := range items {
		if e.Position == nil {
			continue
		}
		if e.SourceTier == "A" || e.SourceTier == "B" {
			return true
		}
	}
	return false
}

// statementCoverage reports how the parties' positions on one statement are
// documented.
//
// /om prints these figures instead of a sentence typed once. The sentence it
// replaces said the positions on utsagn 45 came from programmes "for sju av
// ni partier. Rødt og Sp mangler fortsatt" — both were coded on 2026-10-13,
// and the page kept saying they were missing for twelve days. A number the
// page derives from the dataset cannot survive that way. ProgrammeOnly is
// here for the other half of the same problem: the sentence also said the
// positions "kommer fra partiprogrammene", which stops being true the day
// someone codes a vote on that statement, so the template picks its wording
// from this flag rather than asserting it.
func statementCoverage(ds *data.Dataset, number int) cellCoverage {
	cov := cellCoverage{Total: len(ds.Parties), ProgrammeOnly: true}
	var st *data.Statement
	for i := range ds.Statements {
		if ds.Statements[i].Number == number {
			st = &ds.Statements[i]
			break
		}
	}
	if st == nil {
		return cov
	}
	seen := map[string]bool{}
	for _, p := range ds.Parties {
		if scoring.ComputeCell(ds, st.ID, p.ID).PartyPosition == nil {
			continue
		}
		items := ds.EvidenceForCell(st.ID, p.ID)
		cov.Coded++
		if hasStrongTier(items) {
			cov.ProgrammeOnly = false
		}
		for _, e := range items {
			if e.Position != nil && e.SourceTier != "" {
				seen[e.SourceTier] = true
			}
		}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	cov.Tiers = strings.Join(keys, "/")
	// With no coded position there is nothing to call programme-only.
	cov.ProgrammeOnly = cov.ProgrammeOnly && cov.Coded > 0
	return cov
}

// programmeOnlyCells counts the statement×party cells that carry a coded
// position but rest on no A- or B-class source: the party's own programme
// (C) or a secondary account of it (D) is all there is. That is the one
// place where the site's headline — "vi sammenligner ikke svaret ditt med
// partiprogrammenes egne ord" — does not hold, so /metode states the number
// instead of leaving a reader to discover it. Derived, never typed.
func programmeOnlyCells(ds *data.Dataset) (cells, coded int) {
	for _, s := range ds.Statements {
		for _, p := range ds.Parties {
			if scoring.ComputeCell(ds, s.ID, p.ID).PartyPosition == nil {
				continue
			}
			coded++
			if !hasStrongTier(ds.EvidenceForCell(s.ID, p.ID)) {
				cells++
			}
		}
	}
	return cells, coded
}

// uncomparableStatements returns the statements no party has a coded
// position on, so the quiz can say so before the reader answers instead of
// after. Two statements (60, 61) are in the quiz with 0 av 9 dokumentert:
// the reader answers them, no party is compared, and the answer changes
// nothing. The result panel already names such statements after the fact
// (analyse.js, "telte ikke i noens resultat"), but a question that cannot
// count is worth one line before the reader spends an answer on it.
func uncomparableStatements(ds *data.Dataset) []data.Statement {
	var out []data.Statement
	for _, s := range ds.Statements {
		coded := false
		for _, p := range ds.Parties {
			if scoring.ComputeCell(ds, s.ID, p.ID).PartyPosition != nil {
				coded = true
				break
			}
		}
		if !coded {
			out = append(out, s)
		}
	}
	return out
}

// addendumCount returns how many loaded statements sit outside the
// memo's 1-42 numbering, so the UI can disclose them honestly instead
// of silently folding them into the memo's dimension count.
func addendumCount(ds *data.Dataset) int {
	n := 0
	for _, s := range ds.Statements {
		if s.Dimension == addendumDimension {
			n++
		}
	}
	return n
}

// coreStatementBuilt returns how many loaded statements sit inside the
// dimension structure — i.e. everything except the addendum dimension.
// Derived from the validated dataset, never from a hand-maintained map, so
// it can neither drift behind nor overshoot the content it counts. It is
// the figure the front page and /om show as the number of existing
// statements; the memo's original PLAN is a separate figure
// (plannedMemoTotal), always presented as a plan.
// repoURL is the public tracker a reader is sent to when they want to
// report a bad source reading or propose a statement. It is the same repo
// the footer already links; the difference is that these links arrive with
// the claim or statement already filled in.
const repoURL = "https://github.com/jkorsvik/ektevalgomat"

// issueURL builds a prefilled "new issue" link. Prefilling matters more
// than it looks: a report that starts empty becomes "one of the statements
// about defence, I think the FrP one is wrong", which cannot be acted on
// without a round trip. With the evidence id and its source in the body,
// the report is already specific.
//
// The templates are Markdown files, not issue forms, because only Markdown
// templates honour a prefilled body through the URL.
func issueURL(tmplName, title, body string) string {
	q := url.Values{}
	q.Set("template", tmplName)
	// Empty values are left out rather than sent as "title=": GitHub keeps
	// an empty title parameter, which leaves a blank required field where
	// the template's own placeholder would have been.
	if title != "" {
		q.Set("title", title)
	}
	if body != "" {
		q.Set("body", body)
	}
	return repoURL + "/issues/new?" + q.Encode()
}

// quoteCounts reports how many evidence claims carry an archived verbatim
// excerpt from their source, and how many carry none.
//
// Derived from the dataset on every render, never written down. /metode
// states the two numbers to a visitor as a promise about what they can
// check on the page, so a hand-typed pair would go stale the first time a
// claim was added — the same trap the roadmap records for planned
// statement counts ("it is a PLAN, never a count of statements that
// exist").
func quoteCounts(ds *data.Dataset) (total, withQuote int) {
	for _, e := range ds.Evidence {
		if e.VerbatimExcerpt != "" {
			withQuote++
		}
	}
	return len(ds.Evidence), withQuote
}

func coreStatementBuilt(ds *data.Dataset) int {
	return len(ds.Statements) - addendumCount(ds)
}

// dimLabel formats one dimensionCoverage row as "Navn (N av M utsagn)",
// where M is the number of statements that exist in the dimension (derived
// from the dataset), so the fraction can never exceed 100%. It is kept
// alongside the rows for a compact, text-only dimension list; the front
// page's "Kjerneverdiene" section now renders each row's own count in
// index.html.tmpl instead.
func dimLabel(d dimensionCoverage) string {
	return fmt.Sprintf("%s (%d av %d utsagn)", d.Name, d.Built, d.Total)
}

// allDimensionsSummary returns every dimension the site shows (the memo's
// nine, then the owner-approved additions), in dimensionOrder, with Built
// and Total both derived from the dataset — Total defaults to 0 for a
// dimension with no statement built yet, so it reads "0 av 0", never "0 av
// some planned number that does not exist". Distinct from dimensionSummary
// (which only lists dimensions ds.Statements has actually touched): this
// one is for showing a first-time visitor the whole vision on the front
// page, always every dimension, never silently just the ones already
// started.
func allDimensionsSummary(ds *data.Dataset) []dimensionCoverage {
	counts := dimensionBuiltCounts(ds)
	rows := make([]dimensionCoverage, 0, len(dimensionOrder))
	for _, name := range dimensionOrder {
		rows = append(rows, dimensionCoverage{Name: name, Built: counts[name], Total: counts[name]})
	}
	return rows
}

// dimQuestion looks up the memo's own primary question for a dimension
// row (dimensionQuestions), for display alongside allDimensionsSummary's
// list — quoted from the source memo, not invented copy.
func dimQuestion(d dimensionCoverage) string {
	return dimensionQuestions[d.Name]
}

// slugify turns a dimension name into an ASCII, lowercase,
// hyphen-separated token safe to use as an HTML id/URL fragment
// without percent-encoding: æ/ø/å are transliterated (ae/o/a, the
// common Norwegian URL-slug convention) rather than left as literal
// Unicode, because html/template percent-encodes non-ASCII bytes in an
// href="#..." attribute — "demokratisk-åpenhet" would render as
// "demokratisk-%C3%A5penhet" in the link but as the literal
// "demokratisk-åpenhet" id when quiz.js sets element.id directly,
// leaving correctness resting on browsers percent-decoding the
// fragment before an anchor lookup. Any other run of non-letter/
// non-digit characters becomes a single "-", trimmed at both ends.
func slugify(s string) string {
	replacements := map[rune]string{'æ': "ae", 'ø': "o", 'å': "a"}
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(s) {
		if rep, ok := replacements[r]; ok {
			b.WriteString(rep)
			lastDash = false
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// dimSlug is slugify applied to a dimensionCoverage row's Name — the
// anchor id the front page's "Kjerneverdiene" links jump to
// (quiz.js's renderQuiz computes the identical slug client-side for
// the quiz's own per-dimension group headings, so the two must stay in
// lockstep; see that file's slugify).
func dimSlug(d dimensionCoverage) string {
	return slugify(d.Name)
}

func joinStrings(sep string, items []string) string { return strings.Join(items, sep) }

// dateNO renders an ISO date from the content files ("2026-09-23") the way
// a Norwegian reader writes it ("23.09.2026"). The loader has already
// validated the format, so a parse failure here means a hand-edited value
// got past it; fall back to the raw string rather than printing a zero
// date, which would look like a real date.
func dateNO(iso string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	return t.Format("02.01.2006")
}

// evidenceClassBadge renders an evidence_class value as an icon+text
// label together, never as color alone (WCAG 1.4.1) — the icon is a
// plain geometric/punctuation glyph, not an emoji, so it renders
// consistently for screen readers and across platforms. Uppercase text
// so it reads as a label/tag, not a sentence, matching the tier badge
// next to it (2026-09-23 design review: "FAKTA / Primærkilde / dato",
// not "■ Fakta / dato — aktør / A"). See data.Evidence.EvidenceClass for
// what each class means.
func evidenceClassBadge(class string) string {
	labels := map[string]string{
		"fakta":                "■ FAKTA",
		"partiets_begrunnelse": "❝ PARTIETS BEGRUNNELSE",
		"analyse":              "◆ ANALYSE",
		"uklart":               "? UKLART",
	}
	if l, ok := labels[class]; ok {
		return l
	}
	return class
}

// sourceTierShort is the inline-badge form of a source tier — one or two
// words a general reader recognizes without learning the A–E ontology
// first. The full letter-graded explanation (sourceTierLabel in
// viewmodel.go) stays available as a title attribute / the footer
// legend, never as the only signal in the main scan row.
func sourceTierShort(tier string) string {
	labels := map[string]string{
		"A": "Primærkilde",
		"B": "Institusjonell analyse",
		"C": "Aktørens egen kilde",
		"D": "Sekundærkilde",
		"E": "OSINT (uverifisert)",
	}
	if l, ok := labels[tier]; ok {
		return l
	}
	return tier
}

// evidensstyrkeDots renders the scoring package's Evidensstyrke text
// label (already computed from count + source tier — see
// internal/scoring/scoring.go) as a three-dot strength scale. This is
// ALWAYS rendered alongside the existing text label, never instead of
// it — dots alone would be exactly the kind of single-channel signal
// the 2026-09-23 design brief says to avoid ("ikke bare farge").
func evidensstyrkeDots(label string) string {
	switch label {
	case "Sterk":
		return "●●●"
	case "Moderat":
		return "●●○"
	case "Svak":
		return "●○○"
	default:
		return "○○○"
	}
}

// New builds the full handler. Panics on template or dataset errors —
// build-time bugs, not runtime conditions, so failing fast at startup
// beats serving a half-broken evidence page.
func New(logger *slog.Logger) http.Handler {
	ds := data.Store
	funcs := template.FuncMap{
		"year":               currentYear,
		"asset":              withVersion,
		"positionLabel":      scoring.PositionLabel,
		"positionWord":       scoring.PositionWord,
		"positionText":       func(c scoring.Cell) string { return c.PositionText() },
		"positionDisagrees":  func(c scoring.Cell) bool { return c.PositionDisagrees() },
		"join":               joinStrings,
		"trimPrefix":         strings.TrimPrefix,
		"dimLabel":           dimLabel,
		"dimQuestion":        dimQuestion,
		"dimSlug":            dimSlug,
		"coreStatementBuilt": func() int { return coreStatementBuilt(ds) },
		"plannedMemoTotal":   plannedMemoTotal,
		"acquiescenceSpread": func() int { return scoring.MaxAcquiescenceSpread(data.Store) },
		"evidenceClassBadge": evidenceClassBadge,
		"evidensstyrkeDots":  evidensstyrkeDots,
		"sourceTierShort":    sourceTierShort,
		"issueURL":           issueURL,
		"siteBase":           func() string { return siteBase },
		"nok":                nok,
		"dateNO":             dateNO,
		"pct":                pct,
	}

	// Every page shares partials.tmpl ({{template "head"}}, {{template
	// "header"}}), so head assets, logo and nav have one source of truth.
	page := func(name string) *template.Template {
		return template.Must(template.New(name).Funcs(funcs).
			ParseFS(web.Templates, "templates/"+name, "templates/partials.tmpl"))
	}
	indexTmpl := page("index.html.tmpl")
	statementTmpl := page("statement.html.tmpl")
	metodeTmpl := page("metode.html.tmpl")
	dataTmpl := page("data.html.tmpl")
	omTmpl := page("om.html.tmpl")
	stemmerTmpl := page("stemmer.html.tmpl")
	kartTmpl := page("kart.html.tmpl")
	pengerTmpl := page("penger.html.tmpl")
	eierskapTmpl := page("eierskap.html.tmpl")
	maktTmpl := page("makt.html.tmpl")
	tidslinjeTmpl := page("tidslinje.html.tmpl")
	personvernTmpl := page("personvern.html.tmpl")
	resultatTmpl := page("resultat.html.tmpl")
	programTmpl := page("program.html.tmpl")
	programsTmpl := page("programs.html.tmpl")
	budgetTmpl := page("budget.html.tmpl")
	bidraTmpl := page("bidra.html.tmpl")

	payload := buildQuizPayload(ds)
	payloadJSON, err := payload.MarshalIndent()
	if err != nil {
		panic("server: marshaling quiz payload: " + err.Error())
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		sitemap(w, r, ds)
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.Handle("GET /static/", cacheStatic(http.FileServerFS(web.Static)))

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := struct {
			Meta             pageMeta
			SnapshotDate     string
			PayloadJSON      template.JS
			Statements       int
			StatementList    []data.Statement
			Parties          int
			DimensionRows    []dimensionCoverage
			AllDimensionRows []dimensionCoverage
			AddendumCount    int
			NoPosition       []data.Statement
		}{
			Meta: meta("Ekte valgomat · hvert svar har en kilde",
				"En valgomat der hvert partistandpunkt har en kilde. Hva partiene har lovet, foreslått og stemt, med lenke til dokumentet det står i.",
				"/"),
			SnapshotDate:     SnapshotDate,
			PayloadJSON:      template.JS(payloadJSON), //nolint:gosec // our own generated JSON, never user input
			Statements:       len(ds.Statements),
			StatementList:    ds.Statements,
			Parties:          len(ds.Parties),
			DimensionRows:    dimensionSummary(ds),
			AllDimensionRows: allDimensionsSummary(ds),
			AddendumCount:    addendumCount(ds),
			NoPosition:       uncomparableStatements(ds),
		}
		if err := indexTmpl.Execute(w, data); err != nil {
			logger.Error("render index", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /utsagn/{number}", func(w http.ResponseWriter, r *http.Request) {
		n, err := strconv.Atoi(r.PathValue("number"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		var found *data.Statement
		for _, s := range ds.Statements {
			if s.Number == n {
				st := s
				found = &st
				break
			}
		}
		if found == nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := statementTmpl.Execute(w, buildStatementPage(ds, *found)); err != nil {
			logger.Error("render statement", "err", err, "number", n)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	// Party programme pages (2026-10-03). /partier is every party's
	// programme points on one page; /parti/{id} is one party. They show
	// what a party says it will do (tier C, its own words); the statement
	// pages show what it did (tier A). Empty is a valid state: a party the
	// autopilot has not read yet renders as "not read yet", never as a
	// blank page pretending the programme has no content.
	mux.HandleFunc("GET /partier", func(w http.ResponseWriter, r *http.Request) {
		// ?tema=<slug> narrows every party to one dimension, so a reader can
		// compare them on one theme instead of scrolling nine long sections.
		view := buildProgramIndex(ds, buildProgramPages(ds), r.URL.Query().Get("tema"))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := programsTmpl.Execute(w, view); err != nil {
			logger.Error("render programs index", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /parti/{id}", func(w http.ResponseWriter, r *http.Request) {
		party, ok := ds.Party(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		page := programPage{Meta: partyMeta(party, 0), Party: party}
		if prog, ok := ds.Program(party.ID); ok {
			page.Program = prog
			page.Groups = groupPoints(ds, prog)
			page.HasPoints = len(prog.Points) > 0
			page.Meta = partyMeta(party, len(prog.Points))
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := programTmpl.Execute(w, page); err != nil {
			logger.Error("render party program", "err", err, "party", party.ID)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	// The budget page (2026-10-12): /budsjett is the whole statsbudsjett as
	// we collected and analysed it. A missing budget file is a 404, not an
	// empty page — the difference between "not collected" and "collected
	// and empty" matters here as much as it does for programmes.
	// The vote-alignment map: who votes with whom, computed from the vote index.
	// Every number on it is derived at request time (internal/analysis), and the
	// votes it rests on are listed with their proposal text so a reader can check
	// a cell instead of trusting it.
	mux.HandleFunc("GET /stemmer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := stemmerTmpl.Execute(w, buildStemmerView(ds, SnapshotDate)); err != nil {
			logger.Error("render stemmer", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	// The map: dimension by party, plus the time span behind each party's record.
	// Derived from the dataset at request time; see internal/analysis/profile.go
	// for why it draws no trend line.
	mux.HandleFunc("GET /kart", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := kartTmpl.Execute(w, buildKartView(ds, SnapshotDate)); err != nil {
			logger.Error("render kart", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	// Derived from the party-finance index at request time; the index is generated
	// by tools/analytics/build_finance_index.py and carries its own provenance.
	mux.HandleFunc("GET /penger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := pengerTmpl.Execute(w, buildPengerView(ds, SnapshotDate)); err != nil {
			logger.Error("render penger", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	// Derived from the state-ownership index at request time; the index is
	// generated by tools/analytics/build_ownership_index.py.
	mux.HandleFunc("GET /eierskap", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := eierskapTmpl.Execute(w, buildEierskapView(SnapshotDate)); err != nil {
			logger.Error("render eierskap", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	// Derived from the dataset at request time: the same cells as /kart, with the
	// years kept instead of averaged away.
	mux.HandleFunc("GET /tidslinje", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tidslinjeTmpl.Execute(w, buildYearView(SnapshotDate)); err != nil {
			logger.Error("render tidslinje", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	// Derived from the storting, proposals and audits indexes at request time.
	mux.HandleFunc("GET /makt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := maktTmpl.Execute(w, buildMaktView(SnapshotDate)); err != nil {
			logger.Error("render makt", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /budsjett", func(w http.ResponseWriter, r *http.Request) {
		view, ok := buildBudgetView(ds, "budsjett-2027")
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := budgetTmpl.Execute(w, view); err != nil {
			logger.Error("render budget", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /metode", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		evidenceTotal, evidenceWithQuote := quoteCounts(ds)
		po, cc := programmeOnlyCells(ds)
		vm := struct {
			Meta                 pageMeta
			SnapshotDate         string
			EvidenceTotal        int
			EvidenceWithQuote    int
			EvidenceWithoutQuote int
			ProgrammeOnly        int
			CodedCells           int
		}{
			Meta: meta("Metode · Ekte valgomat",
				"Slik blir en partiposisjon til: hva som er kilde, hva som er tolkning, og hvorfor tre separate mål aldri slås sammen til én score.",
				"/metode"),
			SnapshotDate:         SnapshotDate,
			EvidenceTotal:        evidenceTotal,
			EvidenceWithQuote:    evidenceWithQuote,
			EvidenceWithoutQuote: evidenceTotal - evidenceWithQuote,
			ProgrammeOnly:        po,
			CodedCells:           cc,
		}
		if err := metodeTmpl.Execute(w, vm); err != nil {
			logger.Error("render metode", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	// Open data. The JSON and CSV are generated from the same dataset the
	// pages render, so a reader checking the export and a reader checking a
	// statement card are looking at one source.
	mux.HandleFunc("GET /data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		total, withQuote := quoteCounts(ds)
		vm := struct {
			Meta           pageMeta
			SnapshotDate   string
			Licence        string
			StatementCount int
			EvidenceCount  int
			WithQuote      int
			ArchiveCount   int
			Columns        []evidenceColumn
		}{
			Meta: meta("Data · Ekte valgomat",
				"Hele datasettet bak valgomaten som filer: utsagn og påstander med kilde, som JSON og CSV, med forklaring av hver kolonne.",
				"/data"),
			SnapshotDate:   SnapshotDate,
			Licence:        licenceNote,
			StatementCount: len(ds.Statements),
			EvidenceCount:  total,
			WithQuote:      withQuote,
			ArchiveCount:   data.SourceWindowCount(),
			Columns:        evidenceColumns,
		}
		if err := dataTmpl.Execute(w, vm); err != nil {
			logger.Error("render data", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /bidra", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := bidraTmpl.Execute(w, buildBidraPage(ds, SnapshotDate)); err != nil {
			logger.Error("render bidra", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /data/utsagn.json", func(w http.ResponseWriter, r *http.Request) {
		serveStatementsJSON(w, ds)
	})

	mux.HandleFunc("GET /data/evidens.json", func(w http.ResponseWriter, r *http.Request) {
		serveEvidenceJSON(w, ds)
	})

	mux.HandleFunc("GET /data/evidens.csv", func(w http.ResponseWriter, r *http.Request) {
		serveEvidenceCSV(w, ds)
	})

	mux.HandleFunc("GET /data/kildetekst.json", func(w http.ResponseWriter, r *http.Request) {
		serveWindowsJSON(w)
	})

	mux.HandleFunc("GET /om", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		c44 := statementCoverage(ds, 44)
		c45 := statementCoverage(ds, 45)
		vm := struct {
			Meta          pageMeta
			SnapshotDate  string
			AddendumCount int
			CoreCount     int
			Corrections   []data.Correction
			C44           cellCoverage
			C45           cellCoverage
			NoPosition    []data.Statement
		}{
			Meta: meta("Om · Ekte valgomat",
				"Hvorfor Ekte valgomat finnes, hva som er bygget, og hva som ennå ikke er gjennomgått.",
				"/om"),
			SnapshotDate:  SnapshotDate,
			AddendumCount: addendumCount(ds),
			CoreCount:     coreStatementBuilt(ds),
			Corrections:   ds.Corrections,
			C44:           c44,
			C45:           c45,
			NoPosition:    uncomparableStatements(ds),
		}
		if err := omTmpl.Execute(w, vm); err != nil {
			logger.Error("render om", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /personvern", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		vm := struct {
			Meta         pageMeta
			SnapshotDate string
		}{
			Meta: meta("Personvern · Ekte valgomat",
				"Ingen besøkslogg, ingen konto, ingen sporing. Svarene dine blir i nettleseren din og sendes aldri til en server.",
				"/personvern"),
			SnapshotDate: SnapshotDate,
		}
		if err := personvernTmpl.Execute(w, vm); err != nil {
			logger.Error("render personvern", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	// /resultat is a dedicated, linkable page (2026-09-26) — previously
	// this was an in-page tab on "/" (see quiz.js), which worked but
	// had no URL of its own and no nav entry next to Om/Metode.
	// Rendering still happens entirely in the browser: this handler
	// only serves the same PayloadJSON the quiz itself uses, so
	// quiz.js can score the answers already sitting in localStorage —
	// nothing here ever sees a user's actual answers.
	mux.HandleFunc("GET /resultat", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		vm := struct {
			Meta         pageMeta
			SnapshotDate string
			PayloadJSON  template.JS
		}{
			Meta: meta("Resultat · Ekte valgomat",
				"Resultatet av valgomaten: hvilke partier du er mest på linje med, bygget på partiposisjoner med kilde. Regnes ut og vises bare i nettleseren din.",
				"/resultat"),
			SnapshotDate: SnapshotDate,
			PayloadJSON:  template.JS(payloadJSON), //nolint:gosec // our own generated JSON, never user input
		}
		if err := resultatTmpl.Execute(w, vm); err != nil {
			logger.Error("render resultat", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	logger.Info(fmt.Sprintf("loaded dataset: %d statements, %d parties, %d evidence items",
		len(ds.Statements), len(ds.Parties), len(ds.Evidence)))

	// No per-request log, on purpose: /personvern promises visitors that
	// nothing they do on the site is stored by us. Only errors are logged.
	// TestNoVisitLog holds that promise.
	return withSecurityHeaders(mux)
}
