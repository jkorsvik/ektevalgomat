package data

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Reference data: who governed Norway, and who held which post.
//
// This is not evidence. It carries no position and makes no claim about what a
// party wants. It exists because every government attribution in the project
// rests on it: a 2026 proposition is an Ap-led government's, never "Ap/Sp", and
// Senterpartiet left on 2025-02-04. Before this table, that fact lived in a
// sentence of prose in tools/autopilot/method/budget.md, where nothing checked it.
//
// Shape, and why it is two files rather than one row per minister:
//
//   periods.yaml    one row per span in which the cabinet's party composition was
//                   constant. A change of composition splits a cabinet into two
//                   rows, so Solberg's four compositions are four rows and not one
//                   with a footnote. Boundary convention: a change date is the last
//                   day of the outgoing row AND the first day of the incoming one.
//   ministers.yaml  one row per minister per posting, quoted as regjeringen.no
//                   states it. Ministry renames are not normalised: a renamed
//                   department keeps the name the page uses for that posting.
//
// Provenance is per row, and a row without a source is a defect: an attribution
// table a reader cannot check is worse than no table, because it looks
// authoritative.

// GovernmentPeriod is one span of constant cabinet party composition.
type GovernmentPeriod struct {
	// Period is "YYYY-MM-DD..YYYY-MM-DD", inclusive at both ends, or
	// "YYYY-MM-DD..open" for the sitting cabinet.
	Period         string   `yaml:"period"`
	Cabinet        string   `yaml:"cabinet"`
	Parties        []string `yaml:"parties"`
	SourceName     string   `yaml:"source_name"`
	SourceURL      string   `yaml:"source_url"`
	RetrievedAt    string   `yaml:"retrieved_at"`
	VerbatimQuote  string   `yaml:"verbatim_quote"`
	EvidenceURLFor string   `yaml:"evidence_url_for_quote"`
}

// Start and End split Period. End is empty for a span that is still open.
func (p GovernmentPeriod) Start() string {
	if i := strings.Index(p.Period, ".."); i >= 0 {
		return p.Period[:i]
	}
	return ""
}

// End is the last day of the span, or "" when the cabinet is sitting.
func (p GovernmentPeriod) End() string {
	if i := strings.Index(p.Period, ".."); i >= 0 {
		e := p.Period[i+2:]
		if e == "open" {
			return ""
		}
		return e
	}
	return ""
}

// Covers reports whether a date falls inside the span. Dates are ISO, so string
// comparison is date comparison. A shared boundary date belongs to both rows,
// which is the convention the data records.
func (p GovernmentPeriod) Covers(date string) bool {
	start, end := p.Start(), p.End()
	if start == "" || date == "" {
		return false
	}
	if date < start {
		return false
	}
	return end == "" || date <= end
}

// Minister is one minister holding one post for one span.
type Minister struct {
	Minister   string `yaml:"minister"`
	Party      string `yaml:"party"` // dataset party id, or "" when the source states none
	Post       string `yaml:"post"`
	Department string `yaml:"department"`
	From       string `yaml:"from"`
	To         string `yaml:"to"` // "" while the posting is open
	SourceURL  string `yaml:"source_url"`
}

type governmentFile struct {
	Periods   []GovernmentPeriod
	Ministers []Minister
}

var government = mustLoadGovernment()

func mustLoadGovernment() governmentFile {
	var g governmentFile
	if err := readList(path.Join(contentRoot, "government", "periods.yaml"), &g.Periods); err != nil {
		panic(fmt.Sprintf("government/periods.yaml: %v", err))
	}
	if err := readList(path.Join(contentRoot, "government", "ministers.yaml"), &g.Ministers); err != nil {
		panic(fmt.Sprintf("government/ministers.yaml: %v", err))
	}
	sort.SliceStable(g.Periods, func(i, j int) bool { return g.Periods[i].Start() < g.Periods[j].Start() })
	return g
}

// readList decodes a YAML file holding a sequence. The rest of content/ is one
// object per file, so there is no list reader to reuse.
func readList(path string, out any) error {
	b, err := contentFS.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, out)
}

// GovernmentPeriods returns the periods, oldest first.
func GovernmentPeriods() []GovernmentPeriod { return government.Periods }

// Ministers returns every minister posting.
func Ministers() []Minister { return government.Ministers }

// GovernmentOn returns the period covering a date, and whether one was found.
// It is the function a page or a checker should use instead of reasoning about
// cabinets from memory: it is the only place the composition is looked up.
func GovernmentOn(date string) (GovernmentPeriod, bool) {
	for _, p := range government.Periods {
		if p.Covers(date) {
			return p, true
		}
	}
	return GovernmentPeriod{}, false
}
