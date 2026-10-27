package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
)

// The state-ownership index is generated data: regjeringen.no's own list of what
// the state owns, enriched from Brønnøysund, built by
// tools/analytics/build_ownership_index.py.
//
// It exists because the dataset can say what a party wants done with state
// ownership and nothing about what the state actually owns. Without this, a
// statement about state ownership has a position and no object.
//
// It is deliberately not merged with the party-finance index: that one is money
// going to the parties, this is what the state holds on behalf of everyone.
//
//go:embed ownership/index.json
var ownershipIndexJSON []byte

// OwnedCompany is one company the state holds a share in.
type OwnedCompany struct {
	Company string `json:"company"`
	// Share is the state's ownership percentage, 0-100.
	Share float64 `json:"share"`
	// Category is the ownership category that governs the company: 1 commercial,
	// 2 commercial and sector policy, 3 sector policy, 4 other, - uncategorised.
	Category string `json:"category"`
	// SpecialLawCompany marks the "særlovselskap" the report flags: companies
	// whose governance is set by a special statute rather than the general
	// company acts.
	SpecialLawCompany bool   `json:"special_law_company"`
	Orgnr             string `json:"orgnr"`
	LegalForm         string `json:"legal_form"`
	// Employees is nil when Brønnøysund reports none, which is not zero: a holding
	// company can have no employees while a mismatched record would show another
	// company's staff.
	Employees    *int   `json:"employees"`
	IndustryCode string `json:"industry_code"`
	Industry     string `json:"industry"`
	Municipality string `json:"municipality"`
}

type ownershipSource struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	RetrievedAt string `json:"retrieved_at"`
	BrregAPI    string `json:"brreg_api"`
	Note        string `json:"note"`
}

type ownershipIndexFile struct {
	Source             ownershipSource   `json:"source"`
	Categories         map[string]string `json:"categories"`
	Companies          []OwnedCompany    `json:"companies"`
	UnmatchedCompanies []string          `json:"unmatched_companies"`
}

var ownershipIndexDoc = mustLoadOwnership()

var orgnrPattern = regexp.MustCompile(`^\d{9}$`)

func mustLoadOwnership() ownershipIndexFile {
	var idx ownershipIndexFile
	if err := json.Unmarshal(ownershipIndexJSON, &idx); err != nil {
		panic(fmt.Sprintf("internal/data/ownership/index.json is not valid: %v", err))
	}
	if len(idx.Companies) == 0 {
		panic("internal/data/ownership/index.json holds no companies; run tools/analytics/build_ownership_index.py")
	}
	if idx.Source.RetrievedAt == "" || idx.Source.URL == "" {
		panic("internal/data/ownership/index.json has no provenance")
	}
	return idx
}

// StateOwnedCompanies returns every company the state holds, largest share first
// (the generator sorts, and the order is part of the contract so the page renders
// deterministically).
func StateOwnedCompanies() []OwnedCompany { return ownershipIndexDoc.Companies }

// OwnershipCategories maps a category code to its Norwegian description.
func OwnershipCategories() map[string]string { return ownershipIndexDoc.Categories }

// OwnershipSourceInfo returns the provenance of the index.
func OwnershipSourceInfo() (name, url, retrievedAt, note string) {
	s := ownershipIndexDoc.Source
	return s.Name, s.URL, s.RetrievedAt, s.Note
}

// OwnershipSummary is the derived headline: how many companies, how many wholly
// owned, and how many employees sit in the companies Brønnøysund matched. The
// employee total is a floor, not a census: a company without a reported figure
// contributes nothing rather than being estimated.
type OwnershipSummary struct {
	Companies     int
	WhollyOwned   int
	Majority      int // share > 50
	Employees     int
	WithEmployees int
	Categorised   int
	SpecialLaw    int
	WithoutOrgnr  int
}

// SummariseOwnership derives the summary from the companies rather than storing
// it, so it cannot drift from the rows it describes.
func SummariseOwnership() OwnershipSummary {
	var s OwnershipSummary
	for _, c := range StateOwnedCompanies() {
		s.Companies++

		if c.Share == 100 {
			s.WhollyOwned++
		}
		if c.Share > 50 {
			s.Majority++
		}
		if c.Employees != nil {
			s.Employees += *c.Employees
			s.WithEmployees++
		}
		if c.Category != "-" && c.Category != "" {
			s.Categorised++
		}
		if c.SpecialLawCompany {
			s.SpecialLaw++
		}
		if !orgnrPattern.MatchString(c.Orgnr) {
			s.WithoutOrgnr++
		}
	}
	return s
}
