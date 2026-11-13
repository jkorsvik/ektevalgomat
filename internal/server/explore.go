package server

import (
	"fmt"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The Utforsk pages answer different reader questions; they are not chapters in
// a course. This file defines the question/source/gap contract at the top of each
// page and the related questions at the bottom. The /utforsk hub groups the same
// pages by reader intent, so nobody has to follow nine pages in a fixed order.
//
// The contract is derived from the artefacts rather than typed, so it cannot drift
// away from the data the page itself is drawing. Tests cover every menu page and
// prove that the hub and related-question links actually render.

// exploreStep is one page's public explanation.
type exploreStep struct {
	Href     string
	Label    string
	Question string
	Source   string
	Missing  string
	// Related contains the other pages that answer the same broad reader
	// question. It is deliberately not a single prescribed next step.
	Related []exploreStep
}

// explorePages is the set of analysis pages, in menu order.
var explorePages = []exploreStep{
	{
		Href:     "/kart",
		Label:    "Kartet",
		Question: "Hvor enige er partiene, utsagn for utsagn?",
	},
	{
		Href:     "/rommet",
		Label:    "Uenigheter",
		Question: "Hvilke konkrete saker er partiene uenige om, og hvor langt fra hverandre står de?",
	},
	{
		Href:     "/stemmer",
		Label:    "Stemmemønster",
		Question: "Hvilke partier stemmer sammen i salen, og hvilke gjør det ikke?",
	},
	{
		Href:     "/tidslinje",
		Label:    "År for år",
		Question: "Når i tid ligger kildene vi bygger på?",
	},
	{
		Href:     "/personer",
		Label:    "Personer",
		Question: "Hvem sitter på hvilke verv, i hvilke komiteer, og hvordan stemmer de?",
	},
	{
		Href:     "/makt",
		Label:    "Makt",
		Question: "Hvem har makt til å bestemme, og hvor mange er de?",
	},
	{
		Href:     "/budsjett",
		Label:    "Budsjettet",
		Question: "Hva vil partiene bruke penger på?",
	},
	{
		Href:     "/penger",
		Label:    "Penger",
		Question: "Hvor kommer partienes penger fra?",
	},
	{
		Href:     "/eierskap",
		Label:    "Eierskap",
		Question: "Hva eier staten, og hvem leder det?",
	},
}

// exploreInfo returns a page's contract and the related pages in its reader
// group. Related pages are built without recursion: they need their own contract,
// but not another copy of their own related list.
func exploreInfo(href string) exploreStep {
	var found exploreStep
	for _, candidate := range explorePages {
		if candidate.Href == href {
			found = candidate
			break
		}
	}
	if found.Href == "" {
		return exploreStep{}
	}
	found.Source = exploreSource(href)
	found.Missing = exploreGap(href)
	for _, definition := range exploreGroupDefinitions {
		inGroup := false
		for _, member := range definition.Hrefs {
			if member == href {
				inGroup = true
				break
			}
		}
		if !inGroup {
			continue
		}
		for _, member := range definition.Hrefs {
			if member == href {
				continue
			}
			for _, candidate := range explorePages {
				if candidate.Href == member {
					candidate.Source = exploreSource(member)
					candidate.Missing = exploreGap(member)
					found.Related = append(found.Related, candidate)
				}
			}
		}
	}
	return found
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
