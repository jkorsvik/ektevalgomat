package server

import (
	"html"
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

func getEierskap(t *testing.T) string {
	t.Helper()
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/eierskap", nil))
	if w.Code != 200 {
		t.Fatalf("/eierskap answered %d, want 200", w.Code)
	}
	// Company names contain "&", which the template escapes. Text assertions are
	// about the text, so the body is unescaped once here rather than every caller
	// remembering to.
	return html.UnescapeString(w.Body.String())
}

// The page claims to render the index, so every company and every share is
// recomputed here from the index. A company silently missing from the page is an
// ownership fact that disappears from the public record.
func TestEierskapRendersEveryCompany(t *testing.T) {
	body := getEierskap(t)
	companies := data.StateOwnedCompanies()
	if len(companies) < 50 {
		t.Fatalf("only %d companies in the index", len(companies))
	}
	for _, c := range companies {
		if !strings.Contains(body, c.Company) {
			t.Errorf("company %q is missing from /eierskap", c.Company)
		}
		if !strings.Contains(body, share(c.Share)+" %") {
			t.Errorf("%s: share %s %% is missing", c.Company, share(c.Share))
		}
		if !strings.Contains(body, c.Orgnr) {
			t.Errorf("%s: organisation number %s is missing, so the row cannot be checked", c.Company, c.Orgnr)
		}
	}
	// Each row must carry its register link. Counting them catches a row that
	// renders without one, which the per-company loop above cannot see.
	if got := strings.Count(body, "virksomhet.brreg.no"); got != len(companies) {
		t.Errorf("%d register links rendered for %d companies", got, len(companies))
	}
	t.Logf("%d companies rendered, %d bytes", len(companies), len(body))
}

// The summary is derived; the page must print what the rows say.
func TestEierskapSummaryMatchesTheRows(t *testing.T) {
	body := squash(getEierskap(t))
	s := data.SummariseOwnership()
	for _, want := range []string{
		thousands(s.Companies) + " selskaper",
		thousands(s.WhollyOwned) + " er heleid",
		thousands(s.Majority) + " er flertallseid",
		thousands(s.Employees) + " personer",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("page does not state %q", want)
		}
	}
}

// Absence is not zero here either: a company with no reported employee figure must
// show the absent marker rather than 0.
func TestEierskapDoesNotRenderMissingEmployeesAsZero(t *testing.T) {
	missing := 0
	for _, c := range data.StateOwnedCompanies() {
		if c.Employees == nil {
			missing++
		}
	}
	if missing == 0 {
		t.Skip("every company has an employee figure, so there is no absence to render")
	}
	body := getEierskap(t)
	if !strings.Contains(body, `<td><span class="muted">–</span></td>`) {
		t.Error("page holds a company without employees but never renders the absent marker")
	}
	if thousands(0) == "" {
		t.Error("thousands(0) is empty: a real zero would become an absence")
	}
}

// The formatters are the only place a number can be misread, so they are asserted
// directly rather than only through the page.
func TestNumberFormatters(t *testing.T) {
	if got := thousands(70358); got != "70 358" {
		t.Errorf("thousands(70358) = %q, want 70 358", got)
	}
	if got := thousands(999); got != "999" {
		t.Errorf("thousands(999) = %q, want 999", got)
	}
	if got := thousands(1000); got != "1 000" {
		t.Errorf("thousands(1000) = %q, want 1 000", got)
	}
	if got := share(67); got != "67" {
		t.Errorf("share(67) = %q, want 67", got)
	}
	if got := share(66.67); got != "66,67" {
		t.Errorf("share(66.67) = %q, want 66,67", got)
	}
	if got := share(12.08); got != "12,08" {
		t.Errorf("share(12.08) = %q, want 12,08", got)
	}
	if got := share(28.2); got != "28,2" {
		t.Errorf("share(28.2) = %q, want 28,2", got)
	}
}

// Provenance on the page, so a reader can check a holding without the repo.
func TestEierskapCarriesItsSource(t *testing.T) {
	body := squash(getEierskap(t))
	name, url, retrieved, _ := data.OwnershipSourceInfo()
	for _, want := range []string{name, url, retrieved} {
		if !strings.Contains(body, want) {
			t.Errorf("page does not state %q", want)
		}
	}
	if !strings.Contains(body, "eksakte navnetreff") {
		t.Error("page does not state how companies were matched to the register")
	}
}
