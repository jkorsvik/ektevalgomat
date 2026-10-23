package data

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

// The archive is a bounded window of a source document around each quoted
// excerpt: roughly 400 characters either side of every fragment, joined by
// " […] " when a claim rests on two places far apart. It exists because the
// claim and the citation are not enough on their own — a source that moves,
// rewrites or deletes a page turns "check it yourself" into a dead link.
//
// What it is not: the document. Republishing whole reports, news articles and
// party programmes was never on the table, for size and licence reasons both.
// The window carries the one thing that cannot be recovered when a link dies,
// which is the words the claim actually rests on.
//
// The tooling that builds and re-checks it lives in tools/archive/ and is
// committed with its own README, so a reader can rebuild what is served here
// and compare.
//
//go:embed archive/entries.jsonl
var archiveEntries []byte

//go:embed archive/windows
var archiveWindows embed.FS

// SourceWindow is one stored window plus the provenance that makes it
// checkable. A window with no provenance is an anonymous quotation, so every
// field below is recorded rather than inferred at read time.
type SourceWindow struct {
	EvidenceID     string `json:"evidence_id"`
	SourceURL      string `json:"source_url"`  // as the data wrote it
	FetchedURL     string `json:"fetched_url"` // after widening (Stortinget pages)
	FetchedAt      string `json:"fetched_at"`  // when the bytes were read
	DocumentSHA256 string `json:"document_sha256"`
	DocumentBytes  int    `json:"document_bytes"`
	WindowFile     string `json:"window_file"`
	WindowChars    int    `json:"window_chars"`
	ExcerptChars   int    `json:"excerpt_chars"`
	Fragments      int    `json:"fragments"`
	Text           string `json:"-"` // the window itself, loaded from WindowFile
}

// sourceWindows is loaded once at start. Like content/, the archive is
// embedded: the runtime container is FROM scratch with no writable paths, so
// there is no filesystem to read it from.
var sourceWindows = mustLoadWindows()

// SourceWindowFor returns the stored window for one evidence row.
func SourceWindowFor(evidenceID string) (SourceWindow, bool) {
	w, ok := sourceWindows[evidenceID]
	return w, ok
}

// SourceWindowCount is how many windows are embedded.
func SourceWindowCount() int { return len(sourceWindows) }

func mustLoadWindows() map[string]SourceWindow {
	windows := make(map[string]SourceWindow)
	sc := bufio.NewScanner(bytes.NewReader(archiveEntries))
	sc.Buffer(make([]byte, 0, 1<<16), 1<<22)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var w SourceWindow
		if err := json.Unmarshal(line, &w); err != nil {
			panic(fmt.Sprintf("archive/entries.jsonl: %v", err))
		}
		if w.WindowFile == "" {
			panic(fmt.Sprintf("archive/entries.jsonl: %s has no window_file", w.EvidenceID))
		}
		body, err := archiveWindows.ReadFile("archive/" + w.WindowFile)
		if err != nil {
			panic(fmt.Sprintf("archive: %s: %v", w.WindowFile, err))
		}
		w.Text = string(body)
		windows[w.EvidenceID] = w
	}
	if err := sc.Err(); err != nil {
		panic(fmt.Sprintf("archive/entries.jsonl: %v", err))
	}
	// A page that renders an archive must not be able to render an empty one
	// while looking healthy.
	if len(windows) == 0 {
		panic("archive: no windows embedded")
	}
	return windows
}

// windowText returns the window with its trailing newline removed, so a
// comparison against the recorded character count is exact.
func windowText(w SourceWindow) string {
	return strings.TrimSuffix(w.Text, "\n")
}
