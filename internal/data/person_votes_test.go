package data

import "testing"

// The per-representative vote artefact is only worth shipping if it can be shown
// to agree with the party index it was derived from and to be internally
// consistent. tools/analytics/build_person_vote_index.py already refuses to write
// a file when the party line it derives disagrees with the stored one; these tests
// guard the artefact that actually got embedded, so a stale or hand-edited file
// cannot pass.

func TestPersonVotesAgreeWithPartyIndex(t *testing.T) {
	pv := PersonVoteSummary()
	if pv.PartyLineMismatches != 0 {
		t.Fatalf("artefact records %d party-line mismatches; it must not ship", pv.PartyLineMismatches)
	}
	if pv.DivisionsChecked != len(Votes()) {
		t.Errorf("checked %d divisions but the party index holds %d: the artefact is stale",
			pv.DivisionsChecked, len(Votes()))
	}
	// A file that measured nothing would otherwise pass every check below.
	if pv.DivisionsWithRows == 0 || pv.RepresentativeRows == 0 {
		t.Fatalf("artefact covers %d divisions and %d rows: nothing was measured",
			pv.DivisionsWithRows, pv.RepresentativeRows)
	}
	// The source publishes an empty list for a share of divisions. That is a fact
	// about the source, not a defect here, but if it ever became most of them the
	// timeline would be worthless and someone must notice.
	if pv.DivisionsWithRows*2 < pv.DivisionsChecked {
		t.Errorf("only %d of %d divisions carry representative rows; the coverage has collapsed",
			pv.DivisionsWithRows, pv.DivisionsChecked)
	}
}

func TestPersonVotesAreInternallyConsistent(t *testing.T) {
	pv := PersonVoteSummary()
	if len(pv.Persons) == 0 {
		t.Fatal("no persons in the artefact")
	}
	sum, deviants := 0, 0
	for id, r := range pv.Persons {
		if r.Name == "" {
			t.Errorf("%s has no name; a vote record a reader cannot attribute is noise", id)
		}
		if r.For+r.Mot+r.Absent != r.Divisions {
			t.Errorf("%s: for(%d)+mot(%d)+absent(%d) != divisions(%d)",
				id, r.For, r.Mot, r.Absent, r.Divisions)
		}
		if r.Voted != r.For+r.Mot {
			t.Errorf("%s: voted(%d) != for+mot(%d)", id, r.Voted, r.For+r.Mot)
		}
		if r.Deviation > r.Voted {
			t.Errorf("%s: %d deviations out of %d votes cast is impossible", id, r.Deviation, r.Voted)
		}
		if r.Deviation > 0 {
			deviants++
		}
		sum += r.Divisions
	}
	if sum != pv.RepresentativeRows {
		t.Errorf("per-person divisions sum to %d but the artefact counted %d rows; rows were lost or double counted",
			sum, pv.RepresentativeRows)
	}
	// If nobody ever deviates, either the parties are perfect or the deviation
	// calculation is broken. Measured coverage says the second is likelier.
	if deviants == 0 {
		t.Error("no representative ever deviates from their party line; the calculation is probably blind")
	}
}

// Every sitting representative the site publishes should be reachable in the vote
// artefact. A person page that says "no votes recorded" about someone who has been
// in the chamber for years is worse than no section at all.
func TestEverySittingRepresentativeHasVotes(t *testing.T) {
	missing := 0
	for _, p := range People() {
		if _, ok := PersonVotesFor(p.ID); !ok {
			missing++
		}
	}
	// Historic seat-holders and representatives who left before the earliest
	// division in the index legitimately have none, so this is a floor, not zero.
	if missing > 5 {
		t.Errorf("%d of %d published representatives have no vote record at all", missing, len(People()))
	}
}

func TestPersonVotesForUnknownIdIsAbsent(t *testing.T) {
	if _, ok := PersonVotesFor("NOBODY"); ok {
		t.Error("an unknown id must not return a zero record")
	}
}
