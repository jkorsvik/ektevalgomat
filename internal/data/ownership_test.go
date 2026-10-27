package data

import (
	"strings"
	"testing"
)

// The ownership index is a copy of the state's own list plus a register lookup, so
// the tests are about the copy being faithful and complete. A dropped row is an
// ownership fact that disappears, which is the failure that matters here.
func TestOwnershipIndexIsComplete(t *testing.T) {
	companies := StateOwnedCompanies()
	// Pinned, not floored: a floor of 50 would pass while a builder change quietly
	// dropped nineteen companies, and this test's own loop iterates the index, so
	// it cannot notice a row that is no longer there. The source page listed 69
	// companies on 2026-10-27; a different number means a row was added or lost and
	// a human must look before this number moves.
	const sourceCount = 69
	if len(companies) != sourceCount {
		t.Fatalf("index holds %d companies, the source page listed %d on 2026-10-27", len(companies), sourceCount)
	}
	for _, c := range companies {
		if strings.TrimSpace(c.Company) == "" {
			t.Errorf("a company row has no name: %+v", c)
		}
		if c.Share <= 0 || c.Share > 100 {
			t.Errorf("%s: share %v is outside (0, 100]", c.Company, c.Share)
		}
		if _, ok := OwnershipCategories()[c.Category]; !ok {
			t.Errorf("%s: unknown category %q", c.Company, c.Category)
		}
		if !orgnrPattern.MatchString(c.Orgnr) {
			t.Errorf("%s: organisation number %q is not nine digits", c.Company, c.Orgnr)
		}
	}
}

// Every company must carry a distinct organisation number, or two rows describe
// the same company and a reader counting holdings counts it twice.
func TestOwnershipOrgNumbersAreUnique(t *testing.T) {
	seen := map[string]string{}
	for _, c := range StateOwnedCompanies() {
		if prev, ok := seen[c.Orgnr]; ok {
			t.Errorf("%s and %s share organisation number %s", prev, c.Company, c.Orgnr)
		}
		seen[c.Orgnr] = c.Company
	}
}

// The enrichment must actually have happened, and the count is asserted so a
// regression that silently stops matching Brønnøysund fails rather than rendering
// an emptier page.
func TestOwnershipEnrichmentIsPresent(t *testing.T) {
	companies := StateOwnedCompanies()
	withForm, withIndustry := 0, 0
	for _, c := range companies {
		if c.LegalForm != "" {
			withForm++
		}
		if c.IndustryCode != "" {
			withIndustry++
		}
	}
	if withForm < len(companies) {
		t.Errorf("%d of %d companies have no legal form: enrichment did not run", len(companies)-withForm, len(companies))
	}
	if withIndustry < len(companies)*9/10 {
		t.Errorf("only %d of %d companies have an industry code", withIndustry, len(companies))
	}
}

// The summary is derived, so it must follow the rows. If a stored summary ever
// replaces it, these numbers would drift from the companies they describe.
func TestOwnershipSummaryFollowsTheRows(t *testing.T) {
	s := SummariseOwnership()
	companies := StateOwnedCompanies()

	if s.Companies != len(companies) {
		t.Errorf("summary counts %d companies, rows are %d", s.Companies, len(companies))
	}
	wholly := 0
	for _, c := range companies {
		if c.Share == 100 {
			wholly++
		}
	}
	if s.WhollyOwned != wholly {
		t.Errorf("summary says %d wholly owned, rows say %d", s.WhollyOwned, wholly)
	}
	if s.Majority < s.WhollyOwned {
		t.Errorf("majority holdings (%d) cannot be fewer than wholly owned (%d)", s.Majority, s.WhollyOwned)
	}
	if s.Employees == 0 {
		t.Error("no employees summed at all: either the enrichment or the summary is broken")
	}
	if s.WithoutOrgnr != 0 {
		t.Errorf("%d companies have no organisation number; the builder reports those as unmatched", s.WithoutOrgnr)
	}
	t.Logf("%d companies, %d wholly owned, %d majority, %d employees across %d companies, %d special-law",
		s.Companies, s.WhollyOwned, s.Majority, s.Employees, s.WithEmployees, s.SpecialLaw)
}

// The largest holding is a fact a reader will check, so it is pinned to the
// company the report names rather than to whatever sorts first.
func TestOwnershipIncludesTheKnownLargeHoldings(t *testing.T) {
	want := map[string]float64{
		"Equinor ASA":  67,
		"DNB Bank ASA": 34,
	}
	for name, share := range want {
		found := false
		for _, c := range StateOwnedCompanies() {
			if c.Company == name {
				found = true
				if c.Share != share {
					t.Errorf("%s: share %v, want %v", name, c.Share, share)
				}
			}
		}
		if !found {
			t.Errorf("%s is missing from the ownership index", name)
		}
	}
}

func TestOwnershipCarriesProvenance(t *testing.T) {
	name, url, retrieved, note := OwnershipSourceInfo()
	if name == "" || url == "" || retrieved == "" {
		t.Errorf("incomplete provenance: name=%q url=%q retrieved=%q", name, url, retrieved)
	}
	if !strings.Contains(note, "exact name match") {
		t.Errorf("provenance does not state how companies were matched: %q", note)
	}
}
