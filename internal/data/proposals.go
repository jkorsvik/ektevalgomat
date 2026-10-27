package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// The proposals index is generated data: every record in a Storting session that
// carries named proposers, from data.stortinget.no, built by
// tools/analytics/build_proposals_index.py.
//
// It answers a question the dataset could not: what did each party actually put on
// the table? A programme is a promise, a vote is a decision, and a representative
// proposal is the initiative in between. Counting them per party is the cheapest
// available measure of where a party spent political capital in a session.
//
// A proposal is attributed to every proposer's party. Most are co-signed, so
// picking the first signer would undercount the others, and the per-party counts
// therefore sum to more than the number of proposals.
//
//go:embed proposals/index.json
var proposalsIndexJSON []byte

// Proposer is one person who signed a proposal.
type Proposer struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PartyName string `json:"party_name"`
	// PartyID is empty for a party the dataset does not score, which is kept
	// rather than dropped: the proposal is real even when its signer is not scored.
	PartyID string `json:"party_id"`
}

// Proposal is one record that carries proposers.
type Proposal struct {
	ID        string `json:"id"`
	Reference string `json:"reference"`
	Title     string `json:"title"`
	// StatusCode is the source's own code, not an interpretation. The values seen
	// in this session are 1, 2 and 5; what each means is the source's business.
	StatusCode  *int       `json:"status_code"`
	CommitteeID string     `json:"committee_id"`
	Committee   string     `json:"committee"`
	Subjects    []string   `json:"subjects"`
	Proposers   []Proposer `json:"proposers"`
}

type proposalsSource struct {
	Name        string `json:"name"`
	API         string `json:"api"`
	HumanURL    string `json:"human_url"`
	Session     string `json:"session"`
	RetrievedAt string `json:"retrieved_at"`
	Note        string `json:"note"`
}

type proposalsIndexFile struct {
	Source                       proposalsSource   `json:"source"`
	Parties                      map[string]string `json:"parties"`
	Proposals                    []Proposal        `json:"proposals"`
	ProposalsPerParty            map[string]int    `json:"proposals_per_party"`
	UnscoredParties              map[string]int    `json:"unscored_parties"`
	ScoredPartiesWithoutProposal []string          `json:"scored_parties_without_proposals"`
	RecordsWithoutProposers      int               `json:"records_without_proposers"`
}

var proposalsIndexDoc = mustLoadProposals()

func mustLoadProposals() proposalsIndexFile {
	var idx proposalsIndexFile
	if err := json.Unmarshal(proposalsIndexJSON, &idx); err != nil {
		panic(fmt.Sprintf("internal/data/proposals/index.json is not valid: %v", err))
	}
	if len(idx.Proposals) == 0 {
		panic("internal/data/proposals/index.json holds no proposals; run tools/analytics/build_proposals_index.py")
	}
	if idx.Source.RetrievedAt == "" || idx.Source.Session == "" {
		panic("internal/data/proposals/index.json has no provenance")
	}
	return idx
}

// Proposals returns every proposal in the session, ordered by reference.
func Proposals() []Proposal { return proposalsIndexDoc.Proposals }

// ProposalSession returns the session the index covers.
func ProposalSession() string { return proposalsIndexDoc.Source.Session }

// ProposalsSourceInfo returns the provenance of the index.
func ProposalsSourceInfo() (name, api, humanURL, retrievedAt, note string) {
	s := proposalsIndexDoc.Source
	return s.Name, s.API, s.HumanURL, s.RetrievedAt, s.Note
}

// UnscoredParties returns the parties that appear as proposers but are not scored
// by the dataset, with how many signatures they contributed.
func UnscoredParties() map[string]int { return proposalsIndexDoc.UnscoredParties }

// RecordsWithoutProposers is how many records in the session were excluded because
// no person was named as a proposer: committee reports, delegation reports and
// other business. Stated so the exclusion is visible rather than implied.
func RecordsWithoutProposers() int { return proposalsIndexDoc.RecordsWithoutProposers }

// ProposalsPerParty derives the per-party count from the proposals themselves.
// Derived, never read from the stored map, so a stale count cannot be shown.
func ProposalsPerParty() map[string]int {
	counts := map[string]int{}
	for _, p := range Proposals() {
		seen := map[string]bool{}
		for _, proposer := range p.Proposers {
			if proposer.PartyID == "" || seen[proposer.PartyID] {
				continue
			}
			seen[proposer.PartyID] = true
			counts[proposer.PartyID]++
		}
	}
	return counts
}

// ProposalsForParty returns the proposals one party signed, in reference order.
func ProposalsForParty(partyID string) []Proposal {
	var out []Proposal
	for _, p := range Proposals() {
		for _, proposer := range p.Proposers {
			if proposer.PartyID == partyID {
				out = append(out, p)
				break
			}
		}
	}
	return out
}
