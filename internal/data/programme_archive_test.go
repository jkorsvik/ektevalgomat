package data

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Every programme the site quotes from must have its own document kept, with
// the hash and the time of the fetch, because a party URL is not a source the
// project controls: Rødt answered 403/429 to every automated request on
// 2026-10-25, and Venstre publishes its programme as a web page rather than
// a PDF. The hash is what /partier shows, so it has to be the hash of the
// bytes the quotes were read from — not of whatever the file became later.
//
// This checks the files, not the manifest: a manifest compared with itself
// agrees with itself while the YAML drifts away from it.
func TestArchivedProgrammesMatchYAML(t *testing.T) {
	if len(Store.Programs) == 0 {
		t.Fatal("no programmes loaded: the gate would pass vacuously")
	}

	checked := 0
	for _, p := range Store.Programs {
		if p.DocumentSHA256 == "" {
			t.Errorf("%s: no document_sha256; the quote is only checkable while the party's URL lives", p.ID)
			continue
		}
		if !sha256Hex.MatchString(p.DocumentSHA256) {
			t.Errorf("%s: document_sha256 %q is not a lowercase hex sha256", p.ID, p.DocumentSHA256)
			continue
		}
		if p.DocumentFile == "" {
			t.Errorf("%s: document_sha256 set but document_file empty", p.ID)
			continue
		}
		// The YAML path is relative to the repository root, so a reader can
		// find the file from the path shown on the page.
		b, err := os.ReadFile(filepath.Join("..", "..", p.DocumentFile))
		if err != nil {
			t.Errorf("%s: kept document %s unreadable: %v", p.ID, p.DocumentFile, err)
			continue
		}
		if p.DocumentBytes != int64(len(b)) {
			t.Errorf("%s: %s is %d bytes, yaml says %d", p.ID, p.DocumentFile, len(b), p.DocumentBytes)
		}
		sum := sha256.Sum256(b)
		if got := hex.EncodeToString(sum[:]); got != p.DocumentSHA256 {
			t.Errorf("%s: sha256 of %s is %s, yaml says %s", p.ID, p.DocumentFile, got, p.DocumentSHA256)
		}
		if p.FetchedAt == "" {
			t.Errorf("%s: no fetched_at; a hash without the time of the fetch cannot be reproduced", p.ID)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("checked no programme documents: the gate would pass vacuously")
	}
	t.Logf("verified %d kept programme documents against their published hashes", checked)
}

// programmeManifest is the fetch record: what URL answered, with which status
// and content type, at what time, and the hash of the bytes that arrived. It
// duplicates facts that also sit in programs/*.yaml, which is the point of
// checking it: two records of one fetch can disagree, and a manifest nobody
// reads back is a receipt the project wrote to itself.
type programmeManifest struct {
	Documents map[string]struct {
		File        string `json:"file"`
		SourceURL   string `json:"source_url"`
		HTTPStatus  string `json:"http_status"`
		ContentType string `json:"content_type"`
		Bytes       int64  `json:"bytes"`
		SHA256      string `json:"sha256"`
		FetchedAt   string `json:"fetched_at"`
		Magic       string `json:"magic"`
		FetchNote   string `json:"fetch_note"`
	} `json:"documents"`
}

// TestProgrammeManifestAgreesWithYAMLAndFiles holds the three records of one
// fetch to each other: the file on disk (checked above), the YAML the site
// renders, and the manifest written at fetch time. Any two of them agreeing
// while the third drifts is the failure this catches, which TestArchived-
// ProgrammesMatchYAML alone cannot see because it never opens the manifest.
func TestProgrammeManifestAgreesWithYAMLAndFiles(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "research", "raw", "programs", "manifest.json"))
	if err != nil {
		t.Fatalf("no fetch record at research/raw/programs/manifest.json: %v", err)
	}
	var m programmeManifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("manifest.json is not valid JSON: %v", err)
	}
	if len(m.Documents) == 0 {
		t.Fatal("manifest lists no documents: the check would pass vacuously")
	}

	byFile := map[string]string{} // file -> party id
	byID := map[string]Program{}
	for _, p := range Store.Programs {
		byFile[filepath.Base(p.DocumentFile)] = p.ID
		byID[p.ID] = p
	}

	checked := 0
	for id, doc := range m.Documents {
		p, ok := byID[id]
		if !ok {
			t.Errorf("manifest keeps %s, which no programme in programs/*.yaml points at", id)
			continue
		}
		if got := byFile[doc.File]; got != id {
			t.Errorf("manifest file %q belongs to %q, but %s claims it", doc.File, got, id)
		}
		if doc.File != filepath.Base(p.DocumentFile) {
			t.Errorf("%s: manifest keeps %q, yaml says %q", id, doc.File, p.DocumentFile)
		}
		if doc.SHA256 != p.DocumentSHA256 {
			t.Errorf("%s: manifest sha256 %s, yaml says %s", id, doc.SHA256, p.DocumentSHA256)
		}
		if doc.Bytes != p.DocumentBytes {
			t.Errorf("%s: manifest says %d bytes, yaml says %d", id, doc.Bytes, p.DocumentBytes)
		}
		if doc.FetchedAt != p.FetchedAt {
			t.Errorf("%s: manifest fetched_at %q, yaml says %q", id, doc.FetchedAt, p.FetchedAt)
		}
		if doc.SourceURL == "" || doc.HTTPStatus == "" || doc.ContentType == "" {
			t.Errorf("%s: manifest is missing the fetch itself (url/status/content-type)", id)
		}
		checked++
	}
	for _, p := range Store.Programs {
		if _, ok := m.Documents[p.ID]; !ok {
			t.Errorf("%s has document_file %q but no record in the manifest", p.ID, p.DocumentFile)
		}
	}
	if checked == 0 {
		t.Fatal("compared no manifest records")
	}
	t.Logf("%d manifest records agree with the yaml and the files", checked)
}
