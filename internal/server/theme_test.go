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

// TestThemeBlockShowsEveryPartyOnOneDimension: the theme block is the
// honest answer to "hva sier partiet her" where no precise link exists. It
// must show every party that has points in the statement's own dimension,
// and it must not quietly shrink to the linked ones.
func TestThemeBlockShowsEveryPartyOnOneDimension(t *testing.T) {
	page := buildStatementPage(data.Store, mustStatement(t, "utsagn-01"))
	if len(page.Theme) == 0 {
		t.Fatal("utsagn-01 has no theme block, but Demokratisk åpenhet has programme points from every party")
	}
	want := 0
	for _, p := range data.Store.Parties {
		if len(data.Store.ProgramPointsForDimension(p.ID, "Demokratisk åpenhet")) > 0 {
			want++
		}
	}
	if len(page.Theme) != want {
		t.Errorf("theme block shows %d parties, want %d", len(page.Theme), want)
	}
	counted := 0
	for _, tp := range page.Theme {
		counted += len(tp.Points)
	}
	if counted != page.ThemeCount {
		t.Errorf("ThemeCount is %d but the rows hold %d points", page.ThemeCount, counted)
	}
}

// TestThemeBlockMarksTheLinkedPoints: a point that names this statement is
// marked; a point that only shares the theme is not. The two must not blur,
// or the block makes a promise claim the data does not support.
func TestThemeBlockMarksTheLinkedPoints(t *testing.T) {
	// Every statement, not just one. The first version of this test checked
	// utsagn-01 alone, where the linked set and the theme happen to
	// coincide, so it passed while seven cross-dimension links were being
	// dropped from the page. The loop is the fix, not a wider assertion on
	// the same single case.
	covered := 0
	for _, st := range data.Store.Statements {
		want := len(data.Store.ProgramPointsForStatement(st.ID))
		page := buildStatementPage(data.Store, st)
		if len(page.Bearing) != want {
			t.Errorf("%s: Bearing holds %d points, but %d name it", st.ID, len(page.Bearing), want)
		}
		if want == 0 {
			continue
		}
		covered++
		// A point already shown open must not be repeated in the theme
		// list, and the theme must still hold the rest.
		shown := map[string]bool{}
		for _, b := range page.Bearing {
			shown[b.Party.ID+"\x00"+b.Topic] = true
		}
		for _, tp := range page.Theme {
			for _, pt := range tp.Points {
				if shown[tp.Party.ID+"\x00"+pt.Topic] {
					t.Errorf("%s: %q (%s) is both open and in the theme list", st.ID, pt.Topic, tp.Party.ID)
				}
			}
		}
	}
	if covered != 32 {
		t.Errorf("statements with a linked point: got %d, want 32", covered)
	}
}

// TestThemeBlockAbsentWhereNoPartyHasPoints: three of the twelve dimensions
// have no programme points at all, so the block must be absent rather than
// empty — an empty block would read as "no party has a position here".
func TestThemeBlockAbsentWhereNoPartyHasPoints(t *testing.T) {
	for _, id := range []string{"utsagn-52", "utsagn-46", "utsagn-44"} {
		page := buildStatementPage(data.Store, mustStatement(t, id))
		if len(page.Theme) != 0 {
			t.Errorf("%s is in a dimension with no programme points, but the theme block holds %d parties", id, len(page.Theme))
		}
	}
}

// TestThemeBlockRendersAndKeepsTheChainHonest: the page carries the block,
// names the dimension, and still says the chain reports presence rather
// than causation.
func TestThemeBlockRendersAndKeepsTheChainHonest(t *testing.T) {
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/utsagn/1", nil))
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	body := w.Body.String()
	for _, want := range []string{
		"Partiene om dette utsagnet",
		"Hva partiene sier i dette temaet",
		"Demokratisk åpenhet",
		"Kjeden viser hva vi har funnet, ikke at trinnene henger sammen",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/utsagn/1 does not contain %q", want)
		}
	}
	if n := len(regexp.MustCompile(`class="theme-party"`).FindAllString(body, -1)); n == 0 {
		t.Error("no theme party sections rendered")
	}
}

func mustStatement(t *testing.T, id string) data.Statement {
	t.Helper()
	s, ok := data.Store.Statement(id)
	if !ok {
		t.Fatalf("no statement %s", id)
	}
	return s
}
