package server

import (
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// TestOMQuotesTheDatasetOnAddendumCoverage ties the two figures /om prints
// about utsagn 44 and 45 to the dataset they describe.
//
// The sentence it guards said Rødt and Sp were still missing a position on
// utsagn 45 for twelve days after both were coded, because the figure was
// written by hand. A gate on the sentence is the only thing that makes the
// derived figure worth deriving: without it, the next person to tidy the
// prose can hard-code the number again and nothing will notice.
func TestOMQuotesTheDatasetOnAddendumCoverage(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/om", nil))
	if w.Code != 200 {
		t.Fatalf("/om returned %d", w.Code)
	}
	body := strings.Join(strings.Fields(w.Body.String()), " ")

	checked := 0
	for _, n := range []int{44, 45} {
		cov := statementCoverage(data.Store, n)
		if cov.Total == 0 {
			t.Fatalf("statement %d: no parties loaded", n)
		}
		if cov.Coded == 0 {
			t.Fatalf("statement %d: no coded party positions; the check would be vacuous", n)
		}
		want := fmt.Sprintf("%d av %d partier", cov.Coded, cov.Total)
		if !strings.Contains(body, want) {
			t.Errorf("utsagn %d: /om does not say %q, so it disagrees with the dataset", n, want)
		}
		if n == 45 && !strings.Contains(body, "kildeklasse "+cov.Tiers) {
			t.Errorf("utsagn 45: /om does not name the source tiers behind the positions (%s)", cov.Tiers)
		}
		checked++
	}

	// The half of the same sentence that a derived number cannot protect:
	// "kommer fra partiprogrammene" is only true while every coded position
	// on utsagn 45 is programme-only. Hold the wording to that flag, so
	// coding one vote on the statement flips the sentence rather than
	// leaving a false claim behind.
	c45 := statementCoverage(data.Store, 45)
	claim, _, _ := strings.Cut(body, "Rettelseslogg")
	saysProgramme := strings.Contains(claim, "kommer fra partiprogrammene")
	if c45.ProgrammeOnly && !saysProgramme {
		t.Error("every coded position on utsagn 45 is programme-only, and /om does not say so")
	}
	if !c45.ProgrammeOnly && saysProgramme {
		t.Errorf("utsagn 45 has an A or B source among its positions (tiers %s), but /om still says they come from the programmes alone", c45.Tiers)
	}
	// The exact sentence that was wrong must not come back to the text a
	// reader acts on. The correction log below it quotes the sentence, and
	// must be allowed to: that is what a correction log is for. So the check
	// stops where the log begins.
	for _, gone := range []string{"Rødt og Sp mangler", "sju av ni"} {
		if strings.Contains(claim, gone) {
			t.Errorf("/om still makes the hand-written claim %q", gone)
		}
	}
	if checked == 0 {
		t.Fatal("checked no statements")
	}
}
