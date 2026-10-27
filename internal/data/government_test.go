package data

import (
	"sort"
	"testing"
)

// The government table is the fact every attribution rests on, so it is checked
// harder than prose ever was. Each test here corresponds to a way the table could
// be wrong while looking right.
func TestGovernmentPeriodsAreContinuousAndNonOverlapping(t *testing.T) {
	periods := GovernmentPeriods()
	if len(periods) < 5 {
		t.Fatalf("only %d periods: the table cannot cover 2009-2026", len(periods))
	}
	for i, p := range periods {
		if p.Start() == "" {
			t.Fatalf("period %d has no start: %q", i, p.Period)
		}
		if p.Cabinet == "" {
			t.Errorf("period %s names no cabinet", p.Period)
		}
		if len(p.Parties) == 0 {
			t.Errorf("period %s names no parties", p.Period)
		}
		if p.SourceURL == "" || p.RetrievedAt == "" {
			t.Errorf("period %s has no provenance", p.Period)
		}
		if i == 0 {
			continue
		}
		prev := periods[i-1]
		// Boundary convention: adjacent rows share the change date, so the
		// previous row's end IS this row's start. A gap means a day on which the
		// composition is unknown, which would silently become "unknown cabinet".
		if prev.End() != p.Start() {
			t.Errorf("gap or overlap between %q (ends %q) and %q (starts %q)",
				prev.Period, prev.End(), p.Period, p.Start())
		}
		if prev.End() == "" {
			t.Errorf("%q is open but is followed by %q", prev.Period, p.Period)
		}
	}
}

func TestGovernmentCoversTheEvidenceRange(t *testing.T) {
	periods := GovernmentPeriods()
	if len(periods) == 0 {
		t.Fatal("no periods")
	}
	newest := ""
	for _, e := range Store.Evidence {
		if e.Date > newest {
			newest = e.Date
		}
	}
	if newest == "" {
		t.Fatal("no evidence date to compare against")
	}
	first, last := periods[0], periods[len(periods)-1]
	if first.Start() > "2009-01-01" {
		t.Errorf("table starts %s, after 2009-01-01", first.Start())
	}
	if last.End() == "" {
		return // a sitting cabinet covers everything after its start
	}
	if last.End() < newest {
		t.Errorf("table ends %s but the newest evidence is dated %s: a government would be unknown",
			last.End(), newest)
	}
}

func TestGovernmentPeriodsNameKnownParties(t *testing.T) {
	known := map[string]bool{}
	for _, p := range Store.Parties {
		known[p.ID] = true
	}
	for _, p := range GovernmentPeriods() {
		for _, id := range p.Parties {
			if !known[id] {
				t.Errorf("period %s names unknown party %q", p.Period, id)
			}
		}
	}
}

// The cross-check that makes the table worth having: a minister's party must be
// one of the parties in government on the day the posting started. This is the
// error the table exists to prevent, so it is also the one most worth testing.
func TestMinistersPartyMatchesTheSittingCabinet(t *testing.T) {
	ministers := Ministers()
	if len(ministers) < 100 {
		t.Fatalf("only %d minister rows: the table is incomplete", len(ministers))
	}
	known := map[string]bool{}
	for _, p := range Store.Parties {
		known[p.ID] = true
	}
	partyLess := 0
	for _, m := range ministers {
		if m.Minister == "" || m.Post == "" || m.From == "" {
			t.Errorf("incomplete minister row: %+v", m)
			continue
		}
		if m.To != "" && m.To < m.From {
			t.Errorf("%s %s: ends %s before it starts %s", m.Minister, m.Post, m.To, m.From)
		}
		if m.SourceURL == "" {
			t.Errorf("%s %s has no source", m.Minister, m.Post)
		}
		if m.Party == "" {
			partyLess++
			continue
		}
		if !known[m.Party] {
			t.Errorf("%s: unknown party %q", m.Minister, m.Party)
			continue
		}
		// A change date belongs to both adjacent periods, so any covering period
		// that lists the party is enough.
		ok := false
		for _, p := range GovernmentPeriods() {
			if !p.Covers(m.From) {
				continue
			}
			for _, id := range p.Parties {
				if id == m.Party {
					ok = true
				}
			}
		}
		if !ok {
			t.Errorf("%s (%s) starts %s, when no governing period includes that party", m.Minister, m.Party, m.From)
		}
	}
	// Two ministers are party-less on the source page. That is a fact, not a gap,
	// but the count is asserted so a sudden growth means something went wrong.
	if partyLess > 5 {
		t.Errorf("%d ministers have no party; the source states one for most of them", partyLess)
	}
	t.Logf("%d periods, %d minister postings, %d without a stated party",
		len(GovernmentPeriods()), len(ministers), partyLess)
}

func TestGovernmentOnFindsTheComposition(t *testing.T) {
	cases := []struct {
		date    string
		want    []string
		present bool
	}{
		{"2022-06-01", []string{"ap", "sp"}, true},
		{"2025-06-01", []string{"ap"}, true},
		{"2015-01-01", []string{"frp", "h"}, true},
		{"1990-01-01", nil, false}, // before the table starts
	}
	for _, c := range cases {
		p, ok := GovernmentOn(c.date)
		if ok != c.present {
			t.Errorf("GovernmentOn(%s) found=%v, want %v", c.date, ok, c.present)
			continue
		}
		if !ok {
			continue
		}
		// Copy before sorting: Parties is the loaded slice, and sorting it in place
		// would reorder the dataset for every later test in this package.
		got := append([]string(nil), p.Parties...)
		sort.Strings(got)
		if len(got) != len(c.want) {
			t.Errorf("GovernmentOn(%s) parties = %v, want %v", c.date, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("GovernmentOn(%s) parties = %v, want %v", c.date, got, c.want)
				break
			}
		}
	}
}
