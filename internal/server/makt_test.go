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

func getMakt(t *testing.T) string {
	t.Helper()
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/makt", nil))
	if w.Code != 200 {
		t.Fatalf("/makt answered %d, want 200", w.Code)
	}
	return html.UnescapeString(w.Body.String())
}

// The page claims to render three indexes, so each is recomputed here. The seats
// column is the load-bearing one: it is the denominator for every "N of 9 parties"
// claim on the site.
func TestMaktRendersSeatsAndProposalsPerParty(t *testing.T) {
	body := squash(getMakt(t))
	seats := data.SeatCounts()
	proposals := data.ProposalsPerParty()

	for _, id := range data.SeatRanking() {
		short, full := "", ""
		for _, p := range data.Store.Parties {
			if p.ID == id {
				short, full = p.Short, p.Name
			}
		}
		if short == "" || full == "" {
			t.Fatalf("party %s has no name", id)
		}
		// Anchor on the full name in the row's title attribute. A short name like
		// "H" or "V" occurs all over the page, so searching for it found the wrong
		// place and this test failed on correct output.
		marker := `title="` + full + `"`
		row := strings.Index(body, marker)
		if row < 0 {
			t.Errorf("party %s (%s) is missing from /makt", id, full)
			continue
		}
		end := row + 160
		if end > len(body) {
			end = len(body)
		}
		window := body[row:end]
		if !strings.Contains(window, short) {
			t.Errorf("party %s: short name %q not in its row", id, short)
		}
		// The row must carry both numbers. Checking them separately would pass on a
		// page that printed the right seats beside the wrong proposals.
		if !strings.Contains(window, itoa(seats[id])) {
			t.Errorf("party %s: seats %d not in its row", id, seats[id])
		}
		if !strings.Contains(window, itoa(proposals[id])) {
			t.Errorf("party %s: proposals %d not in its row", id, proposals[id])
		}
	}
	if !strings.Contains(body, itoa(data.SeatTotal())) {
		t.Errorf("page does not state the %d seats in total", data.SeatTotal())
	}
	if !strings.Contains(body, itoa(len(data.Proposals()))) {
		t.Errorf("page does not state the %d proposals", len(data.Proposals()))
	}
}

// The share is derived from the seats, so it must follow them.
func TestMaktSeatSharesAreDerived(t *testing.T) {
	body := squash(getMakt(t))
	total := data.SeatTotal()
	if total != 169 {
		t.Fatalf("seat total is %d, the Storting has 169", total)
	}
	for id, n := range data.SeatCounts() {
		want := percent(100 * float64(n) / float64(total))
		if !strings.Contains(body, want+" %") {
			t.Errorf("party %s: seat share %s %% is missing", id, want)
		}
	}
}

// The audits are the control on the rest. Every report the index holds must be on
// the page, with its date and a working link.
func TestMaktRendersEveryAudit(t *testing.T) {
	body := squash(getMakt(t))
	for _, r := range data.AuditReports() {
		if !strings.Contains(body, r.Title) {
			t.Errorf("audit %q is missing", r.Title)
		}
		if !strings.Contains(body, r.Published) {
			t.Errorf("audit %q: date %s is missing", r.Title, r.Published)
		}
		if !strings.Contains(body, r.URL) {
			t.Errorf("audit %q: link is missing", r.Title)
		}
	}
}

// The edges of the map are stated on the page, not only in the data. A reader who
// does not know about the refusals would assume the gaps are omissions.
func TestMaktStatesTheEdgesOfTheMap(t *testing.T) {
	body := squash(getMakt(t))
	committee, ok := data.CommitteeMembershipRefusal()
	if !ok {
		t.Fatal("no committee refusal in the data")
	}
	archive, ok := data.AuditArchiveRefusal()
	if !ok {
		t.Fatal("no archive refusal in the data")
	}
	for _, want := range []string{"Komitéplasser", "Revisjonsarkivet", committee.Reason, archive.Reason} {
		if !strings.Contains(body, want) {
			t.Errorf("page does not state %q", truncate(want, 60))
		}
	}
	if !strings.Contains(body, "kildens grense") {
		t.Error("page does not frame the missing parts as a source limit")
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
