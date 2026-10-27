package server

import (
	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The power view: who holds the seats, what each party put on the table, and what
// the auditor published. Everything is derived from the three indexes at request
// time.
//
// The page exists because the three layers were ingested and unreadable. It puts
// seats beside proposals on purpose: seats are the resource a party was given,
// proposals are what it did with it, and neither number means much alone. Ap holds
// the most seats and files the fewest proposals; FrP holds the second-most and files
// the most. That contrast is the reason to show them in one table rather than three.
//
// What the page cannot show is stated on it: committee membership and alternates are
// not in the source, and the audit list is the recent reports rather than the whole
// archive.
type maktRow struct {
	ID        string
	Short     string
	Name      string
	Seats     int
	SeatShare string
	Proposals int
	PerSeat   string
}

type maktAudit struct {
	Published  string
	Title      string
	Categories string
	URL        string
}

type maktView struct {
	Meta          pageMeta
	SnapshotDate  string
	Rows          []maktRow
	SeatTotal     int
	ProposalTotal int
	Session       string
	Audits        []maktAudit
	AuditCount    int
	// Refusals, so a reader sees the edges of the map rather than assuming them.
	CommitteeRefusal string
	ArchiveRefusal   string
}

func buildMaktView(snapshotDate string) maktView {
	short := map[string]string{}
	name := map[string]string{}
	for _, p := range data.Store.Parties {
		short[p.ID] = p.Short
		name[p.ID] = p.Name
	}

	seats := data.SeatCounts()
	proposals := data.ProposalsPerParty()
	total := data.SeatTotal()

	var rows []maktRow
	for _, id := range data.SeatRanking() {
		share := ""
		if total > 0 {
			share = percent(100 * float64(seats[id]) / float64(total))
		}
		perSeat := ""
		if seats[id] > 0 && proposals[id] > 0 {
			perSeat = percent(float64(proposals[id]) / float64(seats[id]) * 100)
		}
		rows = append(rows, maktRow{
			ID:        id,
			Short:     short[id],
			Name:      name[id],
			Seats:     seats[id],
			SeatShare: share,
			Proposals: proposals[id],
			PerSeat:   perSeat,
		})
	}

	var audits []maktAudit
	for _, a := range data.AuditReports() {
		cats := ""
		for i, c := range a.Categories {
			if i > 0 {
				cats += ", "
			}
			cats += c
		}
		audits = append(audits, maktAudit{
			Published:  a.Published,
			Title:      a.Title,
			Categories: cats,
			URL:        a.URL,
		})
	}

	committeeRefusal, archiveRefusal := "", ""
	if r, ok := data.CommitteeMembershipRefusal(); ok {
		committeeRefusal = r.Reason
	}
	if r, ok := data.AuditArchiveRefusal(); ok {
		archiveRefusal = r.Reason
	}

	proposalTotal := 0
	for _, n := range proposals {
		proposalTotal += n
	}

	return maktView{
		Meta: meta("Makt · seter, forslag og revisjon",
			"Hvem som har setene, hva hvert parti la på bordet, og hva Riksrevisjonen har publisert.",
			"/makt"),
		SnapshotDate:     snapshotDate,
		Rows:             rows,
		SeatTotal:        total,
		ProposalTotal:    len(data.Proposals()),
		Session:          data.StortingSession(),
		Audits:           audits,
		AuditCount:       len(audits),
		CommitteeRefusal: committeeRefusal,
		ArchiveRefusal:   archiveRefusal,
	}
}
