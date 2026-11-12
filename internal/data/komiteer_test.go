package data

import "testing"

// Committee membership is scraped from a page tree the source may reorganise, so
// these tests check shape and coverage rather than trusting the last run. A page
// that stops parsing yields zero members, which is exactly what must not ship.

func TestCommitteesHaveMembersAndKnownRoles(t *testing.T) {
	committees := CommitteeRosters()
	if len(committees) < 10 {
		t.Fatalf("only %d committees carry a member list; the scrape has probably broken", len(committees))
	}
	known := map[string]bool{"Leder": true, "Første nestleder": true, "Nestleder": true, "Medlem": true}
	leaders := 0
	seen := map[string]string{}
	for _, c := range committees {
		if c.ID == "" || c.Name == "" || c.URL == "" {
			t.Errorf("committee %q is missing id, name or url", c.Name)
		}
		if len(c.Members) < 5 {
			t.Errorf("%s has %d members; a committee of that size is not a committee", c.ID, len(c.Members))
		}
		for _, m := range c.Members {
			if !known[m.Role] {
				t.Errorf("%s: %s has role %q, which is not one the source uses", c.ID, m.PersonID, m.Role)
			}
			if m.PersonID == "" || m.Name == "" {
				t.Errorf("%s: a member row has no id or no name", c.ID)
			}
			if m.Role == "Leder" {
				leaders++
			}
			if prev, dup := seen[m.PersonID+"@"+c.ID]; dup {
				t.Errorf("%s sits on %s twice", prev, c.ID)
			}
			seen[m.PersonID+"@"+c.ID] = m.PersonID
		}
	}
	// Every committee has exactly one leader. Zero in total means the role column
	// stopped parsing while the membership rows still came through.
	if leaders < len(committees)/2 {
		t.Errorf("only %d leaders across %d committees; the role column is probably not parsing",
			leaders, len(committees))
	}
}

// Sitting representatives should be findable on a committee. Ten were not when
// this was built, and the bodies without a published page explain some of them.
func TestMostPublishedRepresentativesHaveASeat(t *testing.T) {
	withSeat := 0
	for _, p := range People() {
		if len(CommitteesForPerson(p.ID)) > 0 {
			withSeat++
		}
	}
	// 159 of 169 at build time. A large drop means the scrape lost people.
	if withSeat < 140 {
		t.Errorf("only %d of %d published representatives are on any committee", withSeat, len(People()))
	}
}

func TestCommitteeArtefactStatesWhatItCannotCover(t *testing.T) {
	ci := CommitteeSummary()
	if !ci.Snapshot {
		t.Error("the artefact must declare itself a snapshot: the pages state no start or end date")
	}
	if ci.SnapshotNote == "" {
		t.Error("no snapshot note, so a caller could render membership as a dated role")
	}
	if len(ci.Unpublished) == 0 {
		t.Error("no unpublished bodies recorded; the source omits six and that must be visible")
	}
	for _, u := range ci.Unpublished {
		if u["id"] == "" || u["name"] == "" {
			t.Errorf("unpublished entry without id or name: %v", u)
		}
	}
}

func TestCommitteesForUnknownPersonIsEmpty(t *testing.T) {
	if got := CommitteesForPerson("NOBODY"); len(got) != 0 {
		t.Errorf("unknown person returned %d seats", len(got))
	}
}

// A leader must come first, and the sort belongs here rather than in the sentence
// builder, so that every caller gets the same order.
func TestCommitteeSeatsPutTheLeaderFirst(t *testing.T) {
	for _, p := range People() {
		seats := CommitteesForPerson(p.ID)
		if len(seats) < 2 {
			continue
		}
		seenMember := false
		for _, s := range seats {
			if s.Role != "Leder" {
				seenMember = true
				continue
			}
			if seenMember {
				t.Errorf("%s: a leader is listed after an ordinary seat", p.Name)
			}
		}
	}
}

// A committee named after one of its own members was a real bug: the page-title
// fallback shared a variable with the member loop, so every committee took the
// name of its last member. This catches that class, and a name that is just the
// bare id (the fallback failing silently).
func TestCommitteeNamesAreNames(t *testing.T) {
	for _, c := range CommitteeRosters() {
		if c.Name == c.ID {
			t.Errorf("%s has no name beyond its id", c.ID)
		}
		for _, m := range c.Members {
			if c.Name == m.Name {
				t.Errorf("%s is named after its own member %q", c.ID, m.Name)
			}
		}
	}
}
