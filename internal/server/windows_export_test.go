package server

import (
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// getJSON fetches a path from the real handler and decodes the envelope.
func getJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	if w.Code != 200 {
		t.Fatalf("%s: status %d", path, w.Code)
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("%s: not JSON: %v", path, err)
	}
	return out
}

// The site now claims, on /data, that it stores the text around each quote.
// A claim like that is only worth anything if a reader can take the file and
// check it, so the export must carry every window the pages render — not a
// summary, and not a subset that happens to be cached.
func TestWindowsExportCarriesEveryArchivedWindow(t *testing.T) {
	out := getJSON(t, "/data/kildetekst.json")

	items, _ := out["items"].([]any)
	count, _ := out["count"].(float64)
	if int(count) != data.SourceWindowCount() {
		t.Errorf("export says %v windows; the archive holds %d", count, data.SourceWindowCount())
	}
	if len(items) != data.SourceWindowCount() {
		t.Fatalf("export has %d items; the archive holds %d", len(items), data.SourceWindowCount())
	}

	seen := map[string]bool{}
	for _, raw := range items {
		it, _ := raw.(map[string]any)
		id, _ := it["evidence_id"].(string)
		if id == "" {
			t.Fatal("an exported window has no evidence_id: it cannot be joined to the claim")
		}
		seen[id] = true

		// The whole point is the text. A metadata-only export would look
		// complete while containing nothing a reader can read.
		text, _ := it["text"].(string)
		if strings.TrimSpace(text) == "" {
			t.Errorf("%s: exported with empty text", id)
		}
		sha, _ := it["document_sha256"].(string)
		if len(sha) != 64 {
			t.Errorf("%s: document_sha256 is %q, not a sha256", id, sha)
		}
		if u, _ := it["source_url"].(string); u == "" {
			t.Errorf("%s: no source_url, so the quotation cannot be traced home", id)
		}
		if d, _ := it["fetched_at"].(string); len(d) != 10 {
			t.Errorf("%s: fetched_at is %q; a date is what makes it checkable", id, d)
		}
		// A stored char count that disagrees with the stored text means one of
		// them was edited, and the reader cannot tell which.
		if wc, _ := it["window_chars"].(float64); int(wc) != len([]rune(strings.TrimSuffix(text, "\n"))) {
			t.Errorf("%s: says %v chars, text holds %d", id, wc, len([]rune(strings.TrimSuffix(text, "\n"))))
		}
	}

	for _, sw := range data.SourceWindows() {
		if !seen[sw.EvidenceID] {
			t.Errorf("archived window %s is missing from the export", sw.EvidenceID)
		}
	}
}

// Third-party quotations must not be handed out under the project's own
// licence. Reusing it would tell a reader they may republish words the
// project does not own.
func TestWindowsExportDoesNotClaimTheProjectLicence(t *testing.T) {
	out := getJSON(t, "/data/kildetekst.json")
	licence, _ := out["licence"].(string)
	if licence == licenceNote {
		t.Fatalf("the source excerpts carry the dataset licence %q", licenceNote)
	}
	if !strings.Contains(licence, "kilde") {
		t.Errorf("licence %q does not point at the sources that own the text", licence)
	}
	if ds, _ := out["dataset"].(string); ds != "kildetekst" {
		t.Errorf("dataset is %q; consumers key on this name", ds)
	}
}

// A map iteration order would reshuffle the file on every request, so a
// reader diffing two snapshots would see changes that are not changes.
func TestWindowsExportIsStable(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	fetch := func() string {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/data/kildetekst.json", nil))
		return w.Body.String()
	}
	first, second := fetch(), fetch()
	if first != second {
		t.Fatal("two requests for the same file returned different bytes")
	}
	// And it must be ascending, which is the order a diff assumes.
	ids := []string{}
	var decoded struct {
		Items []struct {
			EvidenceID string `json:"evidence_id"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(first), &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, it := range decoded.Items {
		ids = append(ids, it.EvidenceID)
	}
	for i := 1; i < len(ids); i++ {
		if ids[i-1] > ids[i] {
			t.Fatalf("export is not sorted at %d: %q before %q", i, ids[i-1], ids[i])
		}
	}
}

// A file nobody can find is not published.
func TestDataPageOffersTheWindowExport(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/data", nil))
	body := w.Body.String()
	if !strings.Contains(body, `href="/data/kildetekst.json"`) {
		t.Error("/data does not link the window export")
	}
	if !strings.Contains(body, "AGPL-3.0") {
		t.Error("/data no longer states the distinction between the two licences")
	}
}
