package server

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
	"github.com/jkorsvik/ektevalgomat/internal/scoring"
)

func TestExploreHubContainsEveryStatementAndCoverageMatchesScoring(t *testing.T) {
	view := buildExploreHub(data.Store, "", "")
	if view.ResultCount != len(data.Store.Statements) {
		t.Fatalf("hub shows %d statements, dataset has %d", view.ResultCount, len(data.Store.Statements))
	}
	if len(view.Statements) < 50 {
		t.Fatal("hub loaded too few statements to prove the index works")
	}
	for _, row := range view.Statements {
		statement, ok := data.Store.Statement(fmt.Sprintf("utsagn-%02d", row.Number))
		if !ok {
			t.Fatalf("hub returned unknown statement %d", row.Number)
		}
		coded := 0
		for _, party := range data.Store.Parties {
			if scoring.ComputeCell(data.Store, statement.ID, party.ID).PartyPosition != nil {
				coded++
			}
		}
		if row.Coded != coded {
			t.Errorf("%s: hub says %d coded parties, scoring says %d", statement.ID, row.Coded, coded)
		}
	}
}

func TestExploreHubSearchNarrowsByAllWords(t *testing.T) {
	statement := data.Statement{Number: 25, Text: "Tannhelse skal utvides innen 2030", Dimension: "Helse"}
	if !statementMatchesQuery(statement, "tannhelse 2030") {
		t.Error("two words present in the statement do not match")
	}
	if statementMatchesQuery(statement, "tannhelse skatt") {
		t.Error("a query matched when one of its words was absent")
	}

	view := buildExploreHub(data.Store, "offentlig lobbyregister", "")
	found := false
	for _, row := range view.Statements {
		if row.Number == 1 {
			found = true
		}
	}
	if !found {
		t.Error("searching for the lobby-register statement did not find utsagn 1")
	}
	if view.ResultCount >= len(data.Store.Statements) {
		t.Error("the search did not narrow the list")
	}
}

func TestExploreHubTopicFilterOnlyReturnsThatTopic(t *testing.T) {
	view := buildExploreHub(data.Store, "", "universell-grunnmur")
	if view.ResultCount == 0 {
		t.Fatal("a known topic returned no statements")
	}
	for _, row := range view.Statements {
		if row.Dimension != "Universell grunnmur" {
			t.Errorf("topic filter leaked %q into Universell grunnmur", row.Dimension)
		}
	}
	if view.TopicName != "Universell grunnmur" {
		t.Errorf("active topic is %q", view.TopicName)
	}
}

func TestExploreHubRendersServerSideSearchWithoutJavaScript(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/utforsk?q=tannhelse&tema=universell-grunnmur")
	for _, want := range []string{
		`action="/utforsk#saker"`,
		`value="tannhelse"`,
		`value="universell-grunnmur" selected`,
		`href="/utsagn/25?q=tannhelse&amp;tema=universell-grunnmur"`,
		"partier har en posisjon med kilde",
		"Hva vil du finne ut?",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered hub is missing %q", want)
		}
	}
	if strings.Contains(body, "<script") {
		t.Error("the statement finder depends on JavaScript")
	}
}

func TestExploreHubGroupsReachEveryAnalysisPage(t *testing.T) {
	groups := buildExploreGroups()
	if len(groups) != 3 {
		t.Fatalf("got %d reader groups, want 3", len(groups))
	}
	reached := map[string]bool{}
	for _, group := range groups {
		if group.Title == "" || group.Description == "" {
			t.Errorf("empty group heading or explanation: %+v", group)
		}
		for _, step := range group.Steps {
			if reached[step.Href] {
				t.Errorf("%s appears in two groups", step.Href)
			}
			reached[step.Href] = true
		}
	}
	for _, step := range explorePages {
		if !reached[step.Href] {
			t.Errorf("%s is not reachable from a hub group", step.Href)
		}
	}
}

func TestFrontOffersQuizAndIndependentExploration(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/")
	for _, want := range []string{
		"Start valgomaten",
		"hurtigmodus med 12 nøkkelutsagn",
		`href="/utforsk"`,
		"Utforsk sakene og kildene",
		"Ingen matchscore",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("front page is missing %q", want)
		}
	}
}

func TestExploreHubUnknownSearchHasAnHonestEmptyState(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/utforsk?q=detfinnesikkei_nogetutsagn")
	if !strings.Contains(body, "Ingen utsagn passer til søket.") {
		t.Error("empty search does not say that nothing matched")
	}
	if strings.Contains(body, "statement-result-coverage") {
		t.Error("empty search rendered a result anyway")
	}
}
