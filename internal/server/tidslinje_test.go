package server

import (
	"log/slog"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/analysis"
	"github.com/jkorsvik/ektevalgomat/internal/data"
)

func getTidslinje(t *testing.T) string {
	t.Helper()
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/tidslinje", nil))
	if w.Code != 200 {
		t.Fatalf("/tidslinje answered %d, want 200", w.Code)
	}
	return w.Body.String()
}

// The page must render, and it must derive its own totals rather than carry a
// number in prose. The counts asserted here come from the dataset, so a change to
// the content that the page does not follow fails.
func TestTidslinjeRendersAndDerivesItsNumbers(t *testing.T) {
	body := getTidslinje(t)
	ds := data.Store

	// Independent derivation, on the page's own definition: one observation is a
	// distinct (statement, party, year). Two rows from the same year are one
	// observation, and that is why this cannot just count evidence rows.
	type obs struct {
		stmt  string
		party string
		year  int
	}
	seen := map[obs]bool{}
	undatedSet := map[[2]string]bool{}
	for _, e := range ds.Evidence {
		if e.Position == nil || e.PartyID == "" {
			continue
		}
		if y, ok := yearFromDate(e.Date); ok {
			seen[obs{e.StatementID, e.PartyID, y}] = true
		} else {
			undatedSet[[2]string{e.StatementID, e.PartyID}] = true
		}
	}
	dated := len(seen)
	undated := len(undatedSet)
	coded := dated + undated
	if coded == 0 {
		t.Fatal("no coded position in the dataset: the page cannot be checked")
	}
	if !strings.Contains(body, strconv.Itoa(dated)+" av "+strconv.Itoa(coded)+" kodede posisjoner") {
		t.Errorf("page does not state %d of %d coded positions with a year", dated, coded)
	}
	t.Logf("rendered %d bytes; %d coded, %d dated, %d undated", len(body), coded, dated, undated)

	// Every dimension block must be present.
	for _, dim := range analysis.DimensionProfile(ds, "ap") {
		if !strings.Contains(body, "<h3>"+dim.Dimension+"</h3>") {
			t.Errorf("dimension %q is missing from the page", dim.Dimension)
		}
	}

	// The year columns must be the years the data actually holds.
	_, years := analysis.YearGrid(ds, analysis.PartyIDs(ds))
	for _, y := range years {
		if !strings.Contains(body, "<th>"+strconv.Itoa(y)+"</th>") {
			t.Errorf("year %d has observations but no column on the page", y)
		}
	}
}

// A cell that holds an observation must print its denominator, so no number on
// the page can be read as stronger than it is. The count of "(n)" groups must
// equal the number of filled cells, and every filled cell must be a mean on the
// -2..+2 scale.
func TestTidslinjeCellsCarryTheirDenominator(t *testing.T) {
	body := getTidslinje(t)
	parts := strings.Split(body, "<tbody>")
	if len(parts) < 2 {
		t.Fatal("page has no table body")
	}
	cellRe := regexp.MustCompile(`<td class="(?:pos|neg|zero)">(-?\d,\d) \((\d+)\)</td>`)
	filled := 0
	for _, part := range parts[1:] {
		for _, m := range cellRe.FindAllStringSubmatch(part, -1) {
			filled++
			mean, err := strconv.ParseFloat(strings.Replace(m[1], ",", ".", 1), 64)
			if err != nil {
				t.Fatalf("cell %q is not a number: %v", m[1], err)
			}
			if mean < -2 || mean > 2 {
				t.Errorf("cell mean %v is outside the -2..+2 scale", mean)
			}
			n, err := strconv.Atoi(m[2])
			if err != nil || n < 1 {
				t.Errorf("cell %q has denominator %q, want a positive count", m[0], m[2])
			}
		}
	}
	if filled == 0 {
		t.Fatal("no dated cell rendered: the assertion would pass on nothing")
	}
	t.Logf("%d filled year cells, each with its denominator", filled)
}

// An empty cell must render as a mark, never as 0,0. The difference between
// "nothing known" and "neutral" is the whole point of the page.
func TestTidslinjeNeverPrintsZeroForAnEmptyYear(t *testing.T) {
	body := getTidslinje(t)
	parts := strings.Split(body, "<tbody>")
	empties := 0
	for _, part := range parts[1:] {
		empties += strings.Count(part, `<td class="muted">·</td>`)
	}
	if empties == 0 {
		t.Fatal("no empty year cell on the page: either the data grew or the mark was dropped")
	}
	// Anchor on the denominator, not on a class attribute: a cell with no
	// observation has no tone, so anchoring the check on class="zero" let the
	// mutation that renders "0,0 (0)" for an empty year pass.
	if strings.Contains(body, "(0)") {
		t.Error("a cell prints a denominator of zero, so an empty year reads as a value")
	}
	t.Logf("%d empty year cells marked with a middle dot", empties)
}

// yearFromDate is the test's own partial year parse, kept separate from the
// production one so a bug in either shows up as a disagreement.
func yearFromDate(d string) (int, bool) {
	d = strings.TrimSpace(d)
	if len(d) < 4 {
		return 0, false
	}
	y, err := strconv.Atoi(d[:4])
	if err != nil || y < 1900 || y > 2100 {
		return 0, false
	}
	return y, true
}

// A cell that holds positions on both sides of zero must not be summarised as a
// mean. +2 and -2 average to 0, and 0 on this scale reads as "neutral / split",
// which is false about both observations. The page must show the spread, and the
// count of such cells is derived here rather than trusted.
func TestTidslinjeShowsTheSpreadWherePositionsDisagree(t *testing.T) {
	body := getTidslinje(t)
	ds := data.Store

	// Independent derivation: per (dimension, party, year), do the coded positions
	// straddle zero?
	dimOf := map[string]string{}
	for _, st := range ds.Statements {
		dimOf[st.ID] = st.Dimension
	}
	type k struct {
		dim, party string
		year       int
	}
	lo := map[k]int{}
	hi := map[k]int{}
	for _, e := range ds.Evidence {
		if e.Position == nil || e.PartyID == "" {
			continue
		}
		y, ok := yearFromDate(e.Date)
		if !ok {
			continue
		}
		kk := k{dimOf[e.StatementID], e.PartyID, y}
		if v, seen := lo[kk]; !seen || *e.Position < v {
			lo[kk] = *e.Position
		}
		if v, seen := hi[kk]; !seen || *e.Position > v {
			hi[kk] = *e.Position
		}
	}
	want := 0
	for kk, l := range lo {
		if l < 0 && hi[kk] > 0 {
			want++
		}
	}

	got := strings.Count(body, `<td class="split">`)
	if got != want {
		t.Errorf("%d cells render a spread, %d cells hold positions on both sides of zero", got, want)
	}
	if want == 0 {
		t.Log("no cell disagrees today: the rule is asserted on the derivation, not on the page")
	}
	t.Logf("%d split cells, %d disagreeing cells in the data", got, want)

	// And a split cell must read as a range, not a mean.
	for _, part := range strings.Split(body, `<td class="split">`)[1:] {
		cell := part[:strings.Index(part, "</td>")]
		if !strings.Contains(cell, " til ") {
			t.Errorf("split cell %q does not show a range", cell)
		}
		if strings.HasPrefix(cell, "0,0") {
			t.Errorf("split cell %q is summarised as a mean", cell)
		}
	}
}

// The statement page is where a reader meets one party's position on one
// question. FrP voted for a statutory max price in 2022 and against one in 2025,
// and on utsagn-12 that is a documented reversal. The page must show the two ends
// and name the change, not print "0 (nøytral/delt)" and "ikke nok data": those are
// the two strings the averaging bug produced, and they are false about both dates.
func TestStatementPageShowsADocumentedChangeOfPosition(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/utsagn/12", nil))
	if w.Code != 200 {
		t.Fatalf("/utsagn/12 answered %d, want 200", w.Code)
	}
	body := w.Body.String()

	if !strings.Contains(body, "Endret posisjon mellom 2022 og 2025") {
		t.Error("the page does not name FrP's change of position between 2022 and 2025")
	}
	// html/template escapes "+" as &#43;, so the spread is asserted on the parts
	// that survive escaping, and the split marker separately.
	for _, want := range []string{"po-word-delt", "1 (for) til −1 (mot)", "(2022 og 2025)"} {
		if !strings.Contains(body, want) {
			t.Errorf("the page does not show %q", want)
		}
	}
	// The stable parties must keep the plain label, so the range does not spread to
	// cells that do not need it.
	if !strings.Contains(body, "Samme posisjon i 2022 og 2025") {
		t.Error("the page does not report the parties whose position held across both years")
	}
}
