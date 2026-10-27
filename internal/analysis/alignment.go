// Package analysis derives the views that need more than one evidence row.
//
// Everything here is computed from data.Store and data.Votes at request time.
// No number in this package is written down, and none is cached: a figure that
// appears on a page must be reproducible from the dataset alone, or it is a
// claim the project cannot defend.
package analysis

import "github.com/jkorsvik/ektevalgomat/internal/data"

// Pair is how often two parties took the same side on the same vote.
//
// Compared counts only votes where BOTH parties actually took a side: a party
// recorded as absent, or split, is not a data point for agreement, because
// "did not vote" and "voted both ways" are not the same as "disagreed".
// Reporting Compared next to Agree is what makes the percentage honest: 3 of 3
// is not stronger evidence than 40 of 50, and a reader has to be able to see
// which one they are looking at.
type Pair struct {
	A, B     string
	Compared int
	Agree    int
}

// Pct is Agree as a percentage of Compared, or 0 when nothing was compared.
func (p Pair) Pct() float64 {
	if p.Compared == 0 {
		return 0
	}
	return float64(p.Agree) * 100 / float64(p.Compared)
}

// Disagree is Compared minus Agree.
func (p Pair) Disagree() int { return p.Compared - p.Agree }

// sideOf returns the party's side on one vote, and whether it counts as a side.
func sideOf(v data.Vote, party string) (string, bool) {
	s, ok := v.Parties[party]
	if !ok || s == "absent" || s == "" {
		return "", false
	}
	return s, true
}

// VoteAgreement returns one Pair per party combination, ordered by the party
// order given (which is the same order the rest of the site uses), and only for
// pairs that were compared at least minCompared times.
//
// "split" IS a side here: a party that split on a vote still has a position on
// the question the vote asked, and treating a split as agreement with everyone
// would overstate consensus. It counts as agreeing only with another split.
func VoteAgreement(votes []data.Vote, parties []string, minCompared int) []Pair {
	var out []Pair
	for i := 0; i < len(parties); i++ {
		for j := i + 1; j < len(parties); j++ {
			p := Pair{A: parties[i], B: parties[j]}
			for _, v := range votes {
				a, okA := sideOf(v, p.A)
				b, okB := sideOf(v, p.B)
				if !okA || !okB {
					continue
				}
				p.Compared++
				if a == b {
					p.Agree++
				}
			}
			if p.Compared >= minCompared {
				out = append(out, p)
			}
		}
	}
	return out
}

// MostAligned returns the single pair that agreed most often, and the pair that
// agreed least, among those compared at least minCompared times. ok is false
// when no pair qualified, so a caller never renders an empty "most aligned".
func MostAligned(pairs []Pair) (most Pair, least Pair, ok bool) {
	for i, p := range pairs {
		if i == 0 {
			most, least, ok = p, p, true
			continue
		}
		if betterThan(p, most) {
			most = p
		}
		if betterThan(least, p) {
			least = p
		}
	}
	return most, least, ok
}

// betterThan orders two pairs by how much they support the claim they make.
//
// The percentage comes first, but a tie must not be broken by whichever pair the
// loop happened to see first. Among equals the pair compared on more votes wins,
// because 77 % of 5878 is a stronger statement than 77 % of 12. Identifiers decide
// the rest so the answer does not depend on map iteration order.
//
// This exists because the page and this function disagreed: the page sorted by
// percentage then count, this one kept the first seen, and they named different
// pairs as "most aligned". Two definitions of one headline is a defect whichever
// one renders.
func betterThan(a, b Pair) bool {
	if a.Pct() != b.Pct() {
		return a.Pct() > b.Pct()
	}
	if a.Compared != b.Compared {
		return a.Compared > b.Compared
	}
	if a.A != b.A {
		return a.A < b.A
	}
	return a.B < b.B
}

// VotesByParty counts, per party, how many indexed votes it took a side on. It
// is the denominator a reader needs before trusting any agreement figure.
func VotesByParty(votes []data.Vote, parties []string) map[string]int {
	counts := make(map[string]int, len(parties))
	for _, party := range parties {
		for _, v := range votes {
			if _, ok := sideOf(v, party); ok {
				counts[party]++
			}
		}
	}
	return counts
}
