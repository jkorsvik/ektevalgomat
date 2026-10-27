package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"
)

// The audits index is generated data: Riksrevisjonen's most recent reports, from
// riksrevisjonen.no, built by tools/analytics/build_audits_index.py.
//
// Riksrevisjonen is the Storting's external auditor, so these reports are the
// control on everything else in this project: the body that tells the Storting
// whether the money was spent as decided. Without them the map describes power and
// never checks it.
//
// The index is partial and says so. The site exposes no sitemap and no working
// pagination, so only the listing page's most recent reports can be enumerated;
// the limit is stored under not_collectable with the attempts that established it.
//
//go:embed audits/index.json
var auditsIndexJSON []byte

// AuditReport is one published Riksrevisjonen report.
type AuditReport struct {
	URL        string   `json:"url"`
	Session    string   `json:"session"`
	Published  string   `json:"published"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Categories []string `json:"categories"`
}

type auditsSource struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	RetrievedAt string `json:"retrieved_at"`
	Note        string `json:"note"`
}

type auditsIndexFile struct {
	Source         auditsSource       `json:"source"`
	Reports        []AuditReport      `json:"reports"`
	NotCollectable map[string]Refusal `json:"not_collectable"`
}

var auditsIndexDoc = mustLoadAudits()

func mustLoadAudits() auditsIndexFile {
	var idx auditsIndexFile
	if err := json.Unmarshal(auditsIndexJSON, &idx); err != nil {
		panic(fmt.Sprintf("internal/data/audits/index.json is not valid: %v", err))
	}
	if len(idx.Reports) == 0 {
		panic("internal/data/audits/index.json holds no reports; run tools/analytics/build_audits_index.py")
	}
	if idx.Source.RetrievedAt == "" || idx.Source.URL == "" {
		panic("internal/data/audits/index.json has no provenance")
	}
	return idx
}

// AuditReports returns the reports, newest first.
func AuditReports() []AuditReport { return auditsIndexDoc.Reports }

// AuditsSourceInfo returns the provenance of the index.
func AuditsSourceInfo() (name, url, retrievedAt, note string) {
	s := auditsIndexDoc.Source
	return s.Name, s.URL, s.RetrievedAt, s.Note
}

// AuditArchiveRefusal returns why the archive behind the listing is not collected.
func AuditArchiveRefusal() (Refusal, bool) {
	r, ok := auditsIndexDoc.NotCollectable["archive"]
	return r, ok
}

// LatestAudit returns the most recently published report, and whether any exist.
func LatestAudit() (AuditReport, bool) {
	if len(auditsIndexDoc.Reports) == 0 {
		return AuditReport{}, false
	}
	return auditsIndexDoc.Reports[0], true
}

// AuditPublishedDate parses a report's date, so a caller does not have to trust the
// string's shape.
func AuditPublishedDate(r AuditReport) (time.Time, error) {
	return time.Parse("2006-01-02", r.Published)
}
