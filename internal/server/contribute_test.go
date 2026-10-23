package server

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
	"github.com/jkorsvik/ektevalgomat/web"
)

// render serves one path through the real handler and returns the body with
// whitespace collapsed, so assertions do not depend on template indentation.
func render(t *testing.T, path string) string {
	t.Helper()
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	if w.Code != 200 {
		t.Fatalf("%s returned %d", path, w.Code)
	}
	return strings.Join(strings.Fields(w.Body.String()), " ")
}

// TestMetodeStatesQuoteCoverageFromTheData holds the promise on /metode: a
// visitor is told how many claims they can check on the page and how many
// only carry a pointer to the source.
//
// The numbers are derived at render time (server.quoteCounts). This test
// recomputes them from the dataset and requires the page to agree, so a
// hand-typed pair that drifts the first time a claim is added shows up here
// instead of on the site.
func TestMetodeStatesQuoteCoverageFromTheData(t *testing.T) {
	total, withQuote := quoteCounts(data.Store)
	withoutQuote := total - withQuote

	// A degenerate dataset (all quoted, or none) would let the assertion
	// below pass while proving nothing about the split.
	if total == 0 || withQuote == 0 || withoutQuote == 0 {
		t.Fatalf("degenerate fixture: total=%d withQuote=%d withoutQuote=%d", total, withQuote, withoutQuote)
	}

	body := render(t, "/metode")
	for _, want := range []string{
		fmt.Sprintf("%d av %d påstander", withQuote, total),
		fmt.Sprintf("øvrige %d", withoutQuote),
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/metode missing %q", want)
		}
	}

	if !strings.Contains(body, `id="sitatdekning"`) {
		t.Error(`/metode lost the id="sitatdekning" anchor the claim cards link to`)
	}
}

// TestQuoteLessClaimsSaySo checks the else branch of the evidence card, in
// both directions. A marker that rendered for every claim would mislabel the
// quoted ones; a marker that rendered for none would leave 299 claims looking
// like ordinary sourced claims whose quote was simply omitted.
func TestQuoteLessClaimsSaySo(t *testing.T) {
	tmpl, err := fs.ReadFile(web.Templates, "templates/statement.html.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(tmpl), "sc-noquote") {
		t.Fatal("statement.html.tmpl has no sc-noquote branch: a claim with no stored quote is rendered as if it had one")
	}

	for _, statementID := range []string{"utsagn-01", "utsagn-44"} {
		statement := statementNumber(t, statementID)
		body := render(t, fmt.Sprintf("/utsagn/%d", statement))

		wantWithout, wantWith := 0, 0
		for _, e := range data.Store.Evidence {
			if !e.AppliesToStatement(statementID) {
				continue
			}
			if e.VerbatimExcerpt == "" {
				wantWithout++
			} else {
				wantWith++
			}
		}
		if wantWithout == 0 || wantWith == 0 {
			t.Fatalf("%s: fixture is one-sided (without=%d with=%d)", statementID, wantWithout, wantWith)
		}

		gotWithout := strings.Count(body, "class=\"sc-noquote\"")
		gotWith := strings.Count(body, "class=\"verbatim\"")
		if gotWithout != wantWithout {
			t.Errorf("%s: rendered %d no-quote markers, data has %d quote-less claims", statementID, gotWithout, wantWithout)
		}
		if gotWith != wantWith {
			t.Errorf("%s: rendered %d quotes, data has %d quoted claims", statementID, gotWith, wantWith)
		}
	}
}

func statementNumber(t *testing.T, id string) int {
	t.Helper()
	for _, s := range data.Store.Statements {
		if s.ID == id {
			return s.Number
		}
	}
	t.Fatalf("statement %s not found", id)
	return 0
}

// TestEveryPrefilledIssueTemplateExists is the gate on the contribution
// links. Every "meld fra" and "foreslå endring" link on the site names a
// template file in .github/ISSUE_TEMPLATE. Renaming or moving that file
// would leave the links pointing at a 404 that only appears after a reader
// has clicked, so the names are checked against the files on disk.
func TestEveryPrefilledIssueTemplateExists(t *testing.T) {
	re := regexp.MustCompile(`issueURL "([^"]+)"`)
	names := map[string]bool{}
	for _, f := range []string{"templates/index.html.tmpl", "templates/statement.html.tmpl"} {
		b, err := fs.ReadFile(web.Templates, f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range re.FindAllStringSubmatch(string(b), -1) {
			names[m[1]] = true
		}
	}

	// A guard that finds nothing to check must fail, not pass.
	if len(names) == 0 {
		t.Fatal("no issueURL calls found in the templates: this guard would pass vacuously")
	}

	for name := range names {
		path := "../../.github/ISSUE_TEMPLATE/" + name
		b, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("the site links to template %q but %s does not exist", name, path)
			continue
		}
		text := string(b)
		for _, want := range []string{"---", "name:", "about:"} {
			if !strings.Contains(text, want) {
				t.Errorf("%s is missing %q, so GitHub will not offer it", path, want)
			}
		}
	}
}

// TestContributionGuideNamesEveryGate keeps CONTRIBUTING.md honest about
// what a contributor has to run. If a gate is added and the guide is not
// updated, a PR arrives green locally and red in CI, which is the most
// common way a first contribution dies.
func TestContributionGuideNamesEveryGate(t *testing.T) {
	b, err := os.ReadFile("../../CONTRIBUTING.md")
	if err != nil {
		t.Fatalf("CONTRIBUTING.md is the front door for a contributor: %v", err)
	}
	guide := string(b)
	for _, gate := range []string{
		"gofmt -l .",
		"go vet ./...",
		"go test ./...",
		"node --check",
		"tools/autopilot/test_gate.py",
		"tools/verify/verify_excerpts.py",
	} {
		if !strings.Contains(guide, gate) {
			t.Errorf("CONTRIBUTING.md does not mention the gate %q", gate)
		}
	}
}
