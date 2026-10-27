package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
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
