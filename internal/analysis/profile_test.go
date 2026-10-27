package analysis

import (
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

func pos(n int) *int { return &n }

// The profile is the only place on the site that claims to say when a party's
// record reaches back to, so each rule that makes those years honest is tested
// separately.
func TestDimensionProfileCountsOnlyCodedCells(t *testing.T) {
	ds := &data.Dataset{
		Statements: []data.Statement{
			{ID: "s1", Dimension: "Demo"},
			{ID: "s2", Dimension: "Demo"},
			{ID: "s3", Dimension: "Other"},
		},
		Evidence: []data.Evidence{
			{StatementID: "s1", PartyID: "ap", Date: "2020-05-01", Position: pos(2)},
			// s2 has a source but no coded position: it must not raise Coded.
			{StatementID: "s2", PartyID: "ap", Date: "2026-01-01"},
			{StatementID: "s3", PartyID: "ap", Date: "2024-03-03", Position: pos(-1)},
			// another party must not leak into this party's profile
			{StatementID: "s1", PartyID: "h", Date: "1999-01-01", Position: pos(-2)},
		},
	}
	rows := DimensionProfile(ds, "ap")
	if len(rows) != 2 {
		t.Fatalf("got %d dimension rows, want 2", len(rows))
	}
	demo := rows[0]
	if demo.Dimension != "Demo" || demo.Statements != 2 {
		t.Fatalf("first row = %+v, want Demo with 2 statements", demo)
	}
	if demo.Coded != 1 {
		t.Errorf("Coded = %d, want 1 (s2 has a source but no position)", demo.Coded)
	}
	if demo.Mean != 2 {
		t.Errorf("Mean = %v, want 2", demo.Mean)
	}
	if demo.OldestYear != 2020 || demo.NewestYear != 2020 {
		t.Errorf("years = %d..%d, want 2020..2020 (the 1999 row belongs to h)", demo.OldestYear, demo.NewestYear)
	}
	if demo.SingleDate != 1 || demo.MultiDate != 0 {
		t.Errorf("single/multi = %d/%d, want 1/0", demo.SingleDate, demo.MultiDate)
	}
}

func TestDimensionProfileSeparatesOneDateFromTwo(t *testing.T) {
	ds := &data.Dataset{
		Statements: []data.Statement{
			{ID: "s1", Dimension: "Demo"},
			{ID: "s2", Dimension: "Demo"},
			{ID: "s3", Dimension: "Demo"},
		},
		Evidence: []data.Evidence{
			// one date, two items: still a single observation in time
			{StatementID: "s1", PartyID: "ap", Date: "2024-01-01", Position: pos(1)},
			{StatementID: "s1", PartyID: "ap", Date: "2024-01-01", Position: pos(1)},
			// two dates: the only shape that can support a trend
			{StatementID: "s2", PartyID: "ap", Date: "2022-01-01", Position: pos(-1)},
			{StatementID: "s2", PartyID: "ap", Date: "2025-01-01", Position: pos(1)},
			// no usable date
			{StatementID: "s3", PartyID: "ap", Date: "ukjent", Position: pos(0)},
		},
	}
	row := DimensionProfile(ds, "ap")[0]
	if row.Coded != 3 {
		t.Fatalf("Coded = %d, want 3", row.Coded)
	}
	if row.SingleDate != 1 || row.MultiDate != 1 || row.WithoutYear != 1 {
		t.Errorf("single/multi/undated = %d/%d/%d, want 1/1/1", row.SingleDate, row.MultiDate, row.WithoutYear)
	}
	if row.OldestYear != 2022 || row.NewestYear != 2025 {
		t.Errorf("years = %d..%d, want 2022..2025", row.OldestYear, row.NewestYear)
	}
	if row.Span() != 3 {
		t.Errorf("Span = %d, want 3", row.Span())
	}
}

func TestMeanLabelIsAbsentWhenNothingIsCoded(t *testing.T) {
	if _, ok := (DimensionRow{}).MeanLabel(); ok {
		t.Error("MeanLabel reported a value for a row with nothing coded: \"unknown\" would render as 0.0")
	}
	if got, ok := (DimensionRow{Coded: 2, Mean: 1.5}).MeanLabel(); !ok || got != "1.5" {
		t.Errorf("MeanLabel = %q/%v, want 1.5/true", got, ok)
	}
}

func TestYearOfRejectsFreeFormDates(t *testing.T) {
	cases := map[string]bool{
		"2026-01-01": true, "2025 (partiprogram)": true,
		"": false, "ukjent": false, "26": false, "0001-01-01": false,
	}
	for in, want := range cases {
		if _, got := yearOf(in); got != want {
			t.Errorf("yearOf(%q) ok = %v, want %v", in, got, want)
		}
	}
}

// And on the real dataset: every party's coded count must equal the number of
// cells the scoring package says are coded, or this view disagrees with the rest
// of the site about the same facts.
func TestDimensionProfileAgreesWithTheDataset(t *testing.T) {
	ds := data.Store
	parties := PartyIDs(ds)
	if len(parties) != 9 {
		t.Fatalf("got %d parties, want 9", len(parties))
	}
	totalStatements := 0
	for _, id := range parties {
		rows := DimensionProfile(ds, id)
		coded, multi, _ := PartyTotals(rows)
		if coded == 0 {
			t.Errorf("party %q has no coded statement in any dimension", id)
		}
		if multi > coded {
			t.Errorf("party %q: %d cells support a trend but only %d are coded", id, multi, coded)
		}
		sumStatements := 0
		for _, r := range rows {
			sumStatements += r.Statements
			if r.Mean < -2 || r.Mean > 2 {
				t.Errorf("party %q dimension %q: mean %v outside -2..+2", id, r.Dimension, r.Mean)
			}
			if r.Coded > r.Statements {
				t.Errorf("party %q dimension %q: %d coded of %d statements", id, r.Dimension, r.Coded, r.Statements)
			}
		}
		if len(rows) > 0 {
			if totalStatements == 0 {
				totalStatements = sumStatements
			} else if sumStatements != totalStatements {
				t.Errorf("party %q: %d statements across dimensions, other parties have %d",
					id, sumStatements, totalStatements)
			}
		}
	}
	t.Logf("%d parties, %d statements each across all dimensions", len(parties), totalStatements)
}

// The per-cell marker answers "can this cell say anything about a year", so the
// rule that decides it is tested on its own: two distinct dates yes, two rows from
// the same day no, and another party's evidence never.
func TestMultiDateStatementsCountsDistinctDatesOnly(t *testing.T) {
	ds := &data.Dataset{
		Evidence: []data.Evidence{
			// Two different dates: a real second observation.
			{StatementID: "s1", PartyID: "ap", Date: "2020-05-01", Position: pos(2)},
			{StatementID: "s1", PartyID: "ap", Date: "2026-01-01", Position: pos(2)},
			// Same day twice: one observation, not a trend.
			{StatementID: "s2", PartyID: "ap", Date: "2025-06-05", Position: pos(1)},
			{StatementID: "s2", PartyID: "ap", Date: "2025-06-05", Position: pos(1)},
			// Undated rows cannot make a cell dated.
			{StatementID: "s3", PartyID: "ap", Date: "", Position: pos(1)},
			{StatementID: "s3", PartyID: "ap", Date: "ukjent", Position: pos(1)},
			// No coded position: not an observation of anything.
			{StatementID: "s4", PartyID: "ap", Date: "2020-01-01"},
			{StatementID: "s4", PartyID: "ap", Date: "2026-01-01"},
			// Another party must not mark this party's cells.
			{StatementID: "s2", PartyID: "h", Date: "1999-01-01", Position: pos(-2)},
			{StatementID: "s2", PartyID: "h", Date: "2024-01-01", Position: pos(-2)},
		},
	}
	got := MultiDateStatements(ds, "ap")
	if !got["s1"] {
		t.Error("s1 has two distinct dates and must be marked")
	}
	if got["s2"] {
		t.Error("s2 has one distinct date twice and must not be marked")
	}
	if got["s3"] {
		t.Error("s3 has no parseable date and must not be marked")
	}
	if got["s4"] {
		t.Error("s4 has no coded position and must not be marked")
	}
	if len(got) != 1 {
		t.Errorf("marked %d statements, want exactly 1: %v", len(got), got)
	}
	// And the other party sees its own answer, not ap's.
	if h := MultiDateStatements(ds, "h"); !h["s2"] {
		t.Error("h has two dates on s2 and must be marked for h")
	}
}
