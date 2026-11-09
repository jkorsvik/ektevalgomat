package data

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Public-office records: who holds which office, when, and on what source.
//
// This is not evidence and carries no position. It exists so a reader can see
// that a politician's declared interests and posts are a matter of record, and
// so a claim about "who was where" has something to point at. Every role names
// its source; a role without one is a defect.
//
// Boundary: only public-office facts. Birth dates, addresses, phone numbers,
// family, health, gifts, travel, property and private money are never stored —
// tools/autopilot/lobby/gate_people.py rejects them before a record is promoted.
// The register's own §5–§11 (former employer, future employer, support, property,
// company holdings, travel, gifts, other) are dropped at index time.
//
// A `verv` role's `from` is either a date the register states ("Fra 29. oktober
// 2026", "2022 - 2025") or the register's snapshot date when it states none. A
// snapshot date is not a role start; SourceNote says so, and the page must not
// present it as one.
type PersonRole struct {
	Role         string `yaml:"role"`
	Organisation string `yaml:"organisation"`
	Orgnr        string `yaml:"orgnr,omitempty"`
	Post         string `yaml:"post,omitempty"`
	From         string `yaml:"from"`
	To           string `yaml:"to"`
	SourceURL    string `yaml:"source_url"`
	SourceTier   string `yaml:"source_tier"`
	RetrievedAt  string `yaml:"retrieved_at"`
	SourceNote   string `yaml:"source_note,omitempty"`
}

// Person is one public official's dated record.
type Person struct {
	ID           string       `yaml:"id"`
	Name         string       `yaml:"name"`
	Scope        string       `yaml:"scope"`
	ConsentBasis string       `yaml:"consent_basis"`
	PartyID      string       `yaml:"party_id,omitempty"`
	Roles        []PersonRole `yaml:"roles"`
}

// Holding reports whether the role is still held at the snapshot. A role with
// no end date is open, which is not the same as "held": the register may have
// dropped it.
func (r PersonRole) Holding() bool { return r.To == "" }

// SnapshotDated reports whether From is the register's snapshot rather than a
// date the register stated.
func (r PersonRole) SnapshotDated() bool {
	return strings.Contains(r.SourceNote, "snapshot")
}

var people = mustLoadPeople()

func mustLoadPeople() []Person {
	entries, err := contentFS.ReadDir(path.Join(contentRoot, "people"))
	if err != nil {
		panic("data: reading content/people: " + err.Error())
	}
	out := make([]Person, 0, len(entries))
	seen := make(map[string]string, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		b, err := contentFS.ReadFile(path.Join(contentRoot, "people", e.Name()))
		if err != nil {
			panic("data: reading person " + e.Name() + ": " + err.Error())
		}
		var p Person
		if err := yaml.Unmarshal(b, &p); err != nil {
			panic("data: parsing person " + e.Name() + ": " + err.Error())
		}
		if p.ID == "" {
			panic("data: person " + e.Name() + " has no id")
		}
		if file, dup := seen[p.ID]; dup {
			panic(fmt.Sprintf("data: person id %s appears in both %s and %s", p.ID, file, e.Name()))
		}
		seen[p.ID] = e.Name()
		if len(p.Roles) == 0 {
			panic("data: person " + p.ID + " has no roles")
		}
		for i := range p.Roles {
			r := &p.Roles[i]
			if r.Role == "" || r.Organisation == "" || r.SourceURL == "" || r.From == "" {
				panic(fmt.Sprintf("data: person %s role %d is incomplete", p.ID, i))
			}
		}
		sort.SliceStable(p.Roles, func(i, j int) bool {
			if p.Roles[i].From != p.Roles[j].From {
				return p.Roles[i].From > p.Roles[j].From
			}
			return p.Roles[i].Role < p.Roles[j].Role
		})
		out = append(out, p)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// RegisterSnapshotDate is the date the verv register we indexed was refreshed.
// It is pinned here because the index is a build artefact, not embedded content;
// TestRegisterSnapshotMatchesIndex fails if the pinned value drifts from
// internal/data/verv/index.json.
const RegisterSnapshotDate = "2026-10-30"

// VervSnapshotDate returns the register snapshot date.
func VervSnapshotDate() string { return RegisterSnapshotDate }

// People returns every public-office record, sorted by name.
func People() []Person { return people }

// PersonByID returns one record, and whether it exists.
func PersonByID(id string) (Person, bool) {
	for _, p := range people {
		if p.ID == id {
			return p, true
		}
	}
	return Person{}, false
}

// RegisterHistory is the diff between two editions of the verv register. A role
// present in an earlier edition and absent from the newest was held in between;
// the register states neither its start nor its end, so none is written here.
type RegisterHistory struct {
	GeneratedAt     string   `json:"generated_at"`
	ComparedBetween []string `json:"compared_between"`
	Counts          struct {
		PeopleInBoth int `json:"people_in_both"`
		RolesCarried int `json:"roles_carried"`
		RolesEnded   int `json:"roles_ended"`
		RolesNew     int `json:"roles_new"`
	} `json:"counts"`
	Ended []RegisterChange `json:"ended"`
	New   []RegisterChange `json:"new"`
}

// RegisterChange is one post that appeared or disappeared between two editions.
type RegisterChange struct {
	Name      string `json:"name"`
	Party     string `json:"party"`
	Section   int    `json:"section"`
	Kind      string `json:"kind"`
	Text      string `json:"text"`
	Seen      string `json:"seen,omitempty"`
	GoneBy    string `json:"gone_by,omitempty"`
	FirstSeen string `json:"first_seen,omitempty"`
}

var registerHistory = mustLoadHistory()

func mustLoadHistory() RegisterHistory {
	b, err := contentFS.ReadFile(path.Join(contentRoot, "verv", "history.json"))
	if err != nil {
		// The history is optional: a build without two register editions still
		// serves the person page, it just cannot say what changed.
		return RegisterHistory{}
	}
	var h RegisterHistory
	if err := json.Unmarshal(b, &h); err != nil {
		panic("data: parsing verv/history.json: " + err.Error())
	}
	return h
}

// RegisterHistoryDiff returns the register comparison, or a zero value when no
// two editions were compared.
func RegisterHistoryDiff() RegisterHistory { return registerHistory }
