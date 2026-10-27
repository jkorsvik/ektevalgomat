package data

import (
	"strconv"
	"testing"
)

// The finance index is a copy of someone else's numbers, so the tests here are
// about the copy being faithful and usable, not about the numbers being right.
// The source is authoritative; what can go wrong is my reading of it.
func TestFinanceIndexReconciles(t *testing.T) {
	rows := PartyFinances()
	if len(rows) < 100 {
		t.Fatalf("only %d party-years: the index is too thin to be the money layer", len(rows))
	}
	checked, unchecked := 0, 0
	for _, r := range rows {
		if r.Income == nil {
			t.Errorf("%s %s has no income figure but was stored as a row", r.Party, r.Year)
			continue
		}
		if r.PublicSupport != nil && r.OwnActivity != nil && r.Contributions != nil {
			sum := *r.PublicSupport + *r.OwnActivity + *r.Contributions
			// SSB rounds to whole kroner, so a few kroner of residual is rounding.
			if diff := sum - *r.Income; diff > 10 || diff < -10 {
				t.Errorf("%s %s: public+own+contributions = %d, income = %d (diff %d)",
					r.Party, r.Year, sum, *r.Income, diff)
			}
			checked++
		} else {
			unchecked++
		}
		if r.Contributions != nil {
			subs := []*int64{r.FromPrivatePersons, r.FromCompanies, r.FromWorkerOrgs, r.FromOtherOrgs, r.FromUncategorised}
			var sum int64
			present := false
			for _, s := range subs {
				if s != nil {
					sum += *s
					present = true
				}
			}
			if present {
				if diff := sum - *r.Contributions; diff > 10 || diff < -10 {
					t.Errorf("%s %s: contribution subtypes = %d, contributions = %d (diff %d)",
						r.Party, r.Year, sum, *r.Contributions, diff)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no row had the three top-level posts: the reconciliation checked nothing")
	}
	t.Logf("%d rows, %d fully reconciled, %d with an absent top-level post", len(rows), checked, unchecked)
}

func TestFinanceValuesAreNonNegative(t *testing.T) {
	for _, r := range PartyFinances() {
		for name, v := range map[string]*int64{
			"income": r.Income, "public": r.PublicSupport, "state": r.StateSupport,
			"own": r.OwnActivity, "contributions": r.Contributions,
			"persons": r.FromPrivatePersons, "companies": r.FromCompanies,
			"worker orgs": r.FromWorkerOrgs, "other orgs": r.FromOtherOrgs,
			"uncategorised": r.FromUncategorised,
		} {
			if v != nil && *v < 0 {
				t.Errorf("%s %s: %s is negative (%d)", r.Party, r.Year, name, *v)
			}
		}
	}
}

func TestFinanceCoversEveryScoredParty(t *testing.T) {
	known := map[string]bool{}
	for _, p := range Store.Parties {
		known[p.ID] = true
	}
	seen := map[string]int{}
	for _, r := range PartyFinances() {
		if !known[r.Party] {
			t.Errorf("row for unknown party %q", r.Party)
		}
		seen[r.Party]++
	}
	for id := range known {
		if seen[id] < 10 {
			t.Errorf("party %s has %d party-years; the money layer is missing for it", id, seen[id])
		}
	}
}

func TestFinanceYearsAreOrderedAndContiguous(t *testing.T) {
	years := FinanceYears()
	if len(years) < 15 {
		t.Fatalf("only %d years", len(years))
	}
	for i, y := range years {
		n, err := strconv.Atoi(y)
		if err != nil {
			t.Fatalf("year %q is not a number", y)
		}
		if i == 0 {
			continue
		}
		prev, _ := strconv.Atoi(years[i-1])
		if n != prev+1 {
			t.Errorf("years jump from %d to %d: a missing year would be invisible in the page", prev, n)
		}
	}
}

// The share is derived, so it must follow its inputs rather than be stored.
func TestStateShareFollowsItsInputs(t *testing.T) {
	income, state := int64(1000), int64(250)
	f := PartyFinance{Income: &income, StateSupport: &state}
	if got := f.StateShare(); got != 25 {
		t.Errorf("StateShare = %v, want 25", got)
	}
	// No income means no denominator, and a silent 0 would read as "no state money".
	if got := (PartyFinance{StateSupport: &state}).StateShare(); got != -1 {
		t.Errorf("StateShare without income = %v, want -1", got)
	}
	zero := int64(0)
	if got := (PartyFinance{Income: &zero, StateSupport: &state}).StateShare(); got != -1 {
		t.Errorf("StateShare with zero income = %v, want -1", got)
	}
	if got := (PartyFinance{Income: &income, Contributions: &state}).ContributionShare(); got != 25 {
		t.Errorf("ContributionShare = %v, want 25", got)
	}
}

// A number without a source is not usable in this project, so provenance is
// asserted rather than assumed.
func TestFinanceCarriesProvenance(t *testing.T) {
	table, api, human, retrieved, unit := FinanceSource()
	if table != "11233" {
		t.Errorf("table = %q, want 11233", table)
	}
	if api == "" || human == "" {
		t.Errorf("no URL for the source: api=%q human=%q", api, human)
	}
	if retrieved == "" {
		t.Error("no retrieved_at: the index cannot be shown to be current")
	}
	if unit != "kroner" {
		t.Errorf("unit = %q, want kroner", unit)
	}
}

func TestFinanceForFindsAndMisses(t *testing.T) {
	if _, ok := FinanceFor("finnes-ikke", "2024"); ok {
		t.Error("FinanceFor found a party that does not exist")
	}
	years := FinanceYears()
	last := years[len(years)-1]
	f, ok := FinanceFor("ap", last)
	if !ok {
		t.Fatalf("ap %s not found", last)
	}
	if f.Income == nil || *f.Income == 0 {
		t.Errorf("ap %s has no income: the row exists but carries nothing", last)
	}
}
