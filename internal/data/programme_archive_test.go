package data

import (
	"crypto/sha256"
	"encoding/hex"
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
