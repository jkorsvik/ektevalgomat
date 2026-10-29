package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

// The vote index is generated data, not source: tools/analytics/build_vote_index.py
// fetches the official tally for every votering the content cites and writes one
// record per vote with each party's side on it. It exists because the site could
// say what one party did on one statement but could not answer "who votes with
// whom": the evidence rows cite 400+ voterings, and the tallies were never
// stored as data.
//
// It is embedded for the same reason content/ and archive/ are: the runtime
// container is FROM scratch with no writable paths, so there is no filesystem to
// read it from at request time.
//
//go:embed votes/index.json
var voteIndexJSON []byte

// Vote is one votering as the official record states it, plus the evidence rows
// that cite it. It carries no date on purpose: the export has no
// votering-to-sak path, so a date here would be missing for most votes.
//
// Parties holds for | mot | split | absent per party id. A split is kept as its
// own value rather than collapsed into a majority: 3 for and 3 against is a
// different fact from either side, and calling it "for" would invent a position
// the record does not support.
type Vote struct {
	VoteringID int `json:"votering_id"`
	// Unanimous marks a vote taken without a division: the API returns no party
	// tally, so Parties is empty and the vote contributes to no comparison.
	Unanimous bool              `json:"unanimous"`
	Forslag   string            `json:"forslag"`
	Parties   map[string]string `json:"parties"`
	CitedBy   []string          `json:"cited_by"`
	// CodedParties are the parties whose coded position in the content rests on
	// this vote. Empty means the vote is cited as context, or by an institutional
	// row that codes nobody.
	CodedParties []string `json:"coded_parties"`
	// Session is the storting session the vote was taken in. The generator takes it
	// from the session walk when the walk named the vote, and otherwise from the
	// vote's own date, since a session runs October to September. Every vote carries
	// one (measured 2026-10-29: 25773 of 25773); a vote with neither a walk claim nor
	// a date is the only case that would leave it empty.
	Session string `json:"session"`
	// Date is the day the vote was taken, YYYY-MM-DD, from the vote's own
	// votering_tid. Every vote in the index carries it (25773 of 25773): a vote met
	// through a citation before any walk ran still had its date once the session was
	// walked. It is a measured timestamp, not the
	// siste_oppdatert_dato maintenance stamp on the case, which moves whenever a
	// case is edited and would date every vote to the last bulk update.
	Date string `json:"date"`
	// UnmappedParties holds the members of a division who belong to a party
	// outside the nine this dataset compares, keyed by their API id, with the
	// same for/mot/split/absent vocabulary as Parties. Empty means the nine were
	// the whole division.
	//
	// It exists because dropping them made the artefact state a tally the
	// division did not have: votering 18486 fell 21 for, the nine account for 20,
	// and the 21st is Pasientfokus. An unknown party id and a party that did not
	// vote both produced no entry, so nothing could tell the two apart. 10 567 of
	// the 14 946 votes carry at least one, almost all of them Pasientfokus or an
	// independent, which is why this is not a rounding detail.
	UnmappedParties map[string]string `json:"unmapped_parties,omitempty"`
}

type voteIndexFile struct {
	GeneratedAt string `json:"generated_at"`
	Votes       []Vote `json:"votes"`
}

var voteIndexDoc = mustLoadVotes()

func mustLoadVotes() voteIndexFile {
	var idx voteIndexFile
	if err := json.Unmarshal(voteIndexJSON, &idx); err != nil {
		// A malformed generated file is a build defect, not a runtime condition:
		// the generator is in the repo and its output is committed, so this can
		// only happen if someone edited the artefact by hand.
		panic(fmt.Sprintf("internal/data/votes/index.json is not valid: %v", err))
	}
	if len(idx.Votes) == 0 {
		panic("internal/data/votes/index.json holds no votes; run tools/analytics/build_vote_index.py")
	}
	return idx
}

// Votes returns the indexed votes, newest first. The order is part of the
// contract so a page renders deterministically.
func Votes() []Vote {
	return voteIndexDoc.Votes
}

// VoteCount is how many votes the index holds.
func VoteCount() int { return len(voteIndexDoc.Votes) }

// VotesGeneratedAt is the date the index was built, so a page can say how fresh
// it is instead of implying it is live.
func VotesGeneratedAt() string { return voteIndexDoc.GeneratedAt }

// VoteSession is one storting session the index covers, with how many votes came
// from walking it.
type VoteSession struct {
	Session string
	Votes   int
}

// VoteSessions reports which sessions the index was built from, oldest first.
//
// It exists so a page can say what it covers instead of implying it covers
// everything: the index holds three sessions out of the eighteen the API serves, and
// a reader looking at an alignment matrix has no other way to know that.
func VoteSessions() []VoteSession {
	counts := map[string]int{}
	for _, v := range voteIndexDoc.Votes {
		if v.Session == "" {
			continue // cited-only: the session is not recorded, so it is not claimed
		}
		counts[v.Session]++
	}
	out := make([]VoteSession, 0, len(counts))
	for s, n := range counts {
		out = append(out, VoteSession{Session: s, Votes: n})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Session < out[j].Session })
	return out
}

// VotesWithSession counts the votes whose session is recorded.
func VotesWithSession() int {
	n := 0
	for _, v := range voteIndexDoc.Votes {
		if v.Session != "" {
			n++
		}
	}
	return n
}
