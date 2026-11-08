package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// The coding review is generated from two blind agent reviews and a separate
// adjudication. It measures the editorial step that quote and vote verification
// cannot measure. Agent review is not independent human review; the artefact
// carries that limit next to its counts.
//
//go:embed coding_review/index.json
var codingReviewJSON []byte

type CodingReviewMeta struct {
	ReviewedAt              string            `json:"reviewed_at"`
	SourceRevision          string            `json:"source_revision"`
	InputSHA256             map[string]string `json:"input_sha256"`
	Protocol                string            `json:"protocol"`
	Scale                   string            `json:"scale"`
	Seed                    string            `json:"seed"`
	PopulationCells         int               `json:"population_cells"`
	SampleCells             int               `json:"sample_cells"`
	ReviewerA               string            `json:"reviewer_a"`
	ReviewerB               string            `json:"reviewer_b"`
	Adjudicator             string            `json:"adjudicator"`
	ExactAgreement          int               `json:"exact_agreement"`
	SameDirection           int               `json:"same_direction"`
	SubstantiveDisagreement int               `json:"substantive_disagreement"`
	ChangedPositions        int               `json:"changed_positions"`
	KeptPositions           int               `json:"kept_positions"`
	UnresolvedPositions     int               `json:"unresolved_positions"`
	Limitations             string            `json:"limitations"`
}

type CodingReviewAmbiguity struct {
	ID      string   `json:"id"`
	Text    string   `json:"text"`
	Affects []string `json:"affects"`
}

type ReviewedEvidence struct {
	ID     string `json:"id"`
	Date   string `json:"date"`
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
}

type CodingReviewPacket struct {
	PacketID              string             `json:"packet_id"`
	StatementID           string             `json:"statement_id"`
	PartyID               string             `json:"party_id"`
	ReviewerA             *int               `json:"reviewer_a"`
	ReviewerB             *int               `json:"reviewer_b"`
	PublishedPosition     *int               `json:"published_position"`
	FinalPosition         *int               `json:"final_position"`
	Agreement             string             `json:"agreement"`
	Adjudicated           bool               `json:"adjudicated"`
	Resolution            string             `json:"resolution"`
	DisputeType           string             `json:"dispute_type"`
	ReviewerAReason       string             `json:"reviewer_a_reason"`
	ReviewerBReason       string             `json:"reviewer_b_reason"`
	AdjudicationRationale string             `json:"adjudication_rationale"`
	SourceURLs            []string           `json:"source_urls"`
	ReviewedEvidence      []ReviewedEvidence `json:"reviewed_evidence"`
}

type CodingReview struct {
	Meta        CodingReviewMeta        `json:"meta"`
	Ambiguities []CodingReviewAmbiguity `json:"ambiguities"`
	Packets     []CodingReviewPacket    `json:"packets"`
}

var codingReviewDoc = mustLoadCodingReview()

func mustLoadCodingReview() CodingReview {
	var review CodingReview
	if err := json.Unmarshal(codingReviewJSON, &review); err != nil {
		panic(fmt.Sprintf("internal/data/coding_review/index.json is invalid: %v", err))
	}
	m := review.Meta
	if m.PopulationCells <= 0 || m.SampleCells != 18 || len(review.Packets) != m.SampleCells {
		panic("coding review must contain 18 packets from a non-empty population")
	}
	parties := map[string]int{}
	agreement := map[string]int{}
	changed, kept, unresolved := 0, 0, 0
	seen := map[string]bool{}
	for _, p := range review.Packets {
		if seen[p.PacketID] || p.StatementID == "" || p.PartyID == "" {
			panic("coding review has a duplicate packet or missing cell identity")
		}
		if p.Resolution != "unresolved" && p.FinalPosition == nil {
			panic("coding review has no final position for a resolved packet")
		}
		if p.ReviewerAReason == "" || p.ReviewerBReason == "" || len(p.ReviewedEvidence) == 0 {
			panic("coding review is missing reviewer reasons or reviewed evidence")
		}
		for _, e := range p.ReviewedEvidence {
			if e.ID == "" || e.Date == "" || e.File == "" || len(e.SHA256) != 64 {
				panic("coding review has incomplete immutable evidence identity")
			}
		}
		seen[p.PacketID] = true
		parties[p.PartyID]++
		agreement[p.Agreement]++
		switch p.Resolution {
		case "changed":
			changed++
		case "kept":
			kept++
		case "unresolved":
			unresolved++
		default:
			panic("unknown coding review resolution: " + p.Resolution)
		}
	}
	if len(parties) != 9 {
		panic("coding review must cover nine parties")
	}
	for party, n := range parties {
		if n != 2 {
			panic(fmt.Sprintf("coding review party %s has %d cells, want 2", party, n))
		}
	}
	if agreement["exact"] != m.ExactAgreement ||
		agreement["same_direction"] != m.SameDirection ||
		agreement["substantive"] != m.SubstantiveDisagreement ||
		changed != m.ChangedPositions || kept != m.KeptPositions ||
		unresolved != m.UnresolvedPositions {
		panic("coding review summary counts do not match its packets")
	}
	if m.SourceRevision == "" || len(m.InputSHA256) != 6 {
		panic("coding review lacks source revision or input hashes")
	}
	for name, digest := range m.InputSHA256 {
		if len(digest) != 64 {
			panic("coding review has an invalid input hash for " + name)
		}
	}
	return review
}

func CodingReviewData() CodingReview { return codingReviewDoc }
