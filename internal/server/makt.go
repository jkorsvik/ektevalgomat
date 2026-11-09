package server

import (
	"fmt"
	"math"
	"sort"

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
	ID           string
	Short        string
	Name         string
	Seats        int
	SeatShare    string
	SeatBarClass string
	Proposals    int
	PropShare    string
	PropBarClass string
	PerSeat      string
}

type hemicycleSeat struct {
	X     string
	Y     string
	Party string
	Short string
}

type maktAudit struct {
	Published  string
	Title      string
	Categories string
	URL        string
}

type maktView struct {
	Meta           pageMeta
	SnapshotDate   string
	Rows           []maktRow
	HemicycleSeats []hemicycleSeat
	SeatTotal      int
	ProposalTotal  int
	Session        string
	Audits         []maktAudit
	AuditCount     int
	// Refusals, so a reader sees the edges of the map rather than assuming them.
	CommitteeRefusal string
	ArchiveRefusal   string
}

// pctStepClass rounds a percentage in 0..100 to the nearest 5% step class ("pct-0" .. "pct-100").
func pctStepClass(pct float64) string {
	step := int(math.Round(pct/5.0)) * 5
	if step < 0 {
		step = 0
	}
	if step > 100 {
		step = 100
	}
	return fmt.Sprintf("pct-%d", step)
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

	maxSeat := 0
	maxProp := 0
	proposalSum := 0
	for _, id := range data.SeatRanking() {
		if seats[id] > maxSeat {
			maxSeat = seats[id]
		}
		if proposals[id] > maxProp {
			maxProp = proposals[id]
		}
		proposalSum += proposals[id]
	}

	var rows []maktRow
	for _, id := range data.SeatRanking() {
		share := ""
		if total > 0 {
			share = percent(100 * float64(seats[id]) / float64(total))
		}
		propShare := ""
		if proposalSum > 0 {
			propShare = percent(100 * float64(proposals[id]) / float64(proposalSum))
		}
		perSeat := ""
		if seats[id] > 0 && proposals[id] > 0 {
			perSeat = percent(float64(proposals[id]) / float64(seats[id]) * 100)
		}
		seatBar := "pct-0"
		if maxSeat > 0 {
			seatBar = pctStepClass(100 * float64(seats[id]) / float64(maxSeat))
		}
		propBar := "pct-0"
		if maxProp > 0 {
			propBar = pctStepClass(100 * float64(proposals[id]) / float64(maxProp))
		}
		rows = append(rows, maktRow{
			ID:           id,
			Short:        short[id],
			Name:         name[id],
			Seats:        seats[id],
			SeatShare:    share,
			SeatBarClass: seatBar,
			Proposals:    proposals[id],
			PropShare:    propShare,
			PropBarClass: propBar,
			PerSeat:      perSeat,
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

	return maktView{
		Meta: meta("Makt · seter, forslag og revisjon",
			"Hvem som har setene, hva hvert parti la på bordet, og hva Riksrevisjonen har publisert.",
			"/makt"),
		SnapshotDate:     snapshotDate,
		Rows:             rows,
		HemicycleSeats:   buildHemicycleSeats(seats, short),
		SeatTotal:        total,
		ProposalTotal:    len(data.Proposals()),
		Session:          data.StortingSession(),
		Audits:           audits,
		AuditCount:       len(audits),
		CommitteeRefusal: committeeRefusal,
		ArchiveRefusal:   archiveRefusal,
	}
}

// buildHemicycleSeats lays out the Storting seats across 6 concentric semicircular
// rings (21+24+27+30+32+35 = 169) and assigns parties in parliamentary seating
// order from left to right (R, SV, Ap, Sp, MDG, V, KrF, H, FrP) by angle.
func buildHemicycleSeats(seats map[string]int, short map[string]string) []hemicycleSeat {
	ringCounts := []int{21, 24, 27, 30, 32, 35}
	ringRadii := []float64{82, 102, 122, 142, 162, 182}
	const cx, cy = 200.0, 196.0

	type pos struct {
		x, y  float64
		angle float64
		ring  int
	}
	var positions []pos
	for rIdx, count := range ringCounts {
		radius := ringRadii[rIdx]
		for i := 0; i < count; i++ {
			frac := 0.5
			if count > 1 {
				frac = float64(i) / float64(count-1)
			}
			// Angle from pi (left) to 0 (right)
			theta := math.Pi * (1.0 - frac)
			x := cx + radius*math.Cos(theta)
			y := cy - radius*math.Sin(theta)
			positions = append(positions, pos{x: x, y: y, angle: theta, ring: rIdx})
		}
	}
	// Sort from left (theta = pi) to right (theta = 0), inner ring first on ties.
	sort.Slice(positions, func(i, j int) bool {
		if math.Abs(positions[i].angle-positions[j].angle) > 1e-6 {
			return positions[i].angle > positions[j].angle
		}
		return positions[i].ring < positions[j].ring
	})

	parliamentOrder := []string{"r", "sv", "ap", "sp", "mdg", "v", "krf", "h", "frp"}
	var partyList []string
	for _, id := range parliamentOrder {
		for k := 0; k < seats[id]; k++ {
			partyList = append(partyList, id)
		}
	}
	// Append any remaining seats if total > len(partyList)
	for id, n := range seats {
		found := false
		for _, p := range parliamentOrder {
			if p == id {
				found = true
				break
			}
		}
		if !found {
			for k := 0; k < n; k++ {
				partyList = append(partyList, id)
			}
		}
	}

	limit := len(positions)
	if len(partyList) < limit {
		limit = len(partyList)
	}
	out := make([]hemicycleSeat, 0, limit)
	for i := 0; i < limit; i++ {
		p := partyList[i]
		out = append(out, hemicycleSeat{
			X:     fmt.Sprintf("%.1f", positions[i].x),
			Y:     fmt.Sprintf("%.1f", positions[i].y),
			Party: p,
			Short: short[p],
		})
	}
	return out
}
