package server

import (
	"log/slog"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// TestEveryEvidenceTypeIsPlaced: the chain groups evidence by type, so a type
// that is in neither a step nor the context list would vanish from the page
// with nothing failing. This makes that a decision instead of an oversight.
func TestEveryEvidenceTypeIsPlaced(t *testing.T) {
	placed := map[string]bool{}
	for _, s := range chainStageOrder {
		for _, ty := range s.Types {
			placed[ty] = true
		}
	}
	for _, ty := range chainContextTypes {
		placed[ty] = true
	}
	seen := map[string]bool{}
	for _, e := range data.Store.Evidence {
		seen[e.EvidenceType] = true
	}
	if len(seen) < 10 {
		t.Fatalf("only %d evidence types in the dataset, expected at least 10", len(seen))
	}
	for ty := range seen {
		if !placed[ty] {
			t.Errorf("evidence type %q is in neither a chain step nor the context list", ty)
		}
	}
}

// TestChainAlwaysShowsEveryStep: an empty step is the finding. The chain must
// render all five steps for a statement that has evidence, and must never
// silently drop one just because it holds nothing.
func TestChainAlwaysShowsEveryStep(t *testing.T) {
	// Every statement id must resolve, or the test is checking nothing. This
	// also caught a real trap: the ids are not one form. Statements 1-9 are
	// "utsagn-01" and the rest are "utsagn-45", and AppliesToStatement is a
	// plain string compare, so an id written in the other form silently
	// matches no evidence at all. (No reference does that today: all 60
	// resolve.) ThemeTest below pins the form.
	resolved := 0
	var withEvidence []string
	for _, s := range data.Store.Statements {
		stages := buildChain(data.Store, s.ID)
		if len(stages) != len(chainStageOrder) {
			t.Fatalf("%s: %d steps, want %d", s.ID, len(stages), len(chainStageOrder))
		}
		if chainHasAny(stages) {
			resolved++
			withEvidence = append(withEvidence, s.ID)
		}
	}
	if resolved < 10 {
		t.Fatalf("only %d statements have any chain evidence, expected at least 10", resolved)
	}
	// The chain must actually differ between statements, or it is decoration.
	var first, second []chainStage
	for _, s := range data.Store.Statements {
		if s.ID == "utsagn-45" {
			first = buildChain(data.Store, s.ID)
		}
		if s.ID == "utsagn-60" {
			second = buildChain(data.Store, s.ID)
		}
	}
	if first == nil || second == nil {
		t.Fatal("utsagn-45 or utsagn-60 is missing from the dataset")
	}
	same := true
	for i := range first {
		if first[i].Count != second[i].Count {
			same = false
		}
	}
	if same {
		t.Error("utsagn-45 and utsagn-60 have identical chains, so the component is not reading the data")
	}
}

// TestChainCountsMatchTheEvidence: the number on a step is the number of
// sources at that step, not an estimate. Summing the steps must equal the
// chain-typed evidence, plus the programme points that name this statement
// — the promise step's other real input, set by hand (2026-10-13). Both
// halves are counted, so a step that drops or invents a source still fails.
func TestChainCountsMatchTheEvidence(t *testing.T) {
	checked := 0
	for _, s := range data.Store.Statements {
		stages := buildChain(data.Store, s.ID)
		got := 0
		for _, st := range stages {
			got += st.Count
		}
		want := 0
		for _, e := range data.Store.EvidenceForStatement(s.ID) {
			if _, ok := chainTypeToStage[e.EvidenceType]; ok {
				want++
			}
		}
		want += len(data.Store.ProgramPointsForStatement(s.ID))
		if got != want {
			t.Errorf("%s: chain holds %d sources, evidence plus linked programme points hold %d", s.ID, got, want)
		}
		checked++
	}
	if checked < 50 {
		t.Fatalf("only checked %d statements, expected at least 50", checked)
	}
}

// TestChainRenders: the band is on the page, names all five steps, and says
// how many are empty.
func TestChainRenders(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/utsagn/60", nil))
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Kjeden for dette utsagnet") {
		t.Fatal("the chain is not on the page")
	}
	for _, label := range []string{"Løfte", "Forslag", "Stemme", "Vedtak", "Utfall"} {
		if !strings.Contains(body, ">"+label+"<") {
			t.Errorf("the chain does not name the step %q", label)
		}
	}
	if !regexp.MustCompile(`trinn mangler kilde`).MatchString(body) {
		t.Error("the chain does not say how many steps are empty")
	}
	if !strings.Contains(body, "Ingen kilde på dette trinnet") {
		t.Error("no empty step is shown as empty")
	}
}
