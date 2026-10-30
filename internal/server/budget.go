package server

import (
	"strconv"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// nok formats thousand-kroner with Norwegian space grouping (779 424).
// The separator is U+00A0, not a plain space: "4 640" broke across a
// line on /budsjett (2026-10-13 review), leaving "4" and "640" on two
// lines. A figure is one thing, so its grouping must not be a break
// opportunity. Use \u00a0 in every template comparison of these numbers.
func nok(n int64) string {
	s := strconv.FormatInt(n, 10)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteRune('\u00a0')
		}
		b.WriteRune(c)
	}
	out := b.String()
	if neg {
		return "−" + out
	}
	return out
}

// pct formats a signed percentage with a comma decimal: +4,8, −1,6.
func pct(f float64) string {
	sign := "+"
	if f < 0 {
		sign = "−"
		f = -f
	}
	return sign + strings.Replace(strconv.FormatFloat(f, 'f', 1, 64), ".", ",", 1)
}

// The budget page (2026-10-12): /budsjett shows the whole statsbudsjett as we
// collected and analysed it — the headline figures, the per-department totals,
// and the findings that bear on the statements. It is the "what did the state
// actually decide about the money" page, built to sit beside the promises
// (/partier) and the votes (the statement pages). A finding with no statement
// is shown as such, never forced into the nearest one.

type budgetView struct {
	Meta     pageMeta
	Budget   data.Budget
	Findings []budgetFinding

	// PartyBudget is what each party would spend per rammeområde, read out of the
	// committee innstillinger. It is a different question from the government's
	// proposal above it, so it stands in its own section with its own caveats.
	PartyBudget     partyBudgetSection
	PartyBudgetLede string
}

type budgetFinding struct {
	Topic         string
	Note          string
	Quote         string
	SourceURL     string
	StatementID   string
	StatementNum  int
	StatementName string
	HasStatement  bool
	Observations  []budgetObservation
}

type budgetObservation struct {
	Year   int
	Stage  string // the party-readable stage label
	Amount string // formatted kroner
	Latest bool
}

// budgetStageLabel names an observation's stage in words a reader can act
// on. "forslag" and "saldert" must never render as the same thing: one is
// what the government asked for, the other is what the Stortinget granted.
func budgetStageLabel(stage string) string {
	labels := map[string]string{
		"forslag":  "Regjeringens forslag",
		"saldert":  "Saldert budsjett (vedtatt)",
		"regnskap": "Regnskap",
	}
	if l, ok := labels[stage]; ok {
		return l
	}
	return stage
}

func buildBudgetView(ds *data.Dataset, budgetID string) (budgetView, bool) {
	b, ok := ds.Budget(budgetID)
	if !ok {
		return budgetView{}, false
	}
	v := budgetView{
		Meta: meta("Statsbudsjettet "+strconv.Itoa(b.Year)+" · Ekte valgomat",
			"Statsbudsjettet "+strconv.Itoa(b.Year)+" samlet og analysert: forslaget, utgiftene per departement, og hva som er endret fra året før.",
			"/budsjett"),
		Budget: b,
	}
	for _, f := range b.Findings {
		bf := budgetFinding{
			Topic: f.Topic, Note: f.Note, Quote: f.Quote,
			SourceURL: f.SourceURL, StatementID: f.StatementID,
		}
		for j, o := range f.Observed {
			bf.Observations = append(bf.Observations, budgetObservation{
				Year: o.FiscalYear, Stage: budgetStageLabel(o.Stage),
				Amount: nok(o.Amount), Latest: j == len(f.Observed)-1,
			})
		}
		if f.StatementID != "" {
			if s, ok := ds.Statement(f.StatementID); ok {
				bf.HasStatement = true
				bf.StatementNum = s.Number
				bf.StatementName = s.ShortName
			}
		}
		v.Findings = append(v.Findings, bf)
	}
	v.PartyBudget = buildPartyBudgetSection(ds)
	v.PartyBudgetLede = partyBudgetLede(v.PartyBudget)
	return v, true
}
