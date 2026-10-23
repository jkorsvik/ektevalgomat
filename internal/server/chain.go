package server

import (
	"sort"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The evidence chain (2026-10-13). The design brief calls a
// claim-to-outcome stepper its signature component: what a party promised,
// what was proposed, how it voted, what was decided, and what followed.
//
// The chain is derived from evidence_type, which every observation already
// carries, so it needs no new data. It is shown per statement: for THIS
// question, which steps in the chain do we actually have a source for, and
// which are empty. The empty steps are the point. A chain that showed only
// the steps with evidence would imply the others did not happen.
//
// What it deliberately does NOT claim: that the steps connect to one
// another. A proposal on this statement and a vote on this statement are
// both real, but nothing here asserts the vote was on that proposal. That
// join is the one the programme's "statement:" field was meant to carry,
// and it is empty everywhere today (0 of 316 points name a statement), so
// the chain reports presence, not causation.

type chainStage struct {
	Key     string
	Label   string
	Meaning string
	Count   int
	Parties []string // party shorts seen at this step, in site order
}

// chainStageOrder is the chain, in the order a claim would become action.
var chainStageOrder = []struct {
	Key, Label, Meaning string
	Types               []string
}{
	{"lofte", "Løfte", "Partiet sier det i sitt eget program", []string{"partiprogram"}},
	{"forslag", "Forslag", "Noen la det fram", []string{"representantforslag", "budsjettsforslag", "budsjettforslag", "proposisjon", "melding_til_stortinget"}},
	{"stemme", "Stemme", "Noen stemte", []string{"voteringsresultat", "voteringsspor"}},
	{"vedtak", "Vedtak", "Noe ble bestemt", []string{"innstilling", "vedtak", "lovendring", "mandat", "utredning"}},
	{"utfall", "Utfall", "Hva som fulgte", []string{"effekt", "ffi_monitorering", "ffi_attribusjon", "rettsstatus"}},
}

// chainContextTypes are observed but sit outside the chain: context, not a
// step. They are named here so a new type cannot be added without a
// decision — TestEveryEvidenceTypeIsPlaced fails on one that is in neither
// list.
var chainContextTypes = []string{
	"kontekst", "faktakontekst", "plenumsinnlegg", "komitemerknad",
	"presse", "presseuttalelse", "nettverksanalyse", "attribusjon",
}

// chainTypeToStage indexes the map once at startup.
var chainTypeToStage = func() map[string]string {
	m := map[string]string{}
	for _, s := range chainStageOrder {
		for _, t := range s.Types {
			m[t] = s.Key
		}
	}
	return m
}()

// buildChain groups one statement's evidence into the chain. Every stage is
// returned, including empty ones, in chain order.
func buildChain(ds *data.Dataset, statementID string) []chainStage {
	shorts := map[string]string{}
	order := map[string]int{}
	for i, p := range ds.Parties {
		shorts[p.ID] = p.Short
		order[p.ID] = i
	}

	stages := make([]chainStage, len(chainStageOrder))
	partiesAt := map[string]map[string]bool{}
	for i, s := range chainStageOrder {
		stages[i] = chainStage{Key: s.Key, Label: s.Label, Meaning: s.Meaning}
		partiesAt[s.Key] = map[string]bool{}
	}

	for _, e := range ds.EvidenceForStatement(statementID) {
		key, ok := chainTypeToStage[e.EvidenceType]
		if !ok {
			continue // context, not a step; placed by chainContextTypes
		}
		for i := range stages {
			if stages[i].Key != key {
				continue
			}
			stages[i].Count++
			if _, ok := shorts[e.PartyID]; ok {
				partiesAt[key][e.PartyID] = true
			}
		}
	}
	// The promise step gains the programme points that name this
	// statement. Those are the party's own words about this exact
	// question, which is what the step is for. A thematic point (same
	// dimension, no link) is deliberately NOT counted: counting it would
	// claim the party promised something about this statement when it
	// only spoke to the theme. Those points live in the theme block
	// instead, labelled as the theme.
	promiseSources := 0
	for _, ref := range ds.ProgramPointsForStatement(statementID) {
		promiseSources++
		if _, ok := shorts[ref.PartyID]; ok {
			partiesAt["lofte"][ref.PartyID] = true
		}
	}
	if promiseSources > 0 {
		for i := range stages {
			if stages[i].Key == "lofte" {
				stages[i].Count += promiseSources
			}
		}
	}
	for i := range stages {
		ids := make([]string, 0, len(partiesAt[stages[i].Key]))
		for id := range partiesAt[stages[i].Key] {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(a, b int) bool { return order[ids[a]] < order[ids[b]] })
		for _, id := range ids {
			stages[i].Parties = append(stages[i].Parties, shorts[id])
		}
	}
	return stages
}

// ChainHasAny reports whether any step in the chain holds a source.
func chainHasAny(stages []chainStage) bool {
	for _, s := range stages {
		if s.Count > 0 {
			return true
		}
	}
	return false
}
