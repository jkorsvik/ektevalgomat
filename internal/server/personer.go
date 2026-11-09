package server

import (
	"sort"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The person view. Every field is read from internal/data/content/people/, which
// is written by tools/autopilot/lobby/promote_people.py only after the privacy
// gate and an adversarial verifier pass. Nothing here is computed from the
// evidence set: a role is a public-office fact, not a position.
//
// Two rules the page must keep:
//   - A verv date that is the register's snapshot is not a role start. It is
//     rendered as "registrert <dato>" and never as "fra".
//   - A role with no end date is "ikke avsluttet i registeret", not "sitter nå":
//     the register is a snapshot, so silence about an end is not proof of one.

type personRoleView struct {
	Label        string
	Organisation string
	Post         string
	From         string
	To           string
	Open         bool
	OpenLabel    string
	Snapshot     bool
	SourceURL    string
	SourceNote   string
}

type personRow struct {
	ID            string
	Name          string
	Party         string
	PartyShort    string
	PartyBarClass string
	Roles         []personRoleView
	VervCount     int
	HasGovernment bool
}

type partyCount struct {
	Party string
	Short string
	Bar   string
	Count int
}

type endedRole struct {
	Party      string
	PartyShort string
	Bar        string
	Name       string
	Text       string
	Seen       string
	GoneBy     string
}

type personerView struct {
	Meta           pageMeta
	SnapshotDate   string
	People         []personRow
	Parties        []partyCount
	Total          int
	WithVerv       int
	WithGovernment int
	RoleTotal      int
	RegisterDate   string
	History        []endedRole
	HistoryFrom    string
	HistoryTo      string
	HistoryEnded   int
	HistoryNew     int
	HistoryPeople  int
}

var roleLabels = map[string]string{
	"stortingsrepresentant": "Stortingsrepresentant",
	"verv":                  "Verv",
	"statsrad":              "Statsråd",
	"statssekretar":         "Statssekretær",
	"selskapsinteresse":     "Selskapsinteresse",
	"styrerverv":            "Styreverv",
	"daglig_leder":          "Daglig leder",
	"verv_i_organisasjon":   "Verv i organisasjon",
	"hoyringsdeltaker":      "Høringsdeltaker",
}

func roleLabel(role string) string {
	if l, ok := roleLabels[role]; ok {
		return l
	}
	return strings.ReplaceAll(role, "_", " ")
}

func buildPersonerView(ds *data.Dataset, snapshotDate string) personerView {
	byID := make(map[string]data.Party, len(ds.Parties))
	for _, p := range ds.Parties {
		byID[p.ID] = p
	}

	var rows []personRow
	partyTally := map[string]int{}
	withVerv, withGov, roleTotal := 0, 0, 0

	for _, p := range data.People() {
		party := byID[p.PartyID]
		row := personRow{
			ID:            p.ID,
			Name:          p.Name,
			Party:         party.Name,
			PartyShort:    party.Short,
			PartyBarClass: "bar-" + p.PartyID,
		}
		for _, r := range p.Roles {
			openLabel := "ikke avsluttet i registeret"
			if r.Role == "stortingsrepresentant" {
				// A seat is not a register entry: an empty end means the person
				// sits now, not that the register went quiet.
				openLabel = "sitter"
			}
			rv := personRoleView{
				Label:        roleLabel(r.Role),
				Organisation: r.Organisation,
				Post:         r.Post,
				From:         r.From,
				To:           r.To,
				Open:         r.Holding(),
				OpenLabel:    openLabel,
				Snapshot:     r.SnapshotDated(),
				SourceURL:    r.SourceURL,
				SourceNote:   r.SourceNote,
			}
			row.Roles = append(row.Roles, rv)
			if r.Role == "verv" {
				row.VervCount++
			}
			if r.Role == "statsrad" || r.Role == "statssekretar" {
				row.HasGovernment = true
			}
			roleTotal++
		}
		if row.VervCount > 0 {
			withVerv++
		}
		if row.HasGovernment {
			withGov++
		}
		partyTally[p.PartyID]++
		rows = append(rows, row)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Party != rows[j].Party {
			return rows[i].Party < rows[j].Party
		}
		return rows[i].Name < rows[j].Name
	})

	var parties []partyCount
	for id, n := range partyTally {
		parties = append(parties, partyCount{
			Party: byID[id].Name,
			Short: byID[id].Short,
			Bar:   "bar-" + id,
			Count: n,
		})
	}
	// Count descending, then name ascending. Without the tie-breaker the order of
	// equal counts came from the map iteration and changed on every request.
	sort.SliceStable(parties, func(i, j int) bool {
		if parties[i].Count != parties[j].Count {
			return parties[i].Count > parties[j].Count
		}
		return parties[i].Party < parties[j].Party
	})

	hist := data.RegisterHistoryDiff()
	apiToID := map[string]string{"A": "ap", "FrP": "frp", "H": "h", "SV": "sv", "Sp": "sp",
		"R": "rodt", "MDG": "mdg", "KrF": "krf", "V": "v"}
	var ended []endedRole
	for _, c := range hist.Ended {
		id := apiToID[c.Party]
		ended = append(ended, endedRole{
			Party:      byID[id].Name,
			PartyShort: byID[id].Short,
			Bar:        "bar-" + id,
			Name:       c.Name,
			Text:       c.Text,
			Seen:       c.Seen,
			GoneBy:     c.GoneBy,
		})
	}
	from, to := "", ""
	if len(hist.ComparedBetween) == 2 {
		from, to = hist.ComparedBetween[0], hist.ComparedBetween[1]
	}

	return personerView{
		Meta: meta("Personer · verv og offentlige roller",
			"Hvem som sitter i hvilke verv: stortingsrepresentanter med daterte roller fra Stortingets vervregister og regjeringsapparatet. Hver rolle har en kilde.",
			"/personer"),
		SnapshotDate:   snapshotDate,
		People:         rows,
		Parties:        parties,
		Total:          len(rows),
		WithVerv:       withVerv,
		WithGovernment: withGov,
		RoleTotal:      roleTotal,
		RegisterDate:   data.VervSnapshotDate(),
		History:        ended,
		HistoryFrom:    from,
		HistoryTo:      to,
		HistoryEnded:   hist.Counts.RolesEnded,
		HistoryNew:     hist.Counts.RolesNew,
		HistoryPeople:  hist.Counts.PeopleInBoth,
	}
}
