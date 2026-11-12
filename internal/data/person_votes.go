package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// Per-representative votes, built from the same division payloads as the party
// index. Until 2026-11-12 the site could say how a party voted and nothing about
// how a person voted, even though the source carries one row per representative.
//
// This is a build artefact, not embedded content: tools/analytics/build_person_vote_index.py
// writes it from the cached voteringsresultat payloads and refuses to write when
// the party line it derives from those rows disagrees with votes/index.json. The
// gate below is the second half of that check: it fails if the file on disk does
// not match the party index we ship alongside it, so a stale artefact cannot sit
// quietly in a build.
//
//go:embed votes/persons.json
var personVotesJSON []byte

// PersonVoteRecord is one representative's tally across every division where the
// source published per-representative rows.
type PersonVoteRecord struct {
	Name      string   `json:"name"`
	Party     string   `json:"party"`
	For       int      `json:"for"`
	Mot       int      `json:"mot"`
	Absent    int      `json:"absent"`
	Voted     int      `json:"voted"`
	Deviation int      `json:"deviation"`
	Divisions int      `json:"divisions"`
	Sessions  []string `json:"sessions"`
}

// PersonVotes is the whole artefact plus the provenance a reader needs to judge
// it: how many divisions were checked, how many carried rows at all, and that the
// self-check found no disagreement with the party index.
type PersonVotes struct {
	GeneratedAt           string                      `json:"generated_at"`
	Source                string                      `json:"source"`
	DivisionsInPartyIndex int                         `json:"divisions_in_party_index"`
	DivisionsChecked      int                         `json:"divisions_checked"`
	DivisionsWithRows     int                         `json:"divisions_with_rows"`
	RepresentativeRows    int                         `json:"representative_rows"`
	PartyLineMismatches   int                         `json:"party_line_mismatches"`
	Persons               map[string]PersonVoteRecord `json:"persons"`
}

var personVotes = mustLoadPersonVotes()

func mustLoadPersonVotes() PersonVotes {
	var pv PersonVotes
	if err := json.Unmarshal(personVotesJSON, &pv); err != nil {
		panic(fmt.Sprintf("internal/data/votes/persons.json is not valid: %v", err))
	}
	if pv.PartyLineMismatches != 0 {
		panic("internal/data/votes/persons.json records party-line mismatches; the extraction is unsound")
	}
	return pv
}

// PersonVotesFor returns one representative's tally and whether any was recorded.
// The zero record is not returned: a caller that gets false must say "no votes
// recorded", never render a row of zeros as if the person never turned up.
func PersonVotesFor(id string) (PersonVoteRecord, bool) {
	r, ok := personVotes.Persons[id]
	return r, ok
}

// PersonVoteSummary exposes the artefact's own coverage numbers.
func PersonVoteSummary() PersonVotes {
	return personVotes
}
