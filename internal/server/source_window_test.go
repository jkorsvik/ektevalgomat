package server

import (
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"log/slog"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The archived window is the only thing on the site that answers a source that
// has moved or been deleted. It can disappear in three ways without any Go
// test noticing: the template block gets dropped in a refactor, withWindow
// stops attaching the pointer, or the archive stops being embedded. So this
// asserts the rendered page, not the pieces.
func TestClaimPageRendersItsArchivedWindow(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	// Find a claim that carries a window rather than hard-coding an id that a
	// dataset edit could retire.
	var evidenceID string
	var statement *data.Statement
	for _, e := range data.Store.Evidence {
		if e.VerbatimExcerpt == "" {
			continue
		}
		if w, ok := data.SourceWindowFor(e.ID); ok && w.WindowChars > 0 {
			stmt, found := data.Store.Statement(e.StatementID)
			if !found {
				continue
			}
			evidenceID, statement = e.ID, &stmt
			break
		}
	}
	if statement == nil {
		t.Fatal("no evidence row has both an excerpt and a window: the archive is not embedded")
	}
	stmt := statement
	pagePath := "/utsagn/" + strconv.Itoa(stmt.Number)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", pagePath, nil))
	if w.Code != 200 {
		t.Fatalf("%s: status %d", pagePath, w.Code)
	}
	body := w.Body.String()

	if !strings.Contains(body, `class="source-window-text"`) {
		t.Fatalf("%s renders no archived window; expected the block for %s", pagePath, evidenceID)
	}
	if strings.Count(body, `class="source-window-text"`) < 1 {
		t.Fatalf("counted no window block")
	}
	// Provenance must reach the reader, not just the data: a window with no
	// fingerprint is an anonymous quotation.
	sha := regexp.MustCompile(`SHA-256 <code>[0-9a-f]{64}</code>`)
	if !sha.MatchString(body) {
		t.Errorf("%s shows a window without its document SHA-256", pagePath)
	}
	if !regexp.MustCompile(`Lest \d{4}-\d{2}-\d{2}`).MatchString(body) {
		t.Errorf("%s shows a window without the date it was read", pagePath)
	}
	// The caption must admit that the text is a raw extract, because it
	// visibly contains site navigation on several sources.
	if !strings.Contains(body, "rå og urenset") {
		t.Errorf("%s does not say the window is raw extracted text", pagePath)
	}
}

// A claim with no archived window must not render an empty block: a blank
// disclosure reads as a broken page rather than as an absent archive.
func TestClaimWithoutWindowRendersNoEmptyBlock(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	checked := 0
	for _, e := range data.Store.Evidence {
		if _, ok := data.SourceWindowFor(e.ID); ok {
			continue
		}
		stmt, ok := data.Store.Statement(e.StatementID)
		if !ok {
			continue
		}
		pagePath := "/utsagn/" + strconv.Itoa(stmt.Number)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", pagePath, nil))
		if strings.Contains(w.Body.String(), `class="source-window-text">\n      </pre>`) {
			t.Errorf("%s renders an empty window block for %s", pagePath, e.ID)
		}
		checked++
		if checked >= 5 {
			break
		}
	}
	if checked == 0 {
		t.Fatal("found no evidence row without a window: this test checked nothing")
	}
}

// The download page must state what is and is not stored, and it must state
// the number it actually holds rather than a hand-written one.
func TestDataPageStatesTheArchivePolicy(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/data", nil))
	body := w.Body.String()

	for _, want := range []string{
		"Kildeteksten rundt sitatet",
		"Det som ikke er med",
		"lagrer <em>ikke</em> hele dokumenter",
		"si fra i en sak",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/data is missing %q", want)
		}
	}
	count := regexp.MustCompile(`<strong>(\d+) tekstiutdrag</strong>`).FindStringSubmatch(body)
	if count == nil {
		t.Fatal("/data does not say how many extracts are stored")
	}
	if got, want := count[1], data.SourceWindowCount(); got != itoa(want) {
		t.Errorf("/data says %s extracts; the embedded archive holds %d", got, want)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
