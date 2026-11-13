package server

import (
	"log/slog"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
	"github.com/jkorsvik/ektevalgomat/internal/scoring"
)

func getRommet(t *testing.T) string {
	t.Helper()
	return getRommetQuery(t, "")
}

func getRommetQuery(t *testing.T, query string) string {
	t.Helper()
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	target := "/rommet"
	if query != "" {
		target += "?" + query
	}
	h.ServeHTTP(w, httptest.NewRequest("GET", target, nil))
	if w.Code != 200 {
		t.Fatalf("%s answered %d, want 200", target, w.Code)
	}
	return w.Body.String()
}

func romStatementByNumber(t *testing.T, number int) data.Statement {
	t.Helper()
	for _, statement := range data.Store.Statements {
		if statement.Number == number {
			return statement
		}
	}
	t.Fatalf("statement %d not found", number)
	return data.Statement{}
}

func TestRommetHasConcreteStatementsAndNo3D(t *testing.T) {
	body := getRommet(t)
	for _, want := range []string{
		"<h1>Hva skiller partiene?</h1>",
		`id="rom-topic"`, `id="rom-party-a"`, `id="rom-party-b"`,
		`action="/rommet#saker"`, `class="rom-statement"`,
		"Sterkt mot", "Sterkt for", "Se hele saken",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the concrete comparison is missing %q", want)
		}
	}
	for _, unwanted := range []string{
		"rommet-canvas", "rommet-data", "/static/js/rom.js", "rommet-zoom",
		"rommet-springs", "Partiene som volum", "Hva aksene består av",
	} {
		if strings.Contains(body, unwanted) {
			t.Errorf("the retired 3D experience is still present: %q", unwanted)
		}
	}
}

func TestRommetDefaultShowsSixRealDisagreements(t *testing.T) {
	view := buildRommetView(data.Store, url.Values{})
	if len(view.Statements) != 6 {
		t.Fatalf("default shows %d statements, want a readable first set of 6", len(view.Statements))
	}
	if view.TotalMatches < len(view.Statements) {
		t.Fatalf("total %d is below the %d displayed statements", view.TotalMatches, len(view.Statements))
	}
	for _, item := range view.Statements {
		statement := romStatementByNumber(t, item.Number)
		lo, hi, coded := 0, 0, 0
		for _, party := range data.Store.Parties {
			position := scoring.ComputeCell(data.Store, statement.ID, party.ID).PartyPosition
			if position == nil {
				continue
			}
			if coded == 0 || *position < lo {
				lo = *position
			}
			if coded == 0 || *position > hi {
				hi = *position
			}
			coded++
		}
		if coded != item.Coded {
			t.Errorf("utsagn %d reports %d coded parties, recomputed %d", item.Number, item.Coded, coded)
		}
		if coded < 2 || hi == lo {
			t.Errorf("utsagn %d is not a real disagreement: lo=%d hi=%d coded=%d", item.Number, lo, hi, coded)
		}
	}
}

func TestRommetEveryPartyChipMatchesItsCellAndEvidenceLink(t *testing.T) {
	view := buildRommetView(data.Store, url.Values{"alle": []string{"1"}})
	if len(view.Statements) != view.TotalMatches || len(view.Statements) <= 6 {
		t.Fatalf("all view has %d of %d statements; the full list is not actually expanded", len(view.Statements), view.TotalMatches)
	}
	labelValue := map[string]int{}
	for _, definition := range romPositionLabels {
		labelValue[definition.label] = definition.value
	}
	for _, item := range view.Statements {
		statement := romStatementByNumber(t, item.Number)
		if len(item.Groups) != 5 {
			t.Fatalf("utsagn %d has %d position groups, want 5", item.Number, len(item.Groups))
		}
		coded := 0
		for _, group := range item.Groups {
			wantPosition := labelValue[group.Label]
			for _, chip := range group.Parties {
				coded++
				cell := scoring.ComputeCell(data.Store, statement.ID, chip.ID)
				if cell.PartyPosition == nil || *cell.PartyPosition != wantPosition {
					t.Errorf("utsagn %d puts %s under %q, but its cell says %v", item.Number, chip.ID, group.Label, cell.PartyPosition)
				}
				wantHref := "/utsagn/" + strconv.Itoa(item.Number) + "?alle=1&fra=rommet#party-" + chip.ID
				if chip.Href != wantHref {
					t.Errorf("utsagn %d %s links to %q, want %q", item.Number, chip.ID, chip.Href, wantHref)
				}
				if chip.Selected {
					t.Errorf("utsagn %d selects %s without a party-pair query", item.Number, chip.ID)
				}
			}
		}
		if coded+len(item.Undocumented) != len(data.Store.Parties) {
			t.Errorf("utsagn %d accounts for %d+%d parties, want %d", item.Number, coded, len(item.Undocumented), len(data.Store.Parties))
		}
	}
}

func TestRommetTopicFilterUsesRealDimensions(t *testing.T) {
	view := buildRommetView(data.Store, url.Values{"tema": []string{"universell-grunnmur"}})
	if view.TopicName != "Universell grunnmur" {
		t.Fatalf("topic name = %q", view.TopicName)
	}
	if len(view.Statements) == 0 || len(view.Statements) != view.TotalMatches {
		t.Fatalf("topic shows %d of %d statements", len(view.Statements), view.TotalMatches)
	}
	for _, item := range view.Statements {
		if item.Dimension != "Universell grunnmur" {
			t.Errorf("topic leaked in %q", item.Dimension)
		}
	}
	empty := buildRommetView(data.Store, url.Values{"tema": []string{"not-a-topic"}})
	if len(empty.Statements) != 0 || empty.TotalMatches != 0 {
		t.Error("an unknown topic invented results")
	}
}

func TestRommetPartyPairShowsOnlyDifferentDocumentedAnswers(t *testing.T) {
	view := buildRommetView(data.Store, url.Values{"a": []string{"ap"}, "b": []string{"h"}})
	if !view.Comparing || len(view.Statements) == 0 {
		t.Fatal("Ap and H should have concrete disagreements to compare")
	}
	if len(view.Statements) != view.TotalMatches {
		t.Fatalf("pair view truncates the result: %d of %d", len(view.Statements), view.TotalMatches)
	}
	for _, item := range view.Statements {
		statement := romStatementByNumber(t, item.Number)
		a := scoring.ComputeCell(data.Store, statement.ID, "ap").PartyPosition
		b := scoring.ComputeCell(data.Store, statement.ID, "h").PartyPosition
		if a == nil || b == nil || *a == *b {
			t.Errorf("pair view includes utsagn %d where Ap=%v and H=%v", item.Number, a, b)
		}
		if item.PairGap == "" {
			t.Errorf("pair view does not explain the gap on utsagn %d", item.Number)
		}
		for _, group := range item.Groups {
			for _, chip := range group.Parties {
				wantSelected := chip.ID == "ap" || chip.ID == "h"
				if chip.Selected != wantSelected {
					t.Errorf("utsagn %d chip %s selected=%v, want %v", item.Number, chip.ID, chip.Selected, wantSelected)
				}
			}
		}
	}
}

func TestRommetRejectsTheSamePartyTwiceWithoutInventingAComparison(t *testing.T) {
	view := buildRommetView(data.Store, url.Values{"a": []string{"ap"}, "b": []string{"ap"}})
	if view.Comparing {
		t.Error("Ap versus Ap must not become a comparison")
	}
	if view.Notice == "" {
		t.Error("the reader gets no instruction after choosing the same party twice")
	}
}

func TestRommetRenderedPairPageKeepsTheMethodAndSourceGap(t *testing.T) {
	body := getRommetQuery(t, "tema=universell-grunnmur&a=ap&b=h")
	for _, want := range []string{
		`value="ap" selected`, `value="h" selected`,
		"steg mellom", "Slik leser du posisjonene",
		"Spørsmålet:", "Kilden:", "Det som mangler:", `href="/metode"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("filtered page is missing %q", want)
		}
	}
}
