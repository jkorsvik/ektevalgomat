package server

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// Page metadata (2026-10-13). Before this, only two pages of seventy-odd
// carried a meta description, none carried a canonical link, and none
// carried an Open Graph tag, so a shared link produced no preview card and
// a search result had no description. Found while checking the site was
// ready to be public.
//
// siteBase is the one canonical host. A page reached through any other name
// still declares this one, which is the point of a canonical link.
const siteBase = "https://ektevalgomat.no"

type pageMeta struct {
	Title       string
	Description string
	Canonical   string
	// Path is the route this page lives on. The navigation marks the current page
	// from it, so a reader can see where they are without reading the URL.
	Path string
}

func meta(title, description, path string) pageMeta {
	return pageMeta{
		Title:       title,
		Description: trimTo(description, 155),
		Canonical:   siteBase + path,
		Path:        path,
	}
}

// trimTo keeps a description near the length search engines show, and cuts
// on a word rather than mid-word.
func trimTo(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	cut := string([]rune(s)[:max])
	if i := strings.LastIndex(cut, " "); i > 0 {
		cut = cut[:i]
	}
	return cut + " …"
}

// statementSuffix says what a statement page is. A statement can be short —
// "Det bør innføres skatt på store arv." is 36 characters — and a
// description that is only the statement never says what the page offers.
const statementSuffix = " Se hva hvert parti har gjort med det, med kilde."

func statementMeta(s data.Statement) pageMeta {
	title := "Utsagn " + strconv.Itoa(s.Number)
	if s.ShortName != "" {
		title += ": " + s.ShortName
	}
	desc := trimTo(s.Text, 105) + statementSuffix
	return meta(title+" · Ekte valgomat", desc, "/utsagn/"+strconv.Itoa(s.Number))
}

func partyMeta(p data.Party, points int) pageMeta {
	desc := "Hva " + p.Name + " sier i sitt stortingsvalgprogram 2025–2029."
	if points > 0 {
		desc += " " + strconv.Itoa(points) + " utvalgte løfter, sitert med side."
	}
	return meta(p.Name+" · partiprogram · Ekte valgomat", desc, "/parti/"+p.ID)
}

// navLink is one destination in the site's navigation.
type navLink struct {
	Href  string
	Label string
}

// navLinks is the single definition of the site's navigation, and every page that a
// reader is meant to reach appears here. Before 2026-10-30 eight pages — Kartet,
// Stemmemønster, År for år, Rommet, Makt, Penger, Eierskap and Data — were reachable
// only from one line of small footer links, so the most substantial work on the site
// was invisible from the top of every page. A test walks the route table and fails if
// a page is missing here.
var navLinks = []navLink{
	{"/", "Valgomat"},
	{"/resultat", "Resultat"},
	{"/partier", "Partiene"},
	{"/kart", "Kartet"},
	{"/stemmer", "Stemmemønster"},
	{"/tidslinje", "År for år"},
	{"/rommet", "Rommet"},
	{"/budsjett", "Budsjettet"},
	{"/penger", "Penger"},
	{"/eierskap", "Eierskap"},
	{"/makt", "Makt"},
	{"/data", "Data"},
	{"/metode", "Metode"},
	{"/om", "Om"},
}
