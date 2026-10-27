package server

import (
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/analysis"
	"github.com/jkorsvik/ektevalgomat/internal/data"
)

func getPenger(t *testing.T) string {
	t.Helper()
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/penger", nil))
	if w.Code != 200 {
		t.Fatalf("/penger answered %d, want 200", w.Code)
	}
	return w.Body.String()
}

// The page claims its figures come from the index, so the figures are recomputed
// here from the index. A page that renders but prints a number the data does not
// support fails, which is the failure that matters for a money page.
func TestPengerPageRendersAndDerivesItsNumbers(t *testing.T) {
	body := getPenger(t)
	if !strings.Contains(body, "Penger") {
		t.Fatal("/penger rendered without its heading")
	}

	years := data.FinanceYears()
	latest := years[len(years)-1]
	ds := data.Store

	checked := 0
	for _, id := range analysis.PartyIDs(ds) {
		f, ok := data.FinanceFor(id, latest)
		if !ok || f.StateSupport == nil || f.Income == nil {
			continue
		}
		wantState := kronerMill(f.StateSupport)
		wantShare := percent(f.StateShare())
		if !strings.Contains(body, wantState) {
			t.Errorf("party %s: page lacks state support %s (mill)", id, wantState)
		}
		if wantShare != "" && !strings.Contains(body, wantShare+" %") {
			t.Errorf("party %s: page lacks the derived share %s %%", id, wantShare)
		}
		checked++
	}
	if checked < 8 {
		t.Fatalf("only %d parties checked: the assertion is not covering the table", checked)
	}
	t.Logf("%d bytes, %d parties checked for %s", len(body), checked, latest)
}

// The claim "absence is not zero" lives in the formatter, not in the template's
// absent branch. An earlier version of this test checked the page for the absent
// marker, which the matrix renders on a path the formatter never touches, so it
// stayed green when the formatter was mutated to print 0,0. The assertion now sits
// on the function that decides.
func TestKronerMillRendersAbsenceAsEmpty(t *testing.T) {
	if got := kronerMill(nil); got != "" {
		t.Errorf("kronerMill(nil) = %q, want empty: absence must not render as a number", got)
	}
	if got := percent(-1); got != "" {
		t.Errorf("percent(-1) = %q, want empty: a share without a denominator is not 0 %%", got)
	}
	v := int64(133118678)
	if got := kronerMill(&v); got != "133,1" {
		t.Errorf("kronerMill(133118678) = %q, want 133,1 (millions, Norwegian comma)", got)
	}
	zero := int64(0)
	if got := kronerMill(&zero); got != "0,0" {
		t.Errorf("kronerMill(0) = %q, want 0,0: a real zero must still render", got)
	}
}

// And at page level: a party-year the index does not hold must use the absent
// marker. This checks the template, and is kept separate from the formatter test
// above because the two paths are different code.
func TestPengerDoesNotRenderAbsenceAsZero(t *testing.T) {
	body := getPenger(t)
	years := data.FinanceYears()
	missing := 0
	for _, id := range analysis.PartyIDs(data.Store) {
		for _, y := range years {
			if _, ok := data.FinanceFor(id, y); ok {
				continue
			}
			missing++
		}
	}
	if missing == 0 {
		t.Skip("every scored party-year is present, so there is no absence to render")
	}
	if !strings.Contains(body, `<td class="muted">–</td>`) {
		t.Error("page holds an absent party-year but never renders the absent marker")
	}
}

// The caveat is the page's most load-bearing sentence: SSB has no anonymity field,
// so a reader must not read the residual as one. If it disappears in a refactor,
// the page silently claims something the source does not say.
func TestPengerStatesWhatTheSourceCannotAnswer(t *testing.T) {
	body := getPenger(t)
	for _, want := range []string{
		"ingen opplysning om anonyme bidrag",
		"ikke kategoriserte givere",
		"restpost",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("page does not state %q: the residual would read as an anonymity count", want)
		}
	}
}

// Provenance is on the page, so a reader can check the number without the repo.
func TestPengerCarriesItsSource(t *testing.T) {
	body := getPenger(t)
	table, api, human, retrieved, _ := data.FinanceSource()
	if !strings.Contains(body, table) {
		t.Errorf("page does not name table %s", table)
	}
	if !strings.Contains(body, human) {
		t.Errorf("page does not link the human source %s", human)
	}
	if !strings.Contains(body, retrieved) {
		t.Errorf("page does not state when the index was read (%s)", retrieved)
	}
	if api == "" {
		t.Error("index has no API provenance")
	}
}

// The matrix must cover every year the index holds, or a reader sees a shorter
// history than exists without being told.
func TestPengerMatrixCoversEveryYear(t *testing.T) {
	body := getPenger(t)
	years := data.FinanceYears()
	for _, y := range years {
		if !strings.Contains(body, fmt.Sprintf("<th>%s</th>", y)) {
			t.Errorf("year %s is missing from the matrix", y)
		}
	}
}
