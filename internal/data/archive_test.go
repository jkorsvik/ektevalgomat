package data

import (
	"regexp"
	"testing"
	"unicode/utf8"
)

// The archive is only honest if it covers the claims it claims to cover, and
// only useful if each window can be tied to one retrieval of one document.
// Coverage is checked against the dataset, never against the archive's own
// manifest: a manifest derived from the archive would agree with itself while
// the site's claims drifted away from it.
//
// The one thing deliberately not checked here is whether the quote appears in
// the window. tools/archive/check_archive.py owns that, because it owns the
// normalisation (whitespace, quotes, dashes, soft hyphens) that makes a
// verbatim comparison fair. A second normaliser in Go, however carefully
// written, would eventually disagree with the first, and the disagreement
// would look like a bad claim instead of a bad test.
func TestArchiveCoversEveryQuotedClaim(t *testing.T) {
	quoted := 0
	for _, e := range Store.Evidence {
		if e.VerbatimExcerpt == "" {
			continue
		}
		quoted++
		w, ok := SourceWindowFor(e.ID)
		if !ok {
			t.Errorf("%s carries a verbatim excerpt but has no archived window", e.ID)
			continue
		}
		if w.SourceURL != e.SourceURL {
			t.Errorf("%s: window is for %s, evidence cites %s", e.ID, w.SourceURL, e.SourceURL)
		}
	}
	if quoted == 0 {
		t.Fatal("checked no quoted claims: the dataset loaded without excerpts, which is not a pass")
	}
	if quoted != SourceWindowCount() {
		t.Errorf("archive has %d windows for %d quoted claims: an extra window means a claim was dropped from the data without dropping its window",
			SourceWindowCount(), quoted)
	}
	t.Logf("%d quoted claims, %d windows, %d evidence rows total", quoted, SourceWindowCount(), len(Store.Evidence))
}

var (
	sha256Hex = regexp.MustCompile(`^[0-9a-f]{64}$`)
	isoDate   = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// A window with no provenance is an anonymous quotation. Each field here is
// the reason a reader can go from "the site says X" to "this document said X
// on this date", so a missing one is a real defect rather than a nit.
func TestEveryWindowCarriesItsProvenance(t *testing.T) {
	checked := 0
	for _, e := range Store.Evidence {
		w, ok := SourceWindowFor(e.ID)
		if !ok {
			continue
		}
		checked++
		switch {
		case w.SourceURL == "":
			t.Errorf("%s: no source_url", e.ID)
		case w.FetchedURL == "":
			t.Errorf("%s: no fetched_url", e.ID)
		case !isoDate.MatchString(w.FetchedAt):
			t.Errorf("%s: fetched_at %q is not an ISO date", e.ID, w.FetchedAt)
		case !sha256Hex.MatchString(w.DocumentSHA256):
			t.Errorf("%s: document_sha256 %q is not a sha256", e.ID, w.DocumentSHA256)
		case w.DocumentBytes == 0:
			t.Errorf("%s: no document_bytes", e.ID)
		case w.WindowChars == 0:
			t.Errorf("%s: no window_chars", e.ID)
		// Runes, not bytes: window_chars is counted in Python, where len() is
		// code points, and Norwegian text makes the two disagree by exactly the
		// number of æøå in the window. Comparing bytes here failed 316 windows
		// that were all correct.
		case utf8.RuneCountInString(windowText(w)) != w.WindowChars:
			t.Errorf("%s: window is %d chars but records %d; the text and the manifest disagree",
				e.ID, utf8.RuneCountInString(windowText(w)), w.WindowChars)
		case w.Fragments < 1:
			t.Errorf("%s: records %d fragments", e.ID, w.Fragments)
		}
	}
	if checked == 0 {
		t.Fatal("checked no windows")
	}
	t.Logf("provenance complete on %d windows", checked)
}

// The window must be long enough to show what the quote meant, and short
// enough to stay a window: an earlier spanning design produced one 358 KB
// "window" that quietly republished most of a Stortinget report.
func TestWindowsStayBounded(t *testing.T) {
	const limit = 6000
	longest := 0
	for _, e := range Store.Evidence {
		w, ok := SourceWindowFor(e.ID)
		if !ok {
			continue
		}
		if w.WindowChars > longest {
			longest = w.WindowChars
		}
		if w.WindowChars > limit {
			t.Errorf("%s: window is %d chars, over the %d limit; that is a document, not a window",
				e.ID, w.WindowChars, limit)
		}
		if w.WindowChars < w.ExcerptChars {
			t.Errorf("%s: window (%d) is shorter than the excerpt it is meant to hold (%d)",
				e.ID, w.WindowChars, w.ExcerptChars)
		}
	}
	t.Logf("longest window %d chars", longest)
}
