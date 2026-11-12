package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

// Committee membership, built from Stortinget's own committee pages.
//
// The export API cannot answer this: `komite?id=` fails server-side naming
// HentKomiteMedlemmer, and there is no komitemedlemmer endpoint. The committee
// pages embed the membership as JSON keyed by the same `perid` the person
// records use, so membership joins without name matching.
//
// What this cannot say, and does not pretend to: when a seat began or ended. The
// pages state who sits there now. This is a snapshot, like the verv register, and
// the artefact carries that word so no caller can render it as a dated role.
//
//go:embed komiteer/index.json
var committeeJSON []byte

// CommitteeRosterMember is one person on one committee, with the role the page states.
type CommitteeRosterMember struct {
	PersonID string `json:"person_id"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

// CommitteeRoster is one body and who sits on it.
type CommitteeRoster struct {
	ID      string                  `json:"id"`
	Name    string                  `json:"name"`
	URL     string                  `json:"url"`
	Members []CommitteeRosterMember `json:"members"`
}

// CommitteeSeat is one person's membership seen from the person's side.
type CommitteeSeat struct {
	CommitteeID   string `json:"committee_id"`
	CommitteeName string `json:"committee_name"`
	Role          string `json:"role"`
}

// CommitteeIndex is the whole artefact plus the bodies the source does not
// publish, so a page can say which ones it cannot describe instead of leaving a
// reader to conclude that nobody sits on them.
type CommitteeIndex struct {
	GeneratedAt  string                     `json:"generated_at"`
	Source       string                     `json:"source"`
	Session      string                     `json:"session"`
	Snapshot     bool                       `json:"snapshot"`
	SnapshotNote string                     `json:"snapshot_note"`
	Committees   []CommitteeRoster          `json:"committees"`
	People       map[string][]CommitteeSeat `json:"people"`
	Unpublished  []map[string]string        `json:"unpublished"`
	Counts       map[string]int             `json:"counts"`
}

var committeeIndex = mustLoadCommittees()

func mustLoadCommittees() CommitteeIndex {
	var ci CommitteeIndex
	if err := json.Unmarshal(committeeJSON, &ci); err != nil {
		panic(fmt.Sprintf("internal/data/komiteer/index.json is not valid: %v", err))
	}
	return ci
}

// CommitteeRosters returns every body with a published member list.
func CommitteeRosters() []CommitteeRoster { return committeeIndex.Committees }

// CommitteesForPerson returns the seats one person holds, leader first.
func CommitteesForPerson(id string) []CommitteeSeat {
	seats := committeeIndex.People[id]
	out := append([]CommitteeSeat(nil), seats...)
	sort.SliceStable(out, func(i, j int) bool {
		if (out[i].Role == "Leder") != (out[j].Role == "Leder") {
			return out[i].Role == "Leder"
		}
		return out[i].CommitteeName < out[j].CommitteeName
	})
	return out
}

// CommitteeSummary exposes the artefact's coverage for a page that must state it.
func CommitteeSummary() CommitteeIndex { return committeeIndex }
