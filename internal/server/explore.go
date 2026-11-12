package server

import (
	"fmt"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The Utforsk pages were nine separate doors. A reader landed on one from the
// menu, the page answered its question in its own visual language, and there was
// no way onward to the page that answers the neighbouring question — three of the
// nine had no link to any other Utforsk page at all.
//
// This file is the spine: one ordered walk in three questions, a fixed contract at
// the top of every page saying what it answers, where the numbers come from and
// what is missing, and one handoff at the bottom naming the next question. The
// contract is derived from the artefacts rather than typed, so it cannot drift
// away from the data the page itself is drawing.
//
// A test walks navLinks and fails if an Utforsk page is missing from the spine, if
// a step has an empty contract line, or if a step points at a page that is not in
// the walk.

// exploreStep is one page's place in the walk.
type exploreStep struct {
	Href     string
	Label    string
	Question string
	Source   string
	Missing  string
	// Next is the neighbouring question, not the next menu item: the walk reads
	// "what do they mean" into "what did they do" into "what does it cost".
	NextHref     string
	NextLabel    string
	NextQuestion string
}

// exploreSpine is the ordered walk. Order is the argument: meaning, then record,
// then money.
var exploreSpine = []exploreStep{
	{
		Href:     "/kart",
		Label:    "Kartet",
		Question: "Hvor enige er partiene, utsagn for utsagn?",
		NextHref: "/rommet",
	},
	{
		Href:     "/rommet",
		Label:    "Rommet",
		Question: "Hvilke dimensjoner skiller partiene når alle utsagnene sees samlet?",
		NextHref: "/stemmer",
	},
	{
		Href:     "/stemmer",
		Label:    "Stemmemønster",
		Question: "Hvilke partier stemmer sammen i salen, og hvilke gjør det ikke?",
		NextHref: "/tidslinje",
	},
	{
		Href:     "/tidslinje",
		Label:    "År for år",
		Question: "Når i tid ligger kildene vi bygger på?",
		NextHref: "/personer",
	},
	{
		Href:     "/personer",
		Label:    "Personer",
		Question: "Hvem sitter på hvilke verv, i hvilke komiteer, og hvordan stemmer de?",
		NextHref: "/makt",
	},
	{
		Href:     "/makt",
		Label:    "Makt",
		Question: "Hvem har makt til å bestemme, og hvor mange er de?",
		NextHref: "/budsjett",
	},
	{
		Href:     "/budsjett",
		Label:    "Budsjettet",
		Question: "Hva vil partiene bruke penger på?",
		NextHref: "/penger",
	},
	{
		Href:     "/penger",
		Label:    "Penger",
		Question: "Hvor kommer partienes penger fra?",
		NextHref: "/eierskap",
	},
	{
		Href:     "/eierskap",
		Label:    "Eierskap",
		Question: "Hva eier staten, og hvem leder det?",
	},
}

// exploreInfo returns a page's contract and handoff, with the source and gap lines
// built from the artefacts.
func exploreInfo(href string) exploreStep {
	for i, s := range exploreSpine {
		if s.Href != href {
			continue
		}
		s.Source = exploreSource(href)
		s.Missing = exploreGap(href)
		if i+1 < len(exploreSpine) {
			next := exploreSpine[i+1]
			s.NextHref = next.Href
			s.NextLabel = next.Label
			s.NextQuestion = next.Question
		} else {
			// The walk closes where a reader's question starts: having seen what
			// the parties mean, did and cost, the next thing to do is answer the
			// valgomat. Ending on the last page with no way out is the dead end
			// this whole file exists to remove.
			s.NextHref = "/"
			s.NextLabel = "Valgomat"
			s.NextQuestion = "Hvilket parti er du mest enig med?"
		}
		return s
	}
	return exploreStep{}
}

// exploreSource states, in one line, what the page draws on.
func exploreSource(href string) string {
	total, withQuote := quoteCounts(data.Store)
	pv := data.PersonVoteSummary()
	ci := data.CommitteeSummary()
	switch href {
	case "/kart", "/rommet":
		return fmt.Sprintf("%s påstander om %s partier, hver med en kilde. %s av dem har et ordrett sitat lagret.",
			thousands(total), thousands(len(data.Store.Parties)), thousands(withQuote))
	case "/stemmer":
		return fmt.Sprintf("%s voteringer. %s av divisjonene oppgir stemmen til hver representant.",
			thousands(len(data.Votes())), thousands(pv.DivisionsWithRows))
	case "/tidslinje":
		return fmt.Sprintf("Datoene på de %s påstandene og %s voteringene, ikke datoene de ble publisert.",
			thousands(len(data.Store.Evidence)), thousands(len(data.Votes())))
	case "/personer":
		return fmt.Sprintf("%s representanter med %s daterte roller, %s komitéplass og %s navngitte stemmer.",
			thousands(len(data.People())), thousands(countRoles()), thousands(ci.Counts["memberships"]),
			thousands(pv.RepresentativeRows))
	case "/makt":
		return fmt.Sprintf("Stortingssalen og regjeringsapparatet: %s representanter, %s statsråd, %s partier.",
			thousands(len(data.People())), thousands(countPost("statsrad")), thousands(len(data.Store.Parties)))
	case "/budsjett":
		return "Partienes egne budsjettforslag og regjeringens forslag, per departement."
	case "/penger":
		return "SSB tabell 11233: statlig støtte og bidrag per parti og år."
	case "/eierskap":
		return "Statens eierposter og selskapenes ledelse, fra regjeringens eierrapport."
	}
	return ""
}

// exploreGap names what the page cannot show. A page that stays silent about its
// own limits reads as complete.
func exploreGap(href string) string {
	total, withQuote := quoteCounts(data.Store)
	without := total - withQuote
	pv := data.PersonVoteSummary()
	ci := data.CommitteeSummary()
	switch href {
	case "/kart", "/rommet":
		return fmt.Sprintf("%s påstander har ingen lagret sitat, og en posisjon uten sitat kan ikke etterprøves uten å åpne kilden selv.",
			thousands(without))
	case "/stemmer":
		return fmt.Sprintf("%s av %s divisjoner er tomme hos kilden selv, så de sier ingenting om hvem som stemte hva.",
			thousands(pv.DivisionsChecked-pv.DivisionsWithRows), thousands(pv.DivisionsChecked))
	case "/tidslinje":
		return "En påstand kan mangle dato. Da teller den ikke som et år, uansett hvor god kilden er."
	case "/personer":
		if len(ci.Unpublished) > 0 {
			return fmt.Sprintf("%d organer har ingen publisert medlemsliste, så noen representanter står uten komitéplass her.",
				len(ci.Unpublished))
		}
		return "Komitéplass er et øyeblikksbilde uten start- eller sluttdato."
	case "/makt":
		return "Antall seter sier ingenting om hva makten faktisk brukes til."
	case "/budsjett":
		return "Budsjettall er partienes egne forslag, ikke vedtatte bevilgninger."
	case "/penger":
		return "Tallene sier hvor mye partiet får, ikke hvem som ga det."
	case "/eierskap":
		return "En eierpost sier ikke hvordan selskapet styres i praksis."
	}
	return ""
}

// countRoles is the number of dated roles across all person records.
func countRoles() int {
	n := 0
	for _, p := range data.People() {
		n += len(p.Roles)
	}
	return n
}

// countPost counts roles of one kind, such as statsrad.
func countPost(kind string) int {
	n := 0
	for _, p := range data.People() {
		for _, r := range p.Roles {
			if r.Role == kind {
				n++
			}
		}
	}
	return n
}
