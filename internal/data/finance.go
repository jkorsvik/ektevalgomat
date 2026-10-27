package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// The party-finance index is generated data, not source: it is fetched from SSB
// table 11233 by tools/analytics/build_finance_index.py and committed, so the
// numbers can be re-derived and a stale index is visible as stale (see
// Source.RetrievedAt).
//
// It exists because the dataset could say what the parties promise and how they
// vote, but not who pays for them. Every statement about public money in this
// project rested on documents; this is the state's own annual accounts, per party
// per year, in kroner.
//
// What it is not: a measure of anonymous donations. SSB has no such field. The
// nearest, FromUncategorised, is a residual for givers outside the other
// categories, and the field comment says so because the temptation to read it as
// anonymity is exactly how this table would be misused.
//
//go:embed finance/index.json
var financeIndexJSON []byte

// PartyFinance is one party's reported accounts for one year, in kroner.
// A nil field means the post is absent from the source for that year, which is
// different from zero: zero income is a claim, an absent post is a gap.
type PartyFinance struct {
	Party              string `json:"party"`
	Year               string `json:"year"`
	Income             *int64 `json:"income"`
	PublicSupport      *int64 `json:"public_support"`
	StateSupport       *int64 `json:"state_support"`
	OwnActivity        *int64 `json:"own_activity"`
	Contributions      *int64 `json:"contributions"`
	FromPrivatePersons *int64 `json:"from_private_persons"`
	FromCompanies      *int64 `json:"from_companies"`
	FromWorkerOrgs     *int64 `json:"from_worker_orgs"`
	FromOtherOrgs      *int64 `json:"from_other_orgs"`
	// FromUncategorised is SSB's "Bidrag fra andre, ikke kategoriserte givere":
	// givers outside the four named categories. It is not an anonymity count.
	FromUncategorised *int64 `json:"from_uncategorised"`
}

// StateShare is the share of income that came from the state, in percent, or -1
// when either figure is absent. Returned as a float so a caller can format it;
// it is never stored, because a stored percentage drifts from its numerator.
func (f PartyFinance) StateShare() float64 {
	if f.Income == nil || f.StateSupport == nil || *f.Income <= 0 {
		return -1
	}
	return 100 * float64(*f.StateSupport) / float64(*f.Income)
}

// ContributionShare is the share of income that came from contributions of any
// kind, in percent, or -1 when a figure is absent.
func (f PartyFinance) ContributionShare() float64 {
	if f.Income == nil || f.Contributions == nil || *f.Income <= 0 {
		return -1
	}
	return 100 * float64(*f.Contributions) / float64(*f.Income)
}

type financeSource struct {
	Table       string `json:"table"`
	Title       string `json:"title"`
	API         string `json:"api"`
	HumanURL    string `json:"human_url"`
	RetrievedAt string `json:"retrieved_at"`
	Unit        string `json:"unit"`
	Note        string `json:"note"`
}

type financeIndexFile struct {
	Source  financeSource     `json:"source"`
	Posts   map[string]string `json:"posts"`
	Parties map[string]struct {
		SSBCode string `json:"ssb_code"`
	} `json:"parties"`
	Years []string       `json:"years"`
	Rows  []PartyFinance `json:"rows"`
}

var financeIndexDoc = mustLoadFinance()

func mustLoadFinance() financeIndexFile {
	var idx financeIndexFile
	if err := json.Unmarshal(financeIndexJSON, &idx); err != nil {
		// Same reasoning as the vote index: the generator is in the repo and its
		// output is committed, so a malformed file is a build defect.
		panic(fmt.Sprintf("internal/data/finance/index.json is not valid: %v", err))
	}
	if len(idx.Rows) == 0 {
		panic("internal/data/finance/index.json holds no rows; run tools/analytics/build_finance_index.py")
	}
	if idx.Source.RetrievedAt == "" || idx.Source.API == "" {
		panic("internal/data/finance/index.json has no provenance; a number without a source is not usable")
	}
	return idx
}

// PartyFinances returns every party-year, oldest first then by party id, so a
// page renders deterministically.
func PartyFinances() []PartyFinance { return financeIndexDoc.Rows }

// FinanceYears returns the years the index covers, oldest first.
func FinanceYears() []string { return financeIndexDoc.Years }

// FinanceSource returns the provenance of the index: the table, the API it came
// from, and the date it was read.
func FinanceSource() (table, api, humanURL, retrievedAt, unit string) {
	s := financeIndexDoc.Source
	return s.Table, s.API, s.HumanURL, s.RetrievedAt, s.Unit
}

// FinanceFor returns one party's accounts for one year, and whether a row exists.
// An absent row is a party that filed no accounts that year, not zero income.
func FinanceFor(partyID, year string) (PartyFinance, bool) {
	for _, r := range financeIndexDoc.Rows {
		if r.Party == partyID && r.Year == year {
			return r, true
		}
	}
	return PartyFinance{}, false
}
