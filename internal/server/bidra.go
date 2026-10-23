package server

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/jkorsvik/ektevalgomat/internal/data"
	"github.com/jkorsvik/ektevalgomat/internal/scoring"
)

// /bidra answers one question: what could I do, right now, that the project
// would actually use? It is generated from the dataset, never from a list
// someone maintains, because a "help us" page that drifts away from the real
// gaps sends volunteers at work that is already done — the fastest way to
// lose them.
//
// Two kinds of gap, and they are different jobs:
//
//   - no source at all for a party on a statement: the volunteer has to go
//     find one.
//   - evidence collected but no coded position: the source is already
//     linked on the page, so the volunteer reads it and says what it says.
//
// Collapsing them into one "missing" count would hide that difference, so
// they carry separate labels and separate prefilled issues.

type bidraCell struct {
	Party         data.Party
	NoSource      bool
	EvidenceCount int
	IssueURL      string
}

type bidraGap struct {
	Statement data.Statement
	Path      string
	Cells     []bidraCell
	// AllParties is true when no party has a coded position on this
	// statement: the whole statement is unread rather than one cell.
	AllParties bool
	PartyCount int
}

type bidraTier struct {
	Tier  string
	Label string
	Count int
}

type bidraPage struct {
	Meta           pageMeta
	SnapshotDate   string
	Gaps           []bidraGap
	GapStatements  int
	GapCells       int
	NoSourceCells  int
	UncodedCells   int
	StatementCount int
	PartyCount     int
	EvidenceCount  int
	QuotedCount    int
	QuoteLessCount int
	WeakTiers      []bidraTier
	WeakCount      int
	Contributing   string
}

func buildBidraPage(ds *data.Dataset, snapshotDate string) bidraPage {
	page := bidraPage{
		Meta: meta("Bidra · Ekte valgomat",
			"Hva som mangler i datasettet akkurat nå, og hvordan du melder fra om én kilde: hvilke partier som mangler kilde på hvilke utsagn, og hvilke kilder som er svake.",
			"/bidra"),
		SnapshotDate:   snapshotDate,
		StatementCount: len(ds.Statements),
		PartyCount:     len(ds.Parties),
		EvidenceCount:  len(ds.Evidence),
		Contributing:   repoURL + "/blob/main/CONTRIBUTING.md",
	}

	tierCount := map[string]int{}
	for _, e := range ds.Evidence {
		tierCount[e.SourceTier]++
		if e.VerbatimExcerpt != "" {
			page.QuotedCount++
		} else {
			page.QuoteLessCount++
		}
	}
	for _, tier := range []string{"D", "C", "B"} {
		if tierCount[tier] == 0 {
			continue
		}
		page.WeakTiers = append(page.WeakTiers, bidraTier{
			Tier: tier, Label: sourceTierLabel(tier), Count: tierCount[tier],
		})
		page.WeakCount += tierCount[tier]
	}

	statements := append([]data.Statement(nil), ds.Statements...)
	sort.Slice(statements, func(i, j int) bool { return statements[i].Number < statements[j].Number })

	for _, s := range statements {
		gap := bidraGap{
			Statement:  s,
			Path:       "/utsagn/" + strconv.Itoa(s.Number),
			PartyCount: len(ds.Parties),
		}
		withPosition := 0
		for _, p := range ds.Parties {
			cell := scoring.ComputeCell(ds, s.ID, p.ID)
			if cell.PartyPosition != nil {
				withPosition++
				continue
			}
			c := bidraCell{
				Party:         p,
				NoSource:      !cell.HasEvidence(),
				EvidenceCount: cell.EvidensstyrkeCount,
			}
			if c.NoSource {
				page.NoSourceCells++
				c.IssueURL = issueURL("rapporter-kilde.md",
					fmt.Sprintf("Kilde mangler: %s / %s", s.ID, p.Name),
					fmt.Sprintf("Kilden som mangler, og hva den sier:\n\n- Utsagn: %s (%s)\n- Parti: %s\n- Kilden jeg fant: \n- Hvor i kilden: \n\nUtsagnet:\n\"%s\"\n",
						s.ID, gap.Path, p.Name, s.Text))
			} else {
				page.UncodedCells++
				c.IssueURL = issueURL("rapporter-kilde.md",
					fmt.Sprintf("Tolkning mangler: %s / %s", s.ID, p.Name),
					fmt.Sprintf("Kilden ligger allerede på siden, men ingen posisjon er kodet. Hva sier den?\n\n- Utsagn: %s (%s)\n- Parti: %s\n- Kilden sier (enig/uenig, og hvorfor): \n\nUtsagnet:\n\"%s\"\n",
						s.ID, gap.Path, p.Name, s.Text))
			}
			gap.Cells = append(gap.Cells, c)
		}
		if len(gap.Cells) == 0 {
			continue
		}
		gap.AllParties = withPosition == 0
		sort.Slice(gap.Cells, func(i, j int) bool { return gap.Cells[i].Party.Name < gap.Cells[j].Party.Name })
		page.Gaps = append(page.Gaps, gap)
		page.GapCells += len(gap.Cells)
	}
	page.GapStatements = len(page.Gaps)
	return page
}
