package analysis

import (
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The year grid is the party x dimension x year view, so every rule that decides
// what lands in a cell is tested on its own. A cell that counts a duplicate as two
// observations, or spreads one observation across years, would read as more
// knowledge than the dataset holds.
func TestYearGridCountsOneObservationPerStatementPerYear(t *testing.T) {
	ds := &data.Dataset{
		Statements: []data.Statement{
			{ID: "s1", Dimension: "Miljø"},
			{ID: "s2", Dimension: "Miljø"},
			{ID: "s3", Dimension: "Skatt"},
		},
		Evidence: []data.Evidence{
			// One statement seen in two years: real in both years.
			{StatementID: "s1", PartyID: "ap", Date: "2021-05-01", Position: pos(2)},
			{StatementID: "s1", PartyID: "ap", Date: "2025-05-01", Position: pos(2)},
			// Two rows from the same year: one observation, not two.
			{StatementID: "s2", PartyID: "ap", Date: "2024-03-01", Position: pos(-1)},
			{StatementID: "s2", PartyID: "ap", Date: "2024-09-01", Position: pos(-1)},
			// No parseable year: counted apart, never assigned to a year.
			{StatementID: "s3", PartyID: "ap", Date: "ukjent", Position: pos(1)},
			// A second dimension, so the grid has two rows.
			{StatementID: "s3", PartyID: "h", Date: "2023-01-01", Position: pos(-2)},
		},
	}

	rows, years := YearGrid(ds, []string{"ap", "h"})

	if len(years) != 4 || years[0] != 2021 || years[1] != 2023 || years[2] != 2024 || years[3] != 2025 {
		t.Fatalf("years = %v, want [2021 2023 2024 2025] ascending", years)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d dimension rows, want 2 (Miljø and Skatt)", len(rows))
	}

	miljo := rows[0]
	if miljo.Dimension != "Miljø" || miljo.Statements != 2 {
		t.Fatalf("first row = %+v, want Miljø with 2 statements", miljo)
	}
	if len(miljo.Parties) != 1 || miljo.Parties[0].Party != "ap" {
		t.Fatalf("Miljø parties = %+v, want only ap", miljo.Parties)
	}
	ap := miljo.Parties[0]
	if got := ap.Years[2021]; got.Positions != 1 || got.Sum != 2 {
		t.Errorf("2021 = %+v, want 1 position summing to 2", got)
	}
	if got := ap.Years[2025]; got.Positions != 1 || got.Sum != 2 {
		t.Errorf("2025 = %+v, want 1 position summing to 2", got)
	}
	if got := ap.Years[2024]; got.Positions != 1 || got.Sum != -1 {
		t.Errorf("2024 = %+v, want 1 position (the same-day duplicate must not count twice)", got)
	}
	if got, want := ap.Years[2024].Mean(), -1.0; got != want {
		t.Errorf("2024 mean = %v, want %v", got, want)
	}

	skatt := rows[1]
	if skatt.Dimension != "Skatt" {
		t.Fatalf("second row = %+v, want Skatt", skatt)
	}
	if len(skatt.Parties) != 2 {
		t.Fatalf("Skatt has %d parties, want 2 (ap undated and h dated)", len(skatt.Parties))
	}
	var apSkatt PartyYears
	for _, p := range skatt.Parties {
		if p.Party == "ap" {
			apSkatt = p
		}
	}
	if apSkatt.Undated != 1 {
		t.Errorf("ap undated in Skatt = %d, want 1", apSkatt.Undated)
	}
	if len(apSkatt.Years) != 0 {
		t.Errorf("ap undated row leaks into years: %v", apSkatt.Years)
	}
	if apSkatt.Total() != 1 {
		t.Errorf("ap total in Skatt = %d, want 1", apSkatt.Total())
	}
}

// A dimension nobody coded must not produce a row: an empty row reads as "we
// looked and found nothing", which is a different claim from "no one has looked".
func TestYearGridOmitsDimensionsAndPartiesWithNoCodedPosition(t *testing.T) {
	ds := &data.Dataset{
		Statements: []data.Statement{
			{ID: "s1", Dimension: "Miljø"},
			{ID: "s2", Dimension: "Tomt"},
		},
		Evidence: []data.Evidence{
			{StatementID: "s1", PartyID: "ap", Date: "2024-01-01", Position: pos(1)},
			// A source with no position is not an observation.
			{StatementID: "s1", PartyID: "h", Date: "2024-01-01"},
		},
	}
	rows, _ := YearGrid(ds, []string{"ap", "h"})
	if len(rows) != 1 || rows[0].Dimension != "Miljø" {
		t.Fatalf("rows = %+v, want only Miljø", rows)
	}
	if len(rows[0].Parties) != 1 || rows[0].Parties[0].Party != "ap" {
		t.Errorf("parties = %+v, want only ap (h has no coded position)", rows[0].Parties)
	}
}

// A party the caller did not ask for must not appear, so a page cannot show a
// party its own filter excludes.
func TestYearGridRespectsThePartyFilter(t *testing.T) {
	ds := &data.Dataset{
		Statements: []data.Statement{{ID: "s1", Dimension: "Miljø"}},
		Evidence: []data.Evidence{
			{StatementID: "s1", PartyID: "ap", Date: "2024-01-01", Position: pos(1)},
			{StatementID: "s1", PartyID: "h", Date: "2024-01-01", Position: pos(-1)},
		},
	}
	rows, _ := YearGrid(ds, []string{"ap"})
	if len(rows) != 1 || len(rows[0].Parties) != 1 || rows[0].Parties[0].Party != "ap" {
		t.Fatalf("rows = %+v, want only ap", rows)
	}
	if rows[0].Parties[0].Years[2024].Sum != 1 {
		t.Errorf("sum = %d, want 1 (h's -1 must not be added)", rows[0].Parties[0].Years[2024].Sum)
	}
}
