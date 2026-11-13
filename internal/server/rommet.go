package server

import (
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/data"
	"github.com/jkorsvik/ektevalgomat/internal/scoring"
)

// /rommet used to project twelve dimensions into a rotating three-dimensional
// space. The geometry was correct, but it made readers interpret a statistical
// model before they could answer a practical question. The page now compares
// concrete statements instead. No averaging, projection, canvas or JavaScript:
// every visible position is scoring.ComputeCell's newest coded evidence row.

type romTopic struct {
	Name   string
	Slug   string
	Active bool
}

type romPartyOption struct {
	ID        string
	Short     string
	Name      string
	SelectedA bool
	SelectedB bool
}

type romPartyChip struct {
	ID        string
	Short     string
	Name      string
	Href      string
	AriaLabel string
	Selected  bool
}

type romPositionGroup struct {
	Label   string
	Parties []romPartyChip
}

type romStatement struct {
	Number       int
	ShortName    string
	Text         string
	Dimension    string
	Href         string
	Coded        int
	Total        int
	Groups       []romPositionGroup
	Undocumented []romPartyChip
	PairGap      string
}

type rommetView struct {
	Meta         pageMeta
	SnapshotDate string
	Topics       []romTopic
	Parties      []romPartyOption
	Statements   []romStatement
	TopicName    string
	PartyAName   string
	PartyBName   string
	Comparing    bool
	Filtered     bool
	ShowAll      bool
	AllHref      string
	Summary      string
	Notice       string
	TotalMatches int
}

type romCandidate struct {
	view     romStatement
	spread   int
	variance float64
	pairGap  int
}

var romPositionLabels = []struct {
	value int
	label string
}{
	{-2, "Sterkt mot"},
	{-1, "Mot"},
	{0, "Nøytral / delt"},
	{1, "For"},
	{2, "Sterkt for"},
}

func buildRommetView(ds *data.Dataset, query url.Values) rommetView {
	topic := strings.TrimSpace(query.Get("tema"))
	showAll := query.Get("alle") == "1"
	partyA, partyB := query.Get("a"), query.Get("b")
	partyByID := map[string]data.Party{}
	for _, party := range ds.Parties {
		partyByID[party.ID] = party
	}
	if _, ok := partyByID[partyA]; !ok {
		partyA = ""
	}
	if _, ok := partyByID[partyB]; !ok {
		partyB = ""
	}
	notice := ""
	if partyA != "" && partyA == partyB {
		partyB = ""
		notice = "Velg to ulike partier for å sammenligne dem."
	}
	comparing := partyA != "" && partyB != ""

	view := rommetView{
		Meta: meta("Uenigheter · Ekte valgomat",
			"Se konkrete saker der partiene har ulike dokumenterte posisjoner. Filtrer på tema eller sammenlign to partier, og åpne kildene bak hvert svar.",
			"/rommet"),
		SnapshotDate: SnapshotDate,
		TopicName:    "Ukjent tema",
		Comparing:    comparing,
		Filtered:     topic != "" || comparing || showAll,
		ShowAll:      showAll,
		AllHref:      "/rommet?alle=1#saker",
		Notice:       notice,
	}
	if comparing {
		view.PartyAName = shortPartyName(partyByID[partyA])
		view.PartyBName = shortPartyName(partyByID[partyB])
	}

	for _, party := range ds.Parties {
		view.Parties = append(view.Parties, romPartyOption{
			ID: party.ID, Short: shortPartyName(party), Name: party.Name,
			SelectedA: party.ID == partyA, SelectedB: party.ID == partyB,
		})
	}

	for _, dimension := range orderedRomTopics(ds) {
		slug := slugify(dimension)
		active := slug == topic
		if active {
			view.TopicName = dimension
		}
		view.Topics = append(view.Topics, romTopic{Name: dimension, Slug: slug, Active: active})
	}
	if topic == "" {
		view.TopicName = ""
	}

	candidates := make([]romCandidate, 0, len(ds.Statements))
	for _, statement := range ds.Statements {
		if topic != "" && slugify(statement.Dimension) != topic {
			continue
		}

		returnQuery := url.Values{"fra": []string{"rommet"}}
		if topic != "" {
			returnQuery.Set("tema", topic)
		}
		if comparing {
			returnQuery.Set("a", partyA)
			returnQuery.Set("b", partyB)
		}
		if showAll {
			returnQuery.Set("alle", "1")
		}
		statementHref := "/utsagn/" + strconv.Itoa(statement.Number) + "?" + returnQuery.Encode()

		positions := map[string]*int{}
		groups := map[int][]romPartyChip{}
		var undocumented []romPartyChip
		coded := 0
		lo, hi := 0, 0
		var values []float64
		for _, party := range ds.Parties {
			cell := scoring.ComputeCell(ds, statement.ID, party.ID)
			chip := romPartyChip{
				ID: party.ID, Short: shortPartyName(party), Name: party.Name,
				Href:     statementHref + "#party-" + party.ID,
				Selected: comparing && (party.ID == partyA || party.ID == partyB),
			}
			if cell.PartyPosition == nil {
				chip.AriaLabel = party.Name + ": ikke vurdert. Se saken."
				undocumented = append(undocumented, chip)
				continue
			}
			position := *cell.PartyPosition
			positions[party.ID] = cell.PartyPosition
			chip.AriaLabel = fmt.Sprintf("%s: %s. Se kilden for %s.",
				party.Name, strings.ToLower(scoring.PositionWord(cell.PartyPosition)), statement.ShortName)
			groups[position] = append(groups[position], chip)
			values = append(values, float64(position))
			if coded == 0 || position < lo {
				lo = position
			}
			if coded == 0 || position > hi {
				hi = position
			}
			coded++
		}
		spread := hi - lo
		if coded < 2 || spread == 0 {
			continue
		}

		pairGap := 0
		if comparing {
			a, b := positions[partyA], positions[partyB]
			if a == nil || b == nil || *a == *b {
				continue
			}
			pairGap = *a - *b
			if pairGap < 0 {
				pairGap = -pairGap
			}
		}

		item := romStatement{
			Number: statement.Number, ShortName: statement.ShortName, Text: statement.Text,
			Dimension: statement.Dimension,
			Href:      statementHref,
			Coded:     coded, Total: len(ds.Parties), Undocumented: undocumented,
		}
		for _, definition := range romPositionLabels {
			item.Groups = append(item.Groups, romPositionGroup{
				Label: definition.label, Parties: groups[definition.value],
			})
		}
		if comparing {
			item.PairGap = fmt.Sprintf("%d steg mellom %s og %s", pairGap, view.PartyAName, view.PartyBName)
		}
		candidates = append(candidates, romCandidate{
			view: item, spread: spread, variance: romVariance(values), pairGap: pairGap,
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if comparing && a.pairGap != b.pairGap {
			return a.pairGap > b.pairGap
		}
		if a.spread != b.spread {
			return a.spread > b.spread
		}
		if math.Abs(a.variance-b.variance) > 1e-12 {
			return a.variance > b.variance
		}
		if a.view.Coded != b.view.Coded {
			return a.view.Coded > b.view.Coded
		}
		return a.view.Number < b.view.Number
	})

	view.TotalMatches = len(candidates)
	if !view.Filtered && len(candidates) > 6 {
		candidates = candidates[:6]
	}
	for _, candidate := range candidates {
		view.Statements = append(view.Statements, candidate.view)
	}
	view.Summary = romResultSummary(view, len(candidates))
	return view
}

func orderedRomTopics(ds *data.Dataset) []string {
	counts := map[string]int{}
	for _, statement := range ds.Statements {
		counts[statement.Dimension]++
	}
	ordered := append([]string{}, dimensionOrder...)
	seen := map[string]bool{}
	for _, dimension := range ordered {
		seen[dimension] = true
	}
	for _, statement := range ds.Statements {
		if !seen[statement.Dimension] {
			seen[statement.Dimension] = true
			ordered = append(ordered, statement.Dimension)
		}
	}
	out := ordered[:0]
	for _, dimension := range ordered {
		if counts[dimension] > 0 {
			out = append(out, dimension)
		}
	}
	return out
}

func shortPartyName(party data.Party) string {
	if party.Short != "" {
		return party.Short
	}
	return strings.ToUpper(party.ID)
}

func romVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))
	variance := 0.0
	for _, value := range values {
		variance += (value - mean) * (value - mean)
	}
	return variance / float64(len(values))
}

func romResultSummary(view rommetView, shown int) string {
	if view.TotalMatches == 0 {
		if view.Comparing {
			return fmt.Sprintf("%s og %s har ingen dokumenterte uenigheter i dette utvalget.", view.PartyAName, view.PartyBName)
		}
		return "Ingen saker med ulike dokumenterte posisjoner passer til dette valget."
	}
	noun := "saker"
	if view.TotalMatches == 1 {
		noun = "sak"
	}
	switch {
	case view.Comparing && view.TopicName != "":
		return fmt.Sprintf("%s og %s er uenige i %d %s innen %s.", view.PartyAName, view.PartyBName, view.TotalMatches, noun, view.TopicName)
	case view.Comparing:
		return fmt.Sprintf("%s og %s er uenige i %d %s.", view.PartyAName, view.PartyBName, view.TotalMatches, noun)
	case view.TopicName != "":
		return fmt.Sprintf("%d %s med ulike posisjoner innen %s.", view.TotalMatches, noun, view.TopicName)
	case view.ShowAll:
		return fmt.Sprintf("%d %s der partiene har ulike dokumenterte posisjoner.", view.TotalMatches, noun)
	default:
		return fmt.Sprintf("De %d sakene med størst uenighet, av %d mulige.", shown, view.TotalMatches)
	}
}
