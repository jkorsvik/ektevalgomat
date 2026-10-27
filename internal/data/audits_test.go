package data

import (
	"regexp"
	"testing"
)

var sessionPattern = regexp.MustCompile(`^no-\d{4}-\d{4}$`)

// The audits index is a copy of a listing page, so the tests are about the copy
// being faithful and about the partial coverage staying visible.
func TestAuditsIndexIsUsable(t *testing.T) {
	reports := AuditReports()
	// A band, not a pin: the listing page changes every time Riksrevisjonen
	// publishes, so an exact count would break on the next publication. The band
	// still catches a parse that silently drops reports or a pagination change.
	if len(reports) < 5 || len(reports) > 12 {
		t.Fatalf("%d reports: the listing page carries about ten", len(reports))
	}
	seen := map[string]bool{}
	for _, r := range reports {
		if r.Title == "" {
			t.Errorf("%s has no title", r.URL)
		}
		if r.URL == "" || seen[r.URL] {
			t.Errorf("missing or duplicate url %q", r.URL)
		}
		seen[r.URL] = true
		if _, err := AuditPublishedDate(r); err != nil {
			t.Errorf("%s: published %q is not an ISO date: %v", r.URL, r.Published, err)
		}
		if !sessionPattern.MatchString(r.Session) {
			t.Errorf("%s: session %q is not no-YYYY-YYYY", r.URL, r.Session)
		}
	}
}

// Newest first is part of the contract, so a page renders deterministically.
func TestAuditsAreOrderedNewestFirst(t *testing.T) {
	reports := AuditReports()
	for i := 1; i < len(reports); i++ {
		prev, err1 := AuditPublishedDate(reports[i-1])
		cur, err2 := AuditPublishedDate(reports[i])
		if err1 != nil || err2 != nil {
			t.Fatalf("unparseable date at %d", i)
		}
		if cur.After(prev) {
			t.Errorf("%s is newer than %s but listed after it", reports[i].Published, reports[i-1].Published)
		}
	}
}

// The partial coverage must be discoverable from the data, not only from a comment
// in the builder. A caller that treats this index as the full archive is wrong, and
// the index has to be able to say so.
func TestAuditArchiveLimitationIsRecorded(t *testing.T) {
	refusal, ok := AuditArchiveRefusal()
	if !ok {
		t.Fatal("no archive limitation recorded: the index would read as the complete archive")
	}
	if len(refusal.Reason) < 40 {
		t.Errorf("limitation reason is too thin: %q", refusal.Reason)
	}
	if len(refusal.Attempts) < 3 {
		t.Errorf("only %d attempts recorded", len(refusal.Attempts))
	}
	for _, a := range refusal.Attempts {
		if len(a) == 0 {
			t.Error("empty attempt record")
		}
	}
}

func TestAuditsCarryProvenanceAndSayWhatTheyAreNot(t *testing.T) {
	name, url, retrieved, note := AuditsSourceInfo()
	if name == "" || url == "" || retrieved == "" {
		t.Errorf("incomplete provenance: %q %q %q", name, url, retrieved)
	}
	if !contains(note, "not the whole archive") {
		t.Errorf("source note does not state the coverage limit: %q", note)
	}
	latest, ok := LatestAudit()
	if !ok {
		t.Fatal("no latest report")
	}
	if latest.Published != AuditReports()[0].Published {
		t.Errorf("LatestAudit returned %s, the list starts at %s", latest.Published, AuditReports()[0].Published)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}
