package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

// The Storting index is generated data: the 169 representatives and the 16
// committees for a session, from the Storting's own open data, built by
// tools/analytics/build_storting_index.py.
//
// It exists because the dataset records what parties did and says nothing about
// who holds the seats those decisions were taken in. A party's strength is a fact
// about power, and it is the denominator for every "N of 9 parties" claim on the
// site.
//
// It also records a refusal. No endpoint returns committee membership, so the
// per-party distribution of committee seats and chair posts is not collected, and
// the attempts are stored beside the data rather than in a session log that the
// next run cannot read.
//
//go:embed storting/index.json
var stortingIndexJSON []byte

// Representative is one person holding a seat for a session.
type Representative struct {
	ID         string `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	PartyName  string `json:"party_name"`
	PartyCode  string `json:"party_code"`
	PartyID    string `json:"party_id"`
	County     string `json:"county"`
	CountyCode string `json:"county_code"`
	Deputy     bool   `json:"deputy"`
}

// Committee is one standing committee, by id and name. Membership is not included;
// see RefusedCommitteeMembership.
type Committee struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Refusal records something the source could not answer, with the evidence that it
// could not, so the next attempt starts from what is known.
type Refusal struct {
	Reason   string   `json:"reason"`
	Attempts []string `json:"attempts"`
}

type stortingCrossCheck struct {
	VoteringID         int            `json:"votering_id"`
	Method             string         `json:"method"`
	PerPartyVoteTotals map[string]int `json:"per_party_vote_totals"`
}

type stortingSource struct {
	Name        string `json:"name"`
	API         string `json:"api"`
	HumanURL    string `json:"human_url"`
	Session     string `json:"session"`
	RetrievedAt string `json:"retrieved_at"`
	Note        string `json:"note"`
}

type stortingIndexFile struct {
	Source          stortingSource     `json:"source"`
	Parties         map[string]string  `json:"parties"`
	Representatives []Representative   `json:"representatives"`
	Committees      []Committee        `json:"committees"`
	CrossCheck      stortingCrossCheck `json:"cross_check"`
	Refused         map[string]Refusal `json:"refused"`
}

var stortingIndexDoc = mustLoadStorting()

func mustLoadStorting() stortingIndexFile {
	var idx stortingIndexFile
	if err := json.Unmarshal(stortingIndexJSON, &idx); err != nil {
		panic(fmt.Sprintf("internal/data/storting/index.json is not valid: %v", err))
	}
	if len(idx.Representatives) == 0 {
		panic("internal/data/storting/index.json holds no representatives; run tools/analytics/build_storting_index.py")
	}
	if idx.Source.RetrievedAt == "" || idx.Source.Session == "" {
		panic("internal/data/storting/index.json has no provenance")
	}
	return idx
}

// Representatives returns every representative, ordered by party then name.
func Representatives() []Representative { return stortingIndexDoc.Representatives }

// Committees returns the session's committees.
func Committees() []Committee { return stortingIndexDoc.Committees }

// StortingSession returns the session the index covers.
func StortingSession() string { return stortingIndexDoc.Source.Session }

// StortingSourceInfo returns the provenance of the index.
func StortingSourceInfo() (name, api, humanURL, retrievedAt string) {
	s := stortingIndexDoc.Source
	return s.Name, s.API, s.HumanURL, s.RetrievedAt
}

// CommitteeMembershipRefusal returns the recorded refusal, and whether one exists.
// A caller that wants committee membership must consult this instead of assuming
// the data is simply missing.
func CommitteeMembershipRefusal() (Refusal, bool) {
	r, ok := stortingIndexDoc.Refused["committee_membership"]
	return r, ok
}

// SeatCounts derives the number of seats per party from the representative list.
// Derived, never stored: a stored seat count could drift from the people it counts.
func SeatCounts() map[string]int {
	counts := map[string]int{}
	for _, r := range stortingIndexDoc.Representatives {
		counts[r.PartyID]++
	}
	return counts
}

// SeatTotal is how many seats the index accounts for.
func SeatTotal() int { return len(stortingIndexDoc.Representatives) }

// SeatRanking returns parties ordered by seats, most first, with ties broken by
// party id so the order is deterministic.
func SeatRanking() []string {
	counts := SeatCounts()
	ids := make([]string, 0, len(counts))
	for id := range counts {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		if counts[ids[i]] != counts[ids[j]] {
			return counts[ids[i]] > counts[ids[j]]
		}
		return ids[i] < ids[j]
	})
	return ids
}

// VoteCrossCheck returns the per-party totals from the cross-check votering. The
// test compares these with SeatCounts: the two come from separately produced
// artefacts of the same authority, so agreement is evidence and disagreement is a
// defect in one of them.
func VoteCrossCheck() (voteringID int, perParty map[string]int, method string) {
	c := stortingIndexDoc.CrossCheck
	return c.VoteringID, c.PerPartyVoteTotals, c.Method
}

// PartyCode returns the source's own code for a dataset party id, e.g. "ap" -> "A".
func PartyCode(partyID string) (string, bool) {
	for _, r := range stortingIndexDoc.Representatives {
		if r.PartyID == partyID && r.PartyCode != "" {
			return r.PartyCode, true
		}
	}
	return "", false
}
