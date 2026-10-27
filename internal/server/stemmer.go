package server

import (
	"sort"

	"github.com/jkorsvik/ektevalgomat/internal/analysis"
	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The vote-alignment view. Every number here is computed from data.Votes and
// data.Store at request time: there is no stored percentage, no cached matrix and
// nothing hand-written. A reader who wants to check a cell can open the votes it
// rests on, which are listed with their proposal text.
//
// The threshold matters more than the ranking. A pair compared on 6 votes can
// show 100 % and mean nothing, so a pair is only ranked once both parties took a
// side on at least minComparedForRanking votes, and every cell reports the
// denominator it was computed from.
const minComparedForRanking = 10

type matrixCell struct {
	A        string
	B        string
	AShort   string
	BShort   string
	Pct      int
	Compared int
	Agree    int
	HasData  bool
}

type stemmerParty struct {
	ID    string
	Short string
	Name  string
	Votes int
}

type stemmerVote struct {
	ID      int
	Forslag string
	Sides   []stemmerSide
}

type stemmerSide struct {
	Party string
	Side  string
}

type stemmerView struct {
	Meta          pageMeta
	SnapshotDate  string
	GeneratedAt   string
	VoteTotal     int
	VoteringsCite int
	Parties       []stemmerParty
	Matrix        [][]matrixCell
	Ranked        []matrixCell
	Most          matrixCell
	Least         matrixCell
	HasRanking    bool
	Votes         []stemmerVote
	Unanimous     int
	// Coverage: which storting sessions the index was built from. A matrix that
	// silently covers three of eighteen sessions reads as if it covers the record,
	// and the reader has no way to tell.
	Sessions        []data.VoteSession
	SessionCount    int
	VotesInSessions int
}

func buildStemmerView(ds *data.Dataset, snapshotDate string) stemmerView {
	votes := data.Votes()

	ids := make([]string, 0, len(ds.Parties))
	byID := make(map[string]data.Party, len(ds.Parties))
	for _, p := range ds.Parties {
		ids = append(ids, p.ID)
		byID[p.ID] = p
	}
	counts := analysis.VotesByParty(votes, ids)

	parties := make([]stemmerParty, 0, len(ids))
	for _, id := range ids {
		parties = append(parties, stemmerParty{
			ID: id, Short: byID[id].Short, Name: byID[id].Name, Votes: counts[id],
		})
	}

	pairs := analysis.VoteAgreement(votes, ids, 1)
	lookup := make(map[string]matrixCell, len(pairs))
	for _, p := range pairs {
		lookup[p.A+"/"+p.B] = matrixCell{A: p.A, B: p.B,
			AShort: byID[p.A].Short, BShort: byID[p.B].Short,
			Pct: int(p.Pct() + 0.5), Compared: p.Compared, Agree: p.Agree, HasData: p.Compared > 0}
	}

	matrix := make([][]matrixCell, len(ids))
	for i := range ids {
		matrix[i] = make([]matrixCell, len(ids))
		for j := range ids {
			if i == j {
				matrix[i][j] = matrixCell{A: ids[i], B: ids[j]}
				continue
			}
			key, flipped := ids[i]+"/"+ids[j], false
			if i > j {
				key, flipped = ids[j]+"/"+ids[i], true
			}
			c := lookup[key]
			if flipped {
				c.A, c.B = ids[i], ids[j]
				c.AShort, c.BShort = byID[ids[i]].Short, byID[ids[j]].Short
			}
			matrix[i][j] = c
		}
	}

	var ranked []matrixCell
	for _, p := range pairs {
		if p.Compared >= minComparedForRanking {
			ranked = append(ranked, lookup[p.A+"/"+p.B])
		}
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Pct != ranked[j].Pct {
			return ranked[i].Pct > ranked[j].Pct
		}
		return ranked[i].Compared > ranked[j].Compared
	})

	view := stemmerView{
		Meta: meta("Stemmemønster · hvem stemmer med hvem",
			"Hvor ofte to partier stemte likt på de samme voteringene, regnet ut fra Stortingets voteringsresultat, med nevneren synlig i hver celle.",
			"/stemmer"),
		SnapshotDate:    snapshotDate,
		GeneratedAt:     data.VotesGeneratedAt(),
		VoteTotal:       len(votes),
		Sessions:        data.VoteSessions(),
		SessionCount:    len(data.VoteSessions()),
		VotesInSessions: data.VotesWithSession(),
		VoteringsCite:   len(votes),
		Parties:         parties,
		Matrix:          matrix,
		Ranked:          ranked,
	}
	if len(ranked) > 0 {
		view.Most, view.Least, view.HasRanking = ranked[0], ranked[len(ranked)-1], true
	}

	for _, v := range votes {
		if v.Unanimous {
			view.Unanimous++
		}
		row := stemmerVote{ID: v.VoteringID, Forslag: v.Forslag}
		for _, id := range ids {
			if s, ok := v.Parties[id]; ok && s != "absent" && s != "" {
				row.Sides = append(row.Sides, stemmerSide{Party: byID[id].Short, Side: s})
			}
		}
		view.Votes = append(view.Votes, row)
	}
	return view
}
