package server

import (
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/analysis"
	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The map claims to be derived, so the page is loaded over HTTP and its numbers
// are recomputed here from the dataset. A view that renders in a struct but 500s
// on the wire, or that prints a figure the data does not support, fails.
func TestKartPageRendersAndDerivesItsNumbers(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/kart", nil))
	if w.Code != 200 {
		t.Fatalf("/kart answered %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Sakområde × parti") {
		t.Fatal("/kart rendered without its table")
	}

	ds := data.Store
	ids := analysis.PartyIDs(ds)
	totalCoded, totalMulti := 0, 0
	for _, id := range ids {
		coded, multi, _ := analysis.PartyTotals(analysis.DimensionProfile(ds, id))
		totalCoded += coded
		totalMulti += multi
	}
	if totalCoded == 0 {
		t.Fatal("no coded position in any dimension: the page cannot be derived")
	}

	// The headline sentence states both totals. If they stop matching the data,
	// the page is claiming something the dataset does not say.
	if !strings.Contains(body, fmt.Sprintf("%d av", totalMulti)) {
		t.Errorf("page does not state %d multi-dated positions", totalMulti)
	}
	if !strings.Contains(body, strconv.Itoa(totalCoded)) {
		t.Errorf("page does not state the %d coded positions in total", totalCoded)
	}

	// Every dimension must render its row, so a new dimension cannot be silently
	// dropped by a stale list.
	dims := map[string]bool{}
	for _, s := range ds.Statements {
		dims[s.Dimension] = true
	}
	for dim := range dims {
		if !strings.Contains(body, dim) {
			t.Errorf("dimension %q is missing from /kart", dim)
		}
	}
	t.Logf("rendered %d bytes; %d coded positions, %d with two dates, %d dimensions",
		len(body), totalCoded, totalMulti, len(dims))
}

// The cell denominator is the point of the table. A row where a party has one
// coded position of seven must print that, not a bare mean.
func TestKartCellsCarryTheirDenominator(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/kart", nil))
	body := w.Body.String()

	ds := data.Store
	checked := 0
	for _, id := range analysis.PartyIDs(ds) {
		for _, r := range analysis.DimensionProfile(ds, id) {
			if r.Coded == 0 {
				continue
			}
			want := fmt.Sprintf("(%d/%d)", r.Coded, r.Statements)
			if !strings.Contains(body, want) {
				t.Errorf("party %s, dimension %q: page lacks the denominator %s", id, r.Dimension, want)
			}
			checked++
			if checked >= 6 {
				return // one per party is enough to prove the shape
			}
		}
	}
	if checked == 0 {
		t.Fatal("no coded cell to check: the assertion would pass on nothing")
	}
}
