package server

import (
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
	"github.com/jkorsvik/ektevalgomat/internal/scoring"
)

// TestNoEvidenceAfterSnapshotDate makes the badge text true by
// construction. Every page says "Øyeblikksbilde: hendelser til og med
// SnapshotDate". Sources may be retrieved later (most were), but no
// event itself may be dated after the cutoff. Dates are compared as
// strings, which also orders the free-form ones ("2025 (partiprogram)",
// "2026 (sesjon)") before "2026-09-23".
func TestNoEvidenceAfterSnapshotDate(t *testing.T) {
	checked := 0
	for _, e := range data.Store.Evidence {
		checked++
		if e.Date > SnapshotDate {
			t.Errorf("%s: event date %q is after SnapshotDate %s", e.ID, e.Date, SnapshotDate)
		}
	}
	if checked == 0 {
		t.Fatal("no evidence loaded, the check ran on nothing")
	}
	t.Logf("checked %d evidence dates against %s", checked, SnapshotDate)
}

// TestStatementHeadlineCountsCodedPositions checks every /utsagn/N page:
// the headline must equal the number of parties with a coded position,
// counted here independently through scoring.ComputeCell. Parties with a
// source but no coded stance (utsagn-45 on 2026-09-27: Innst. 464 S
// context for R and Sp, position null) must not inflate it. Generic on
// purpose: coding more parties later must not break this test.
func TestStatementHeadlineCountsCodedPositions(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	ds := data.Store
	pages, sourcedOnly := 0, 0
	for _, s := range ds.Statements {
		coded, sourced := 0, 0
		for _, p := range ds.Parties {
			c := scoring.ComputeCell(ds, s.ID, p.ID)
			if c.PartyPosition != nil {
				coded++
			} else if c.HasEvidence() {
				sourced++
			}
		}
		if coded+sourced == 0 {
			continue // compact "0 av N" card, different template branch
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", fmt.Sprintf("/utsagn/%d", s.Number), nil))
		body := w.Body.String()
		want := fmt.Sprintf("%d av %d partier har en posisjon med kilde.", coded, len(ds.Parties))
		if !strings.Contains(body, want) {
			t.Errorf("%s: headline missing %q", s.ID, want)
		}
		if sourced > 0 {
			sourcedOnly++
			if !strings.Contains(body, fmt.Sprintf("%d til har kilder", sourced)) {
				t.Errorf("%s: %d sourced-only parties not stated", s.ID, sourced)
			}
		}
		pages++
	}
	if pages == 0 {
		t.Fatal("no statement pages checked")
	}
	t.Logf("checked %d pages, %d with sourced-only parties", pages, sourcedOnly)
}

// TestMetodeStatementFormNumbers ties the /metode text to its sources:
// the flip count to FLIPPED_2026_09_27 in quiz.js (the list that migrates
// returning users' answers), and the spread to the same function the
// ratchet test uses.
func TestMetodeStatementFormNumbers(t *testing.T) {
	js, err := os.ReadFile("../../web/static/js/quiz.js")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?s)FLIPPED_2026_09_27 = \[(.*?)\]`).FindSubmatch(js)
	if m == nil {
		t.Fatal("FLIPPED_2026_09_27 not found in quiz.js")
	}
	flipped := len(regexp.MustCompile(`"utsagn-\d+"`).FindAll(m[1], -1))
	if flipped == 0 {
		t.Fatal("FLIPPED_2026_09_27 parsed as empty")
	}

	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/metode", nil))
	body := strings.Join(strings.Fields(w.Body.String()), " ")
	for _, want := range []string{
		fmt.Sprintf("%d av utsagnene er derfor formulert motsatt vei", flipped),
		fmt.Sprintf("skiller det nå %d prosentpoeng", scoring.MaxAcquiescenceSpread(data.Store)),
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/metode missing %q", want)
		}
	}
}
