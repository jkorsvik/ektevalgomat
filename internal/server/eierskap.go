package server

import (
	"strconv"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The ownership view: what the state owns, how much, and what governs it.
// Everything is derived from the ownership index at request time.
//
// The page exists because a statement about state ownership otherwise has a
// position and no object. It is also the one page where a reader can check a claim
// against the register in one click, so every row carries the organisation number
// that makes that possible.
type eierskapRow struct {
	Company      string
	Share        string
	Category     string
	CategoryWhy  string
	LegalForm    string
	Employees    string
	Industry     string
	Municipality string
	Orgnr        string
	SpecialLaw   bool
	BrregURL     string
}

type eierskapView struct {
	Meta          pageMeta
	SnapshotDate  string
	Rows          []eierskapRow
	Companies     int
	WhollyOwned   int
	Majority      int
	Employees     string
	WithEmployees int
	SpecialLaw    int
	SourceName    string
	SourceURL     string
	RetrievedAt   string
}

// thousands formats an integer with spaces as group separators, the Norwegian
// convention, so 70358 reads as 70 358.
func thousands(n int) string {
	s := strconv.Itoa(n)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// share formats a percentage with a Norwegian decimal comma, dropping a trailing
// ",0" because the source writes whole percentages that way.
func share(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	return strings.Replace(s, ".", ",", 1)
}

func buildEierskapView(snapshotDate string) eierskapView {
	companies := data.StateOwnedCompanies()
	categories := data.OwnershipCategories()
	summary := data.SummariseOwnership()

	rows := make([]eierskapRow, 0, len(companies))
	for _, c := range companies {
		employees := ""
		if c.Employees != nil {
			employees = thousands(*c.Employees)
		}
		rows = append(rows, eierskapRow{
			Company:      c.Company,
			Share:        share(c.Share),
			Category:     c.Category,
			CategoryWhy:  categories[c.Category],
			LegalForm:    c.LegalForm,
			Employees:    employees,
			Industry:     c.Industry,
			Municipality: c.Municipality,
			Orgnr:        c.Orgnr,
			SpecialLaw:   c.SpecialLawCompany,
			BrregURL:     "https://virksomhet.brreg.no/nb/oppslag/enheter/" + c.Orgnr,
		})
	}

	name, url, retrieved, _ := data.OwnershipSourceInfo()
	return eierskapView{
		Meta: meta("Statens eierskap · hva staten eier, og hvor mye",
			"Alle selskapene staten eier, med eierandel, kategori og organisasjonsnummer fra Brønnøysund.",
			"/eierskap"),
		SnapshotDate:  snapshotDate,
		Rows:          rows,
		Companies:     summary.Companies,
		WhollyOwned:   summary.WhollyOwned,
		Majority:      summary.Majority,
		Employees:     thousands(summary.Employees),
		WithEmployees: summary.WithEmployees,
		SpecialLaw:    summary.SpecialLaw,
		SourceName:    name,
		SourceURL:     url,
		RetrievedAt:   retrieved,
	}
}
