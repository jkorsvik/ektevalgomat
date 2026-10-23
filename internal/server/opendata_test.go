package server

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// get fetches a path through the real handler and returns the recorder, so a
// test can assert on status, headers and body together.
func get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	return w
}

// jsonKeysInOrder returns the top-level keys of a marshalled value in the
// order encoding/json emitted them, which is the struct's field order.
func jsonKeysInOrder(t *testing.T, v any) []string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	if _, err := dec.Token(); err != nil { // opening '{'
		t.Fatal(err)
	}
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatal(err)
		}
		keys = append(keys, tok.(string))
		var skip any
		if err := dec.Decode(&skip); err != nil {
			t.Fatal(err)
		}
	}
	return keys
}

// TestEvidenceColumnsAreTheJSONKeys is the guard that ties the three
// descriptions of an evidence item together: the JSON field names, the CSV
// header, and the table on /data all come from evidenceColumns. A field added
// to the JSON export without a matching column (or the reverse) fails here,
// instead of shipping a file that documents one shape and carries another.
func TestEvidenceColumnsAreTheJSONKeys(t *testing.T) {
	keys := jsonKeysInOrder(t, evidenceExportOf(data.Evidence{}))

	if len(evidenceColumns) == 0 || len(keys) == 0 {
		t.Fatal("empty column list: this guard would pass vacuously")
	}
	if len(keys) != len(evidenceColumns) {
		t.Fatalf("JSON has %d fields, evidenceColumns has %d: %v vs %v", len(keys), len(evidenceColumns), keys, columnNames())
	}
	for i, c := range evidenceColumns {
		if c.Name != keys[i] {
			t.Errorf("column %d is %q, JSON field %d is %q", i, c.Name, i, keys[i])
		}
		if c.Doc == "" || c.Value == nil {
			t.Errorf("column %q has no description or no value function", c.Name)
		}
	}
}

func columnNames() []string {
	names := make([]string, len(evidenceColumns))
	for i, c := range evidenceColumns {
		names[i] = c.Name
	}
	return names
}

// TestEvidenceCSVMatchesTheDataset parses the file a visitor downloads and
// holds it to the dataset: same header, same rows, and the one distinction
// that matters most in a spreadsheet — "not coded" must not arrive as 0.
func TestEvidenceCSVMatchesTheDataset(t *testing.T) {
	w := get(t, "/data/evidens.csv")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Errorf("Content-Type = %q", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); !strings.Contains(cd, "ektevalgomat-evidens.csv") {
		t.Errorf("Content-Disposition = %q", cd)
	}

	body := w.Body.Bytes()
	// A BOM would make the first header "\ufeffevidence_id" for any reader
	// using plain UTF-8, and the failure shows up as a missing column, not
	// as an error. /data tells an Excel user how to open the file instead.
	if bytes.HasPrefix(body, []byte("\xef\xbb\xbf")) {
		t.Fatal("CSV starts with a UTF-8 BOM")
	}

	records, err := csv.NewReader(bytes.NewReader(body)).ReadAll()
	if err != nil {
		t.Fatalf("the CSV does not parse: %v", err)
	}
	if len(records) < 2 {
		t.Fatalf("CSV has %d lines: nothing to check", len(records))
	}
	if got, want := strings.Join(records[0], ","), strings.Join(columnNames(), ","); got != want {
		t.Fatalf("header\n got: %s\nwant: %s", got, want)
	}
	rows := records[1:]
	if len(rows) != len(data.Store.Evidence) {
		t.Fatalf("CSV has %d rows, dataset has %d evidence items", len(rows), len(data.Store.Evidence))
	}

	posCol := columnIndex(t, "position")
	var seenEmpty, seenZero int
	for i, e := range data.Store.Evidence {
		cell := rows[i][posCol]
		switch {
		case e.Position == nil:
			seenEmpty++
			if cell != "" {
				t.Fatalf("%s has no position but the CSV cell is %q", e.ID, cell)
			}
		case *e.Position == 0:
			seenZero++
			if cell != "0" {
				t.Fatalf("%s has position 0 (nøytral) but the CSV cell is %q: an empty cell and a coded zero must not look the same", e.ID, cell)
			}
		}
	}
	// Both states have to exist in the data for the check above to mean
	// anything. If one disappears, the distinction stops being tested.
	if seenEmpty == 0 || seenZero == 0 {
		t.Fatalf("fixture no longer proves the empty-vs-zero distinction: %d uncoded, %d coded zero", seenEmpty, seenZero)
	}
}

func columnIndex(t *testing.T, name string) int {
	t.Helper()
	for i, c := range evidenceColumns {
		if c.Name == name {
			return i
		}
	}
	t.Fatalf("no column named %q", name)
	return -1
}

// TestExportsCoverTheSameRows holds the two formats to each other. A filter
// that drops rows from one of them would otherwise be invisible: both files
// still parse, and only a count would be wrong.
func TestExportsCoverTheSameRows(t *testing.T) {
	var js envelope[evidenceExport]
	if err := json.Unmarshal(get(t, "/data/evidens.json").Body.Bytes(), &js); err != nil {
		t.Fatalf("evidens.json does not parse: %v", err)
	}
	records, err := csv.NewReader(bytes.NewReader(get(t, "/data/evidens.csv").Body.Bytes())).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if js.Count != len(js.Items) {
		t.Errorf("count is %d but the file carries %d items", js.Count, len(js.Items))
	}
	if len(js.Items) != len(records)-1 {
		t.Fatalf("JSON has %d items, CSV has %d rows", len(js.Items), len(records)-1)
	}
	idCol := columnIndex(t, "evidence_id")
	for i, item := range js.Items {
		if item.EvidenceID != records[i+1][idCol] {
			t.Fatalf("row %d: JSON id %q, CSV id %q", i, item.EvidenceID, records[i+1][idCol])
		}
	}
}

// TestStatementsJSONIsTheDataset checks the statements export against the
// dataset by id, not by count alone: a file that returned the right number
// of the wrong statements would pass a count check.
func TestStatementsJSONIsTheDataset(t *testing.T) {
	var env envelope[statementExport]
	if err := json.Unmarshal(get(t, "/data/utsagn.json").Body.Bytes(), &env); err != nil {
		t.Fatalf("utsagn.json does not parse: %v", err)
	}
	if env.SchemaVersion != openDataSchemaVersion {
		t.Errorf("schema_version = %d", env.SchemaVersion)
	}
	if env.SnapshotDate != SnapshotDate {
		t.Errorf("snapshot_date = %q, want %q", env.SnapshotDate, SnapshotDate)
	}
	if env.Count != len(data.Store.Statements) || len(env.Items) != env.Count {
		t.Fatalf("count=%d items=%d dataset=%d", env.Count, len(env.Items), len(data.Store.Statements))
	}

	want := map[string]bool{}
	for _, s := range data.Store.Statements {
		want[s.ID] = true
	}
	for _, item := range env.Items {
		if !want[item.ID] {
			t.Errorf("exported statement %q is not in the dataset", item.ID)
			continue
		}
		delete(want, item.ID)
		if item.Text == "" || item.Dimension == "" {
			t.Errorf("%s exported without text or dimension", item.ID)
		}
		if !strings.HasPrefix(item.URL, siteBase+"/utsagn/") {
			t.Errorf("%s has url %q", item.ID, item.URL)
		}
		if item.Number <= 0 {
			t.Errorf("%s has number %d", item.ID, item.Number)
		}
	}
	for id := range want {
		t.Errorf("statement %s is in the dataset but not in the export", id)
	}
}

// TestDataPageIsTheFrontDoorToTheFiles covers the page a person reads: it has
// to name every column and every file, and every file it links has to answer.
func TestDataPageIsTheFrontDoorToTheFiles(t *testing.T) {
	w := get(t, "/data")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	body := strings.Join(strings.Fields(w.Body.String()), " ")

	for _, c := range evidenceColumns {
		if !strings.Contains(body, "<code>"+c.Name+"</code>") {
			t.Errorf("/data does not name the column %q", c.Name)
		}
		// The name alone is not the documentation. A table of column names
		// with the descriptions dropped still lists every column, and is
		// exactly as unreadable as no table at all.
		if want := html.EscapeString(c.Doc); !strings.Contains(body, want) {
			t.Errorf("/data does not explain the column %q", c.Name)
		}
	}

	total, withQuote := quoteCounts(data.Store)
	for _, want := range []string{
		`href="/data/utsagn.json"`,
		`href="/data/evidens.json"`,
		`href="/data/evidens.csv"`,
		strconv.Itoa(len(data.Store.Statements)) + " utsagn",
		strconv.Itoa(total) + " påstander",
		strconv.Itoa(withQuote) + " av " + strconv.Itoa(total),
		licenceNote,
		SnapshotDate,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/data missing %q", want)
		}
	}

	for _, path := range []string{"/data/utsagn.json", "/data/evidens.json", "/data/evidens.csv"} {
		if resp := get(t, path); resp.Code != http.StatusOK {
			t.Errorf("%s returned %d", path, resp.Code)
		}
	}
}

// TestSitemapListsTheDataPage keeps the export inside the crawlable site.
// The page is worth nothing to a reader who cannot find it, and it is one
// link away from nowhere if the sitemap does not mention it.
func TestSitemapListsTheDataPage(t *testing.T) {
	body, err := io.ReadAll(get(t, "/sitemap.xml").Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "<loc>"+siteBase+"/data</loc>") {
		t.Error("sitemap does not list /data")
	}
}
