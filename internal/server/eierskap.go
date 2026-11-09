package server

import (
	"sort"
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
	Company       string
	Share         string
	ShareBarClass string
	Category      string
	CategoryWhy   string
	LegalForm     string
	Employees     string
	Industry      string
	Municipality  string
	Orgnr         string
	SpecialLaw    bool
	BrregURL      string
}

type eierskapCategoryStat struct {
	Code     string
	Label    string
	Count    int
	BarClass string
}

type eierskapEmployer struct {
	Company   string
	Share     string
	Employees string
	BarClass  string
}

type eierskapView struct {
	Meta            pageMeta
	SnapshotDate    string
	Rows            []eierskapRow
	Categories      []eierskapCategoryStat
	TopEmployers    []eierskapEmployer
	Companies       int
	WhollyOwned     int
	Majority        int
	PartialMajority int
	Minority        int
	Employees       string
	WithEmployees   int
	SpecialLaw      int
	SourceName      string
	SourceURL       string
	RetrievedAt     string
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

	catCounts := map[string]int{}
	rows := make([]eierskapRow, 0, len(companies))
	for _, c := range companies {
		catCounts[c.Category]++
		employees := ""
		if c.Employees != nil {
			employees = thousands(*c.Employees)
		}
		rows = append(rows, eierskapRow{
			Company:       c.Company,
			Share:         share(c.Share),
			ShareBarClass: pctStepClass(c.Share),
			Category:      c.Category,
			CategoryWhy:   categories[c.Category],
			LegalForm:     c.LegalForm,
			Employees:     employees,
			Industry:      c.Industry,
			Municipality:  c.Municipality,
			Orgnr:         c.Orgnr,
			SpecialLaw:    c.SpecialLawCompany,
			BrregURL:      "https://virksomhet.brreg.no/nb/oppslag/enheter/" + c.Orgnr,
		})
	}

	catOrder := []string{"1", "2", "3", "-"}
	var catStats []eierskapCategoryStat
	for _, code := range catOrder {
		cnt := catCounts[code]
		if cnt == 0 {
			continue
		}
		bar := "pct-0"
		if summary.Companies > 0 {
			bar = pctStepClass(100 * float64(cnt) / float64(summary.Companies))
		}
		catStats = append(catStats, eierskapCategoryStat{
			Code:     code,
			Label:    categories[code],
			Count:    cnt,
			BarClass: bar,
		})
	}

	// Top 10 employers by reported Brønnøysund employee count
	withEmp := make([]data.OwnedCompany, 0, len(companies))
	for _, c := range companies {
		if c.Employees != nil && *c.Employees > 0 {
			withEmp = append(withEmp, c)
		}
	}
	sort.Slice(withEmp, func(i, j int) bool {
		if *withEmp[i].Employees != *withEmp[j].Employees {
			return *withEmp[i].Employees > *withEmp[j].Employees
		}
		return withEmp[i].Company < withEmp[j].Company
	})
	topN := 8
	if len(withEmp) < topN {
		topN = len(withEmp)
	}
	maxEmp := 0
	if topN > 0 {
		maxEmp = *withEmp[0].Employees
	}
	var topEmployers []eierskapEmployer
	for i := 0; i < topN; i++ {
		c := withEmp[i]
		emp := *c.Employees
		bar := "pct-0"
		if maxEmp > 0 {
			bar = pctStepClass(100 * float64(emp) / float64(maxEmp))
		}
		topEmployers = append(topEmployers, eierskapEmployer{
			Company:   c.Company,
			Share:     share(c.Share),
			Employees: thousands(emp),
			BarClass:  bar,
		})
	}

	partialMajority := summary.Majority - summary.WhollyOwned
	if partialMajority < 0 {
		partialMajority = 0
	}
	minority := summary.Companies - summary.Majority
	if minority < 0 {
		minority = 0
	}

	name, url, retrieved, _ := data.OwnershipSourceInfo()
	return eierskapView{
		Meta: meta("Statens eierskap · hva staten eier, og hvor mye",
			"Alle selskapene staten eier, med eierandel, kategori og organisasjonsnummer fra Brønnøysund.",
			"/eierskap"),
		SnapshotDate:    snapshotDate,
		Rows:            rows,
		Categories:      catStats,
		TopEmployers:    topEmployers,
		Companies:       summary.Companies,
		WhollyOwned:     summary.WhollyOwned,
		Majority:        summary.Majority,
		PartialMajority: partialMajority,
		Minority:        minority,
		Employees:       thousands(summary.Employees),
		WithEmployees:   summary.WithEmployees,
		SpecialLaw:      summary.SpecialLaw,
		SourceName:      name,
		SourceURL:       url,
		RetrievedAt:     retrieved,
	}
}
