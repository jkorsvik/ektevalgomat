package data

import (
	"testing"
)

// The proposals index is a copy of the Storting's own session list, so the tests are
// about the copy being faithful. The per-party counts are a political claim, which
// makes them the part most worth checking twice.
func TestProposalsIndexIsComplete(t *testing.T) {
	proposals := Proposals()
	// Pinned to what the session returned on 2026-10-27. A floor would pass while
	// proposals were quietly dropped by a filter change.
	const sourceCount = 362
	if len(proposals) != sourceCount {
		t.Fatalf("index holds %d proposals, the session returned %d on 2026-10-27", len(proposals), sourceCount)
	}
	for _, p := range proposals {
		if p.ID == "" || p.Reference == "" || p.Title == "" {
			t.Errorf("incomplete proposal: %+v", p)
		}
		if len(p.Proposers) == 0 {
			t.Errorf("%s has no proposer but was stored", p.Reference)
		}
	}
	if RecordsWithoutProposers() == 0 {
		t.Error("no excluded records reported: the exclusion would be invisible")
	}
}

// Every proposer must be attributable, and the only allowed empty party id is a
// party the dataset does not score, which must be named in the report.
func TestProposersAreAttributable(t *testing.T) {
	known := map[string]bool{}
	for _, p := range Store.Parties {
		known[p.ID] = true
	}
	unscored := UnscoredParties()
	empty := map[string]int{}
	for _, p := range Proposals() {
		for _, proposer := range p.Proposers {
			if proposer.PartyID == "" {
				empty[proposer.PartyName]++
				continue
			}
			if !known[proposer.PartyID] {
				t.Errorf("%s: unknown party id %q", p.Reference, proposer.PartyID)
			}
		}
	}
	for name, n := range empty {
		if _, reported := unscored[name]; !reported {
			t.Errorf("%d signatures from unscored party %q are not in the report", n, name)
		}
	}
	for name, n := range unscored {
		if n <= 0 {
			t.Errorf("unscored party %q reported with %d signatures", name, n)
		}
	}
}

// The per-party count is derived. It must equal the stored map, which is only a
// receipt from the builder: if the two ever disagree, the stored one is wrong.
func TestProposalsPerPartyIsDerivedAndMatchesTheReceipt(t *testing.T) {
	derived := ProposalsPerParty()
	stored := proposalsIndexDoc.ProposalsPerParty
	if len(derived) == 0 {
		t.Fatal("no per-party count derived")
	}
	for id, n := range stored {
		if derived[id] != n {
			t.Errorf("party %s: derived %d, stored %d", id, derived[id], n)
		}
	}
	for id, n := range derived {
		if _, ok := stored[id]; !ok {
			t.Errorf("party %s has %d proposals but is absent from the stored receipt", id, n)
		}
	}
	// Co-signing means the sum exceeds the number of proposals. If it ever equals
	// it, attribution has collapsed to one party per proposal.
	sum := 0
	for _, n := range derived {
		sum += n
	}
	if sum <= len(Proposals()) {
		t.Errorf("per-party sum %d is not greater than %d proposals: co-signing was lost", sum, len(Proposals()))
	}
	t.Logf("%d proposals, %d party attributions across %d parties", len(Proposals()), sum, len(derived))
}

// A party with zero proposals would be a real finding, but it must not be confused
// with a party whose proposals failed to attribute. Every scored party appears here.
func TestEveryScoredPartyHasProposals(t *testing.T) {
	counts := ProposalsPerParty()
	for _, p := range Store.Parties {
		if counts[p.ID] == 0 {
			t.Errorf("party %s has no proposal in %s", p.ID, ProposalSession())
		}
	}
}

// The status code is the source's own, and the observed set is small. Asserting it
// catches a schema change that would otherwise render as an unknown code.
func TestProposalStatusCodesAreTheObservedSet(t *testing.T) {
	allowed := map[int]bool{1: true, 2: true, 5: true}
	for _, p := range Proposals() {
		if p.StatusCode == nil {
			t.Errorf("%s has no status code", p.Reference)
			continue
		}
		if !allowed[*p.StatusCode] {
			t.Errorf("%s: unexpected status code %d (observed set is 1, 2, 5)", p.Reference, *p.StatusCode)
		}
	}
}

func TestProposalsForPartyFilters(t *testing.T) {
	all := Proposals()
	got := ProposalsForParty("frp")
	if len(got) == 0 {
		t.Fatal("no proposals for frp")
	}
	if len(got) >= len(all) {
		t.Errorf("frp has %d of %d proposals: the filter did not filter", len(got), len(all))
	}
	for _, p := range got {
		found := false
		for _, proposer := range p.Proposers {
			if proposer.PartyID == "frp" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s is attributed to frp but no proposer is frp", p.Reference)
		}
	}
	if len(ProposalsForParty("finnes-ikke")) != 0 {
		t.Error("a party that does not exist returned proposals")
	}
}

func TestProposalsCarryProvenance(t *testing.T) {
	name, api, human, retrieved, note := ProposalsSourceInfo()
	if name == "" || api == "" || human == "" || retrieved == "" {
		t.Errorf("incomplete provenance: %q %q %q %q", name, api, human, retrieved)
	}
	if note == "" {
		t.Error("no note: the attribution rule would be undocumented")
	}
	if ProposalSession() != "2025-2026" {
		t.Errorf("session = %q", ProposalSession())
	}
}
