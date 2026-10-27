package server

import (
	"strconv"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/analysis"
	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The money view: who paid for the parties, per year, from SSB's annual party
// accounts. Everything is derived from the finance index at request time.
//
// Two choices shape the page. First, kroner are shown in millions with one
// decimal, because the useful comparison is between parties and between years, and
// nine-digit numbers defeat that. Second, a share is always shown next to the
// absolute figure it comes from: state support of 133 million is a different fact
// for a party with 270 million in income than for one with 60, and a page that
// prints only the absolute invites the wrong comparison.
//
// The page states what the source cannot answer. SSB has no field for anonymous
// donations, so this page does not claim one, and says so where a reader would
// otherwise assume the gap is my omission.
type pengerCell struct {
	Millions string
	Has      bool
}

type pengerRow struct {
	Year  string
	Cells []pengerCell
}

type pengerParty struct {
	ID    string
	Short string
	Name  string
}

type pengerLatest struct {
	Short         string
	Name          string
	Income        string
	StateSupport  string
	StateShare    string
	Contrib       string
	ContribShare  string
	Uncategorised string
}

// RegisterRows is the register's published total per year, and RegisterLatest the
// per-party figure for the newest year it covers.
type pengerRegisterTotal struct {
	Year     string
	Millions string
}

type pengerRegisterParty struct {
	Short    string
	Name     string
	Millions string
}

type pengerView struct {
	Meta            pageMeta
	SnapshotDate    string
	Parties         []pengerParty
	Rows            []pengerRow
	Latest          []pengerLatest
	LatestYear      string
	FirstYear       string
	TotalPartyYears int
	Table           string
	API             string
	HumanURL        string
	RetrievedAt     string
	Unit            string

	// The register block. Kept apart from the SSB figures above on the page as
	// well as in the data, because the two measure different things.
	RegisterTotals    []pengerRegisterTotal
	RegisterLatest    []pengerRegisterParty
	RegisterYear      string
	RegisterName      string
	RegisterURL       string
	RegisterRetrieved string
	RegisterNote      string
}

// kronerMill formats kroner as millions with a Norwegian decimal comma, or "" for
// an absent figure. Absent is rendered as an empty cell, never as 0,00: a party
// that filed no accounts has no income figure, which is not the same as none.
func kronerMill(k *int64) string {
	if k == nil {
		return ""
	}
	s := strconv.FormatFloat(float64(*k)/1e6, 'f', 1, 64)
	return strings.Replace(s, ".", ",", 1)
}

// percent formats a derived share, or "" when the share has no denominator.
func percent(v float64) string {
	if v < 0 {
		return ""
	}
	s := strconv.FormatFloat(v, 'f', 1, 64)
	return strings.Replace(s, ".", ",", 1)
}

func buildPengerView(ds *data.Dataset, snapshotDate string) pengerView {
	short := map[string]string{}
	name := map[string]string{}
	for _, p := range ds.Parties {
		short[p.ID] = p.Short
		name[p.ID] = p.Name
	}

	// Party order follows the dataset's own party order, so the columns match the
	// rest of the site rather than SSB's numbering.
	var parties []pengerParty
	for _, id := range analysis.PartyIDs(ds) {
		if _, ok := data.FinanceFor(id, data.FinanceYears()[0]); ok {
			parties = append(parties, pengerParty{ID: id, Short: short[id], Name: name[id]})
			continue
		}
		// A party may be missing the first year and present later; check any year.
		for _, y := range data.FinanceYears() {
			if _, ok := data.FinanceFor(id, y); ok {
				parties = append(parties, pengerParty{ID: id, Short: short[id], Name: name[id]})
				break
			}
		}
	}

	years := data.FinanceYears()
	var rows []pengerRow
	for _, y := range years {
		row := pengerRow{Year: y}
		for _, p := range parties {
			f, ok := data.FinanceFor(p.ID, y)
			if !ok || f.StateSupport == nil {
				row.Cells = append(row.Cells, pengerCell{})
				continue
			}
			row.Cells = append(row.Cells, pengerCell{Millions: kronerMill(f.StateSupport), Has: true})
		}
		rows = append(rows, row)
	}

	var latest []pengerLatest
	latestYear := ""
	if len(years) > 0 {
		latestYear = years[len(years)-1]
	}
	for _, p := range parties {
		f, ok := data.FinanceFor(p.ID, latestYear)
		if !ok {
			continue
		}
		latest = append(latest, pengerLatest{
			Short:         p.Short,
			Name:          p.Name,
			Income:        kronerMill(f.Income),
			StateSupport:  kronerMill(f.StateSupport),
			StateShare:    percent(f.StateShare()),
			Contrib:       kronerMill(f.Contributions),
			ContribShare:  percent(f.ContributionShare()),
			Uncategorised: kronerMill(f.FromUncategorised),
		})
	}

	var regTotals []pengerRegisterTotal
	for _, r := range data.RegisterTotals() {
		amount := r.Amount
		regTotals = append(regTotals, pengerRegisterTotal{Year: r.Year, Millions: kronerMill(&amount)})
	}
	regYears := data.RegisterTotals()
	regYear := ""
	if len(regYears) > 0 {
		regYear = regYears[len(regYears)-1].Year
	}
	var regLatest []pengerRegisterParty
	for _, p := range parties {
		if amount, ok := data.RegisterFor(p.ID, regYear); ok {
			v := amount
			regLatest = append(regLatest, pengerRegisterParty{Short: p.Short, Name: p.Name, Millions: kronerMill(&v)})
		}
	}
	regName, _, regURL, regRetrieved, regNote := data.RegisterSourceInfo()

	table, api, human, retrieved, unit := data.FinanceSource()
	firstYear := ""
	if len(years) > 0 {
		firstYear = years[0]
	}
	return pengerView{
		Meta: meta("Penger · hvem betaler for partiene",
			"Partifinansiering per parti og år, fra SSB tabell 11233: statlig støtte, bidrag og hvor stor del av inntekten de utgjør.",
			"/penger"),
		SnapshotDate:    snapshotDate,
		Parties:         parties,
		Rows:            rows,
		Latest:          latest,
		LatestYear:      latestYear,
		FirstYear:       firstYear,
		TotalPartyYears: len(data.PartyFinances()),
		Table:           table,
		API:             api,
		HumanURL:        human,
		RetrievedAt:     retrieved,
		Unit:            unit,

		RegisterTotals:    regTotals,
		RegisterLatest:    regLatest,
		RegisterYear:      regYear,
		RegisterName:      regName,
		RegisterURL:       regURL,
		RegisterRetrieved: regRetrieved,
		RegisterNote:      regNote,
	}
}
