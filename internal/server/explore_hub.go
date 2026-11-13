package server

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/data"
	"github.com/jkorsvik/ektevalgomat/internal/scoring"
)

// exploreTopic is one server-side topic filter. Slug is shared with the rest
// of the site, so a topic URL is stable and can be linked without JavaScript.
type exploreTopic struct {
	Name   string
	Slug   string
	Href   string
	Count  int
	Active bool
}

// exploreStatement is one search result. Coded is counted from scoring cells,
// not from evidence rows: a source without a coded position is not a position.
type exploreStatement struct {
	Number    int
	ShortName string
	Text      string
	Dimension string
	Href      string
	Coded     int
	Total     int
}

// exploreHubGroup is one reader question and the pages that help answer it.
type exploreHubGroup struct {
	ID          string
	Title       string
	Description string
	Steps       []exploreStep
}

type exploreHubView struct {
	Meta            pageMeta
	Query           string
	Topic           string
	TopicName       string
	Filtered        bool
	Topics          []exploreTopic
	Statements      []exploreStatement
	ResultCount     int
	TotalStatements int
	TotalParties    int
	Groups          []exploreHubGroup
}

var exploreGroupDefinitions = []struct {
	ID          string
	Title       string
	Description string
	Hrefs       []string
}{
	{
		ID:          "posisjoner",
		Title:       "Hva mener partiene?",
		Description: "Sammenlign dokumenterte posisjoner, se hvor partiene skiller seg, og kontroller hvor langt tilbake kildene rekker.",
		Hrefs:       []string{"/kart", "/rommet", "/tidslinje"},
	},
	{
		ID:          "handling",
		Title:       "Hva gjør de, og hvem bestemmer?",
		Description: "Se hvem som stemmer sammen, hvilke verv representantene har, og hvor den formelle makten sitter.",
		Hrefs:       []string{"/stemmer", "/personer", "/makt"},
	},
	{
		ID:          "ressurser",
		Title:       "Hva koster det, og hvem eier?",
		Description: "Følg budsjettene, partienes finansiering og statens eierposter. Tallene sier ikke det samme, og sidene forklarer forskjellen.",
		Hrefs:       []string{"/budsjett", "/penger", "/eierskap"},
	},
}

// buildExploreGroups keeps the nine analysis pages, but groups them by reader
// intent instead of asking everyone to walk through all nine in one order.
func buildExploreGroups() []exploreHubGroup {
	groups := make([]exploreHubGroup, 0, len(exploreGroupDefinitions))
	for _, definition := range exploreGroupDefinitions {
		group := exploreHubGroup{
			ID: definition.ID, Title: definition.Title, Description: definition.Description,
		}
		for _, href := range definition.Hrefs {
			group.Steps = append(group.Steps, exploreInfo(href))
		}
		groups = append(groups, group)
	}
	return groups
}

// exploreMenuGroups is the template-facing form used by the global menu.
func exploreMenuGroups() []exploreHubGroup { return buildExploreGroups() }

func buildExploreHub(ds *data.Dataset, rawQuery, rawTopic string) exploreHubView {
	query := strings.TrimSpace(rawQuery)
	topic := strings.TrimSpace(rawTopic)
	view := exploreHubView{
		Meta: meta("Utforsk sakene og kildene · Ekte valgomat",
			"Søk i utsagnene, sammenlign partienes dokumenterte posisjoner, og åpne kildene bak svarene.",
			"/utforsk"),
		Query: query, Topic: topic, Filtered: query != "" || topic != "",
		TotalStatements: len(ds.Statements), TotalParties: len(ds.Parties),
		Groups: buildExploreGroups(),
	}

	// Keep the site's established topic order, then include any real dimension
	// outside it (currently the addendum) instead of silently dropping statements.
	counts := map[string]int{}
	for _, statement := range ds.Statements {
		counts[statement.Dimension]++
	}
	orderedDimensions := append([]string{}, dimensionOrder...)
	seen := map[string]bool{}
	for _, dimension := range orderedDimensions {
		seen[dimension] = true
	}
	for _, statement := range ds.Statements {
		if !seen[statement.Dimension] {
			seen[statement.Dimension] = true
			orderedDimensions = append(orderedDimensions, statement.Dimension)
		}
	}
	for _, dimension := range orderedDimensions {
		if counts[dimension] == 0 {
			continue
		}
		slug := slugify(dimension)
		href := "/utforsk?tema=" + url.QueryEscape(slug) + "#saker"
		active := slug == topic
		if active {
			view.TopicName = dimension
		}
		view.Topics = append(view.Topics, exploreTopic{
			Name: dimension, Slug: slug, Href: href, Count: counts[dimension], Active: active,
		})
	}
	if topic != "" && view.TopicName == "" {
		view.TopicName = "Ukjent tema"
	}

	for _, statement := range ds.Statements {
		if topic != "" && slugify(statement.Dimension) != topic {
			continue
		}
		if !statementMatchesQuery(statement, query) {
			continue
		}
		coded := 0
		for _, party := range ds.Parties {
			if scoring.ComputeCell(ds, statement.ID, party.ID).PartyPosition != nil {
				coded++
			}
		}
		href := "/utsagn/" + strconv.Itoa(statement.Number)
		returnQuery := url.Values{}
		if query != "" {
			returnQuery.Set("q", query)
		}
		if topic != "" {
			returnQuery.Set("tema", topic)
		}
		if len(returnQuery) > 0 {
			href += "?" + returnQuery.Encode()
		}
		view.Statements = append(view.Statements, exploreStatement{
			Number: statement.Number, ShortName: statement.ShortName, Text: statement.Text,
			Dimension: statement.Dimension, Href: href,
			Coded: coded, Total: len(ds.Parties),
		})
	}
	view.ResultCount = len(view.Statements)
	return view
}

// statementMatchesQuery uses AND between words. A search for "tannhelse 2030"
// must narrow the list, not return every statement that contains either word.
func statementMatchesQuery(statement data.Statement, query string) bool {
	if strings.TrimSpace(query) == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		strconv.Itoa(statement.Number), statement.ShortName, statement.Text,
		statement.Dimension, statement.PrimaryQuestion,
	}, " "))
	for _, word := range strings.Fields(strings.ToLower(query)) {
		if !strings.Contains(haystack, word) {
			return false
		}
	}
	return true
}

func (v exploreHubView) ResultSummary() string {
	if v.ResultCount == 0 {
		return "Ingen utsagn passer til søket."
	}
	if v.Filtered {
		return fmt.Sprintf("Viser %d av %d utsagn.", v.ResultCount, v.TotalStatements)
	}
	return fmt.Sprintf("Alle %d utsagn. Velg ett for å se partiposisjonene og kildene.", v.TotalStatements)
}
