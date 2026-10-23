package data

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// content/ is embedded, not read from a live filesystem — the runtime
// container is FROM scratch, read-only, with no shell and no writable
// paths (see docker-compose.yml). Editing evidence means: change the
// YAML, open a pull request, and let the existing poll-deploy.sh cron
// rebuild within a few minutes.
//
// The record of what changed is NOT primarily the git history. The published
// history starts at the initial public release (fbc64d8, 2026-10-23) and every
// change since is a commit, but nothing before that release is published, so
// the history a reader can see begins where the project became public rather
// than where it began. Commits are also a developer's note to other
// developers. Two files carry the record instead, and both ship with the
// code:
//
//   - docs/roadmap.md — the day-by-day build log, including the research
//     reversals that never reached the site.
//   - internal/data/content/corrections/ — the same mistakes, in the
//     project's own words, rendered on /om for a reader who does not want
//     to read a build log (see the Correction type).
//
// Neither is tamper-evident. That is the honest trade: an append-only
// record a reader can check is worth more than a cryptographic one they
// cannot, and a reader can check these two.
//
//go:embed content
var contentFS embed.FS

const contentRoot = "content"

// validCorrectionStatementNumber is the shape of a statement number in the
// corrections log: two digits, as the site writes them ("07", "44"). Not
// validated against the live statement set — the log records history, and
// one of the corrections removed a statement number that no longer exists.
var validCorrectionStatementNumber = regexp.MustCompile(`^[0-9]{2}$`)

// Store is the fully loaded, ready-to-query dataset — built once at
// package init. A malformed evidence file panics at startup: the same
// "fail fast, don't quietly publish half-broken evidence" choice.
var Store = mustLoad()

func mustLoad() *Dataset {
	ds, err := load()
	if err != nil {
		panic("data: loading content: " + err.Error())
	}
	return ds
}

// Dataset is every statement, party, and evidence object loaded at
// startup, plus small lookup indexes so callers don't re-scan the slices
// on every request.
type Dataset struct {
	Statements     []Statement
	Parties        []Party
	Evidence       []Evidence
	Programs       []Program
	Budgets        []Budget
	Corrections    []Correction // newest first
	statementByID  map[string]Statement
	partyByID      map[string]Party
	programByID    map[string]Program
	budgetByID     map[string]Budget
	evidenceByStmt map[string][]Evidence // includes secondary tags
}

func (d *Dataset) Statement(id string) (Statement, bool) {
	s, ok := d.statementByID[id]
	return s, ok
}

func (d *Dataset) Party(id string) (Party, bool) {
	p, ok := d.partyByID[id]
	return p, ok
}

// Program returns the party's loaded 2025-2029 programme points.
func (d *Dataset) Program(id string) (Program, bool) {
	p, ok := d.programByID[id]
	return p, ok
}

// Budget returns a loaded statsbudsjett by id (e.g. "budsjett-2027").
func (d *Dataset) Budget(id string) (Budget, bool) {
	b, ok := d.budgetByID[id]
	return b, ok
}

// EvidenceForStatement returns every evidence item tagged (primary or
// secondary) to statementID, sorted by date string ascending — good
// enough while every date is either an ISO date or a year range, which
// sorts correctly as plain text in this small a dataset.
func (d *Dataset) EvidenceForStatement(statementID string) []Evidence {
	items := d.evidenceByStmt[statementID]
	out := make([]Evidence, len(items))
	copy(out, items)
	sort.Slice(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out
}

// EvidenceForCell returns evidence tying a specific party to a specific
// statement — the actual scoring input for one grid cell. Institutional
// evidence (PartyID == "") is deliberately excluded: it's context for the
// statement, never a party's own position.
func (d *Dataset) EvidenceForCell(statementID, partyID string) []Evidence {
	var out []Evidence
	for _, e := range d.EvidenceForStatement(statementID) {
		if e.PartyID == partyID {
			out = append(out, e)
		}
	}
	return out
}

// validEvidenceClasses are the only evidence_class values the UI knows
// how to badge (see types.go's Evidence.EvidenceClass doc for the rule
// each one follows) — an unrecognized value fails the build the same
// way an unknown party_id or statement_id does, rather than silently
// rendering as an unlabeled or default-styled card.
var validEvidenceClasses = map[string]bool{
	"fakta":                true,
	"partiets_begrunnelse": true,
	"analyse":              true,
	"uklart":               true,
}

// validEvidensstyrkeCaps are the only evidensstyrke_cap values authors
// may set (see Evidence.EvidensstyrkeCap doc) — "" (absent, no cap) is
// valid and by far the common case, so it is not in this map; only a
// non-empty value must match one of these.
var validEvidensstyrkeCaps = map[string]bool{
	"svak":    true,
	"moderat": true,
}

// validSourceTiers are the source classes /metode explains (A strongest,
// E an unverified lead). The strength label picks the best tier by
// string order, so an empty or unknown tier would sort before "A" and
// silently turn a cell "Svak". It fails the build instead.
var validSourceTiers = map[string]bool{"A": true, "B": true, "C": true, "D": true, "E": true}

func load() (*Dataset, error) {

	statements, err := readDir[Statement](path.Join(contentRoot, "statements"))
	if err != nil {
		return nil, fmt.Errorf("statements: %w", err)
	}
	sort.Slice(statements, func(i, j int) bool { return statements[i].Number < statements[j].Number })

	parties, err := readDir[Party](path.Join(contentRoot, "parties"))
	if err != nil {
		return nil, fmt.Errorf("parties: %w", err)
	}

	evidence, err := readDir[Evidence](path.Join(contentRoot, "evidence"))
	if err != nil {
		return nil, fmt.Errorf("evidence: %w", err)
	}

	// programs/ is allowed to be absent: the autopilot fills it party by
	// party, so an empty dataset must not fail the build (same "empty is a
	// valid result" rule the whole project runs on).
	programs, err := readDir[Program](path.Join(contentRoot, "programs"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("programs: %w", err)
	}

	// budget/ is allowed to be absent, same rule.
	budgets, err := readDir[Budget](path.Join(contentRoot, "budget"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("budget: %w", err)
	}

	// corrections/ is required, unlike programs/ and budget/. An empty
	// corrections log is not an honest empty result: this project has
	// corrected itself in public-facing ways, and a missing directory
	// would silently turn /om's log into a claim that it never had.
	corrections, err := readDir[Correction](path.Join(contentRoot, "corrections"))
	if err != nil {
		return nil, fmt.Errorf("corrections: %w", err)
	}
	if len(corrections) == 0 {
		return nil, errors.New("corrections: no entries; the public corrections log must not be empty")
	}
	// Newest first. Two entries from the same day keep the order their
	// filenames give them (oldest event first, as the roadmap records
	// them), so a correction and the correction OF that correction read
	// in the order they happened.
	sort.Slice(corrections, func(i, j int) bool {
		if corrections[i].Date != corrections[j].Date {
			return corrections[i].Date > corrections[j].Date
		}
		return corrections[i].ID < corrections[j].ID
	})
	seenCorrection := make(map[string]bool, len(corrections))
	for _, c := range corrections {
		if _, err := time.Parse("2006-01-02", c.Date); err != nil {
			return nil, fmt.Errorf("correction %s: date %q is not YYYY-MM-DD", c.ID, c.Date)
		}
		if c.ID == "" {
			return nil, errors.New("correction: missing id")
		}
		if seenCorrection[c.ID] {
			return nil, fmt.Errorf("correction %s: duplicate id", c.ID)
		}
		seenCorrection[c.ID] = true
		if !validCorrectionKinds[c.Kind] {
			return nil, fmt.Errorf("correction %s: unknown or missing kind %q (want rettelse or nesten)", c.ID, c.Kind)
		}
		for _, field := range []struct{ name, value string }{
			{"title", c.Title},
			{"what", c.What},
			{"reason", c.Reason},
			{"effect", c.Effect},
		} {
			if strings.TrimSpace(field.value) == "" {
				return nil, fmt.Errorf("correction %s: missing %s", c.ID, field.name)
			}
		}
		for _, n := range c.Statements {
			if !validCorrectionStatementNumber.MatchString(n) {
				return nil, fmt.Errorf("correction %s: statement %q is not a statement number (want two digits, e.g. \"07\")", c.ID, n)
			}
		}
	}

	ds := &Dataset{
		Statements:     statements,
		Parties:        parties,
		Evidence:       evidence,
		Programs:       programs,
		Budgets:        budgets,
		Corrections:    corrections,
		statementByID:  make(map[string]Statement, len(statements)),
		partyByID:      make(map[string]Party, len(parties)),
		programByID:    make(map[string]Program, len(programs)),
		budgetByID:     make(map[string]Budget, len(budgets)),
		evidenceByStmt: make(map[string][]Evidence),
	}
	for _, s := range statements {
		ds.statementByID[s.ID] = s
	}
	for _, p := range parties {
		ds.partyByID[p.ID] = p
	}
	for _, p := range programs {
		if _, ok := ds.partyByID[p.ID]; !ok {
			return nil, fmt.Errorf("program %s: unknown party_id", p.ID)
		}
		if p.SourceURL == "" || p.SourceName == "" {
			return nil, fmt.Errorf("program %s: missing source_name or source_url", p.ID)
		}
		for i, pt := range p.Points {
			if strings.TrimSpace(pt.Quote) == "" {
				return nil, fmt.Errorf("program %s point %d: empty quote", p.ID, i)
			}
			seen := map[string]bool{}
			for _, sid := range pt.Statements {
				if _, ok := ds.statementByID[sid]; !ok {
					return nil, fmt.Errorf("program %s point %d: unknown statement %q", p.ID, i, sid)
				}
				if seen[sid] {
					return nil, fmt.Errorf("program %s point %d: statement %s listed twice", p.ID, i, sid)
				}
				seen[sid] = true
				// A cross-dimension link is allowed: the statement
				// taxonomy and the programme taxonomy overlap without
				// being identical, so the same subject can sit in two
				// places (data-centre governance is "Teknologisk
				// kapasitet" in the programmes and "Norsk handlefrihet"
				// for utsagn-11). Seven such links exist and are listed,
				// with a reason, in program_links_test.go — that test is
				// the gate, so an eighth cannot appear unnoticed.
			}
		}
		ds.programByID[p.ID] = p
	}
	for _, e := range evidence {
		if e.PartyID != "" {
			if _, ok := ds.partyByID[e.PartyID]; !ok {
				return nil, fmt.Errorf("evidence %s: unknown party_id %q", e.ID, e.PartyID)
			}
		}
		if _, ok := ds.statementByID[e.StatementID]; !ok {
			return nil, fmt.Errorf("evidence %s: unknown statement_id %q", e.ID, e.StatementID)
		}
		if !validEvidenceClasses[e.EvidenceClass] {
			return nil, fmt.Errorf("evidence %s: unknown or missing evidence_class %q", e.ID, e.EvidenceClass)
		}
		if !validSourceTiers[e.SourceTier] {
			return nil, fmt.Errorf("evidence %s: unknown or missing source_tier %q", e.ID, e.SourceTier)
		}
		if e.EvidensstyrkeCap != "" && !validEvidensstyrkeCaps[e.EvidensstyrkeCap] {
			return nil, fmt.Errorf("evidence %s: unknown evidensstyrke_cap %q", e.ID, e.EvidensstyrkeCap)
		}
		if e.ScanSummary == "" {
			return nil, fmt.Errorf("evidence %s: missing scan_summary", e.ID)
		}
		ds.evidenceByStmt[e.StatementID] = append(ds.evidenceByStmt[e.StatementID], e)
		for _, sec := range e.SecondaryStatementIDs {
			if _, ok := ds.statementByID[sec]; !ok {
				return nil, fmt.Errorf("evidence %s: unknown secondary_statement_id %q", e.ID, sec)
			}
			ds.evidenceByStmt[sec] = append(ds.evidenceByStmt[sec], e)
		}
	}
	for _, b := range budgets {
		if b.SourceURL == "" || b.SourceName == "" {
			return nil, fmt.Errorf("budget %s: missing source_name or source_url", b.ID)
		}
		for i, f := range b.Findings {
			if f.StatementID != "" {
				if _, ok := ds.statementByID[f.StatementID]; !ok {
					return nil, fmt.Errorf("budget %s finding %d: unknown statement_id %q", b.ID, i, f.StatementID)
				}
			}
			if strings.TrimSpace(f.Note) == "" {
				return nil, fmt.Errorf("budget %s finding %d: empty note", b.ID, i)
			}
			for j, o := range f.Observed {
				if !validBudgetStages[o.Stage] {
					return nil, fmt.Errorf("budget %s finding %d observation %d: unknown stage %q (want forslag, saldert or regnskap)", b.ID, i, j, o.Stage)
				}
				if o.FiscalYear < 2000 {
					return nil, fmt.Errorf("budget %s finding %d observation %d: fiscal_year %d looks wrong", b.ID, i, j, o.FiscalYear)
				}
			}
		}
		ds.budgetByID[b.ID] = b
	}
	return ds, nil
}

func readYAML(name string, v any) error {
	b, err := contentFS.ReadFile(name)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, v)
}

func readDir[T any](dirPath string) ([]T, error) {
	entries, err := fs.ReadDir(contentFS, dirPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", dirPath, err)
	}
	items := make([]T, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		var item T
		if err := readYAML(path.Join(dirPath, e.Name()), &item); err != nil {
			return nil, fmt.Errorf("%s/%s: %w", dirPath, e.Name(), err)
		}
		items = append(items, item)
	}
	return items, nil
}

// ProgramPointRef is a programme point plus the party that holds it.
type ProgramPointRef struct {
	PartyID string
	Point   ProgramPoint
}

// ProgramPointsForStatement returns every programme point that names
// statementID. This is the promise step's real input: unlike evidence, a
// programme point is not tagged to a statement in the evidence files —
// the link lives in the programme itself and is set by hand, with a
// reason, because no rule finds it (see the join analysis doc).
func (d *Dataset) ProgramPointsForStatement(statementID string) []ProgramPointRef {
	var out []ProgramPointRef
	for _, p := range d.Programs {
		for _, pt := range p.Points {
			for _, sid := range pt.Statements {
				if sid == statementID {
					out = append(out, ProgramPointRef{PartyID: p.ID, Point: pt})
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PartyID != out[j].PartyID {
			return out[i].PartyID < out[j].PartyID
		}
		return out[i].Point.Topic < out[j].Point.Topic
	})
	return out
}

// ProgramPointsForDimension returns the party's programme points in one
// dimension. A statement's dimension and a programme's dimension share a
// namespace — a statement in "Norsk handlefrihet" is answered by the
// party's points in "Norsk handlefrihet" — which is what makes the
// theme-level promise possible with no new data and no interpretation.
func (d *Dataset) ProgramPointsForDimension(partyID, dimension string) []ProgramPoint {
	p, ok := d.programByID[partyID]
	if !ok {
		return nil
	}
	var out []ProgramPoint
	for _, pt := range p.Points {
		if pt.Dimension == dimension {
			out = append(out, pt)
		}
	}
	return out
}
