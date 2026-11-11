package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The person records are the only place the site stores facts about named
// individuals. These tests hold the two lines that matter: every role is
// sourced and dated, and no private field ever reaches the embedded content.

func TestPeopleAreLoadableAndComplete(t *testing.T) {
	ppl := People()
	if len(ppl) < 150 {
		t.Fatalf("only %d person records; the register covers 169 sitting representatives", len(ppl))
	}
	seen := map[string]bool{}
	withMandate := 0
	for _, p := range ppl {
		if seen[p.ID] {
			t.Errorf("duplicate person id %s", p.ID)
		}
		seen[p.ID] = true
		if p.Name == "" || p.Scope == "" || p.ConsentBasis != "public_office" {
			t.Errorf("%s: incomplete header (%q %q %q)", p.ID, p.Name, p.Scope, p.ConsentBasis)
		}
		mandate := false
		for _, r := range p.Roles {
			if r.Role == "stortingsrepresentant" {
				mandate = true
			}
			if r.SourceURL == "" || r.From == "" || r.SourceTier != "A" || r.RetrievedAt == "" {
				t.Errorf("%s/%s: role not fully sourced: url=%q from=%q tier=%q retrieved=%q",
					p.ID, r.Role, r.SourceURL, r.From, r.SourceTier, r.RetrievedAt)
			}
		}
		if mandate {
			withMandate++
		}
	}
	if withMandate != len(ppl) {
		t.Errorf("%d of %d records carry a mandate role", withMandate, len(ppl))
	}
}

func TestNoPersonRoleEndsAfterSnapshot(t *testing.T) {
	for _, p := range People() {
		for _, r := range p.Roles {
			if r.To == "" {
				continue
			}
			if r.To > "2026-10-12" {
				t.Errorf("%s/%s ends %s, after the snapshot 2026-10-12", p.ID, r.Role, r.To)
			}
			if r.From > r.To {
				t.Errorf("%s/%s starts %s and ends %s", p.ID, r.Role, r.From, r.To)
			}
		}
	}
}

// TestRegisterSnapshotMatchesIndex is the drift gate for the pinned constant.
// The index is a build artefact outside content/, so nothing else compares them.
//
// The index is committed, so a missing file is a defect and not an environment
// condition: this gate failed open for weeks because the path resolved to
// internal/verv/index.json, the read errored, and the error was turned into a
// skip. A gate that skips when its input is absent is green and blind.
func TestRegisterSnapshotMatchesIndex(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("verv", "index.json"))
	if err != nil {
		t.Fatalf("verv index is committed and must be readable: %v", err)
	}
	var idx struct {
		SnapshotDate string `json:"snapshot_date"`
		People       []struct {
			Name  string `json:"name"`
			Roles []struct {
				Section int `json:"section"`
			} `json:"roles"`
		} `json:"people"`
	}
	if err := json.Unmarshal(b, &idx); err != nil {
		t.Fatalf("index.json: %v", err)
	}
	if idx.SnapshotDate != RegisterSnapshotDate {
		t.Errorf("pinned RegisterSnapshotDate %q but index.json says %q; rebuild or repin",
			RegisterSnapshotDate, idx.SnapshotDate)
	}
	if len(idx.People) < 150 {
		t.Errorf("index.json holds %d people", len(idx.People))
	}
}

// TestNoForbiddenFieldInPeopleFiles scans the raw embedded YAML, not the parsed
// struct, so a field the struct does not model still fails the test.
func TestNoForbiddenFieldInPeopleFiles(t *testing.T) {
	forbidden := []string{
		"personnummer", "fodselsnummer", "foedselsdato", "birth",
		"privat_adresse", "private_address", "adresse",
		"telefon", "phone", "privat_epost", "private_email",
		"familie", "family", "ektefelle", "spouse", "barn", "children",
		"helse", "health", "sykdom", "diagnose", "religion",
		"seksuell_orientering", "bankkonto", "inntekt_privat", "formue_privat",
	}
	entries, err := contentFS.ReadDir(contentRoot + "/people")
	if err != nil {
		t.Fatalf("reading people: %v", err)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		b, err := contentFS.ReadFile(contentRoot + "/people/" + e.Name())
		if err != nil {
			t.Fatalf("%s: %v", e.Name(), err)
		}
		lower := strings.ToLower(string(b))
		for _, f := range forbidden {
			if strings.Contains(lower, f+":") {
				t.Errorf("%s carries forbidden field %q", e.Name(), f)
			}
		}
	}
}

// TestNoAmountInPersonRoles is the boundary gate for declared interests: the
// register discloses the interest, not its size, and the site must not carry a
// kroner amount or a share percentage in a role name.
func TestNoAmountInPersonRoles(t *testing.T) {
	amount := regexp.MustCompile(`\d[\d\s.,]*\s*(%|pst|NOK|kr\b)`)
	for _, p := range People() {
		for _, r := range p.Roles {
			if amount.MatchString(r.Organisation) || amount.MatchString(r.Post) {
				t.Errorf("%s/%s carries an amount: %q", p.ID, r.Role, r.Organisation)
			}
		}
	}
}
