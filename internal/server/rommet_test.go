package server

import (
	"encoding/json"
	"log/slog"
	"math"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
	"github.com/jkorsvik/ektevalgomat/internal/space"
)

func getRommet(t *testing.T) string {
	t.Helper()
	h := New(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/rommet", nil))
	if w.Code != 200 {
		t.Fatalf("/rommet answered %d, want 200", w.Code)
	}
	return w.Body.String()
}

// The page's headline claim is the share of the variation the three axes keep. It
// must be the projection's own number, recomputed here, not a figure typed into a
// template that would keep saying 82,6 after the data moved.
func TestRommetPrintsTheProjectionsOwnShare(t *testing.T) {
	body := getRommet(t)
	sp := space.Project(data.Store)

	want := strings.Replace(strconv.FormatFloat(sp.Explained*100, 'f', 1, 64), ".", ",", -1)
	if !strings.Contains(body, want+" prosent") {
		t.Errorf("the page does not state %q, the share the projection actually keeps", want+" prosent")
	}
	rest := strings.Replace(strconv.FormatFloat((1-sp.Explained)*100, 'f', 1, 64), ".", ",", -1)
	if !strings.Contains(body, rest) {
		t.Errorf("the page does not state the %s %% that falls outside the three axes: "+
			"a projection that hides its own loss lies by omission", rest)
	}
	if !strings.Contains(body, "standardisert") {
		t.Error("the page does not say the dimensions were standardised, which is the choice " +
			"that keeps the first axis from being an artefact of statement counts")
	}
}

// The payload the canvas draws must be the projection, to the last digit. A picture
// built from a stale or rounded copy of the data would be a drawing of a claim.
func TestRommetPayloadIsTheProjection(t *testing.T) {
	body := getRommet(t)
	re := regexp.MustCompile(`(?s)<script id="rommet-data" type="application/json">(.*?)</script>`)
	m := re.FindStringSubmatch(body)
	if m == nil {
		t.Fatal("no rommet-data JSON block: the canvas would have nothing to draw")
	}
	var p struct {
		Explained float64 `json:"explained"`
		Axes      []struct {
			Variance float64   `json:"variance"`
			Loadings []float64 `json:"loadings"`
		} `json:"axes"`
		Dims []struct {
			Name       string  `json:"name"`
			Statements int     `json:"statements"`
			Coded      int     `json:"coded"`
			Variance   float64 `json:"variance"`
		} `json:"dims"`
		Clouds []struct {
			ID     string       `json:"id"`
			Centre [3]float64   `json:"centre"`
			Spread float64      `json:"spread"`
			Points [][3]float64 `json:"points"`
		} `json:"clouds"`
		Springs []struct {
			A          string  `json:"a"`
			B          string  `json:"b"`
			Agreement  float64 `json:"agreement"`
			Comparable int     `json:"comparable"`
		} `json:"springs"`
	}
	if err := json.Unmarshal([]byte(m[1]), &p); err != nil {
		t.Fatalf("rommet-data is not valid JSON: %v", err)
	}

	sp := space.Project(data.Store)
	if math.Abs(p.Explained-sp.Explained) > 1e-12 {
		t.Errorf("payload explained = %.6f, projection = %.6f", p.Explained, sp.Explained)
	}
	if len(p.Dims) != len(sp.Dimensions) {
		t.Fatalf("payload has %d dimensions, projection %d", len(p.Dims), len(sp.Dimensions))
	}
	if len(p.Clouds) != len(sp.Clouds) {
		t.Fatalf("payload has %d clouds, projection %d", len(p.Clouds), len(sp.Clouds))
	}
	if len(p.Springs) != len(sp.Springs) {
		t.Errorf("payload has %d springs, projection %d", len(p.Springs), len(sp.Springs))
	}
	for i, c := range sp.Clouds {
		got := p.Clouds[i]
		if got.ID != c.PartyID {
			t.Fatalf("cloud %d is %q, want %q", i, got.ID, c.PartyID)
		}
		for a := 0; a < 3; a++ {
			if math.Abs(got.Centre[a]-c.Centre[a]) > 1e-9 {
				t.Errorf("%s centre axis %d = %.6f, want %.6f", c.PartyID, a+1, got.Centre[a], c.Centre[a])
			}
		}
		if len(got.Points) != len(c.Points) {
			t.Errorf("%s has %d points, want %d (one per dimension)", c.PartyID, len(got.Points), len(c.Points))
		}
	}
	// The loadings line up with Dims, so a client can zip them without guessing.
	for ai, ax := range p.Axes {
		if len(ax.Loadings) != len(p.Dims) {
			t.Errorf("axis %d has %d loadings for %d dimensions", ai+1, len(ax.Loadings), len(p.Dims))
		}
	}
}

// The drawing is an addition to the page, not a replacement for the numbers. A
// reader without WebGL, or with a screen reader, must still get every figure.
func TestRommetCarriesTheNumbersWithoutWebGL(t *testing.T) {
	body := getRommet(t)
	sp := space.Project(data.Store)
	for _, c := range sp.Clouds {
		if !strings.Contains(body, c.PartyID) {
			t.Errorf("party %q is missing from the page", c.PartyID)
		}
	}
	if !strings.Contains(body, "<table") {
		t.Error("no table: without WebGL the page would show a picture it cannot draw")
	}
	if !strings.Contains(body, "rommet-canvas") {
		t.Error("no canvas element")
	}
	// A thin dimension must be visible as thin. Generasjonsansvar rests on two
	// statements, and a page that hides that gives it the same authority as one
	// with seven.
	if !strings.Contains(body, "Generasjonsansvar") {
		t.Error("the dimensions table does not list Generasjonsansvar, so a reader cannot " +
			"see how thinly some dimensions are supported")
	}
}

// The exploration is the page's point: a reader must be able to choose what the
// drawing shows and to ask one party for its own numbers. Each control is named
// here, because a control that disappears from the template leaves the canvas
// readable but the space unexplorable, and nothing else would notice.
func TestRommetCarriesTheExplorationControls(t *testing.T) {
	body := getRommet(t)
	// Match the id attribute exactly. A substring check would pass on
	// "rommet-springs-renamed", which is how this guard first failed to fail:
	// renaming the control left the test green while the script found nothing.
	for _, id := range []string{"rommet-dim", "rommet-springs", "rommet-zoom", "rommet-focus"} {
		if !strings.Contains(body, `id="`+id+`"`) {
			t.Errorf("the page has no %q control, so the drawing cannot be explored", id)
		}
	}
	// Every dimension must be offered by the selector, or a reader cannot reach it.
	p := rommetPayload(t)
	for _, d := range p.Dims {
		if !strings.Contains(body, d.Name) {
			t.Errorf("dimension %q is not offered anywhere on the page", d.Name)
		}
	}
	// The focus panel reads these two off the payload; empty ones would render an
	// unnamed party and a blank answer count.
	for _, c := range p.Clouds {
		if c.Short == "" {
			t.Errorf("party %q carries no short name, so the panel cannot name it", c.ID)
		}
		if c.Answers == 0 {
			t.Errorf("party %q carries no answer count", c.ID)
		}
		if len(c.Points) != len(p.Dims) {
			t.Errorf("party %q has %d dimension offsets for %d dimensions, so highlighting "+
				"a dimension would draw nothing for it", c.ID, len(c.Points), len(p.Dims))
		}
	}
}

// rommetPayload is the canvas payload, parsed from the page it is served in. Parsing
// the rendered page rather than structs built in the test keeps the check on what
// the browser actually receives.
func rommetPayload(t *testing.T) struct {
	Explained float64 `json:"explained"`
	Dims      []struct {
		Name string `json:"name"`
	} `json:"dims"`
	Clouds []struct {
		ID      string       `json:"id"`
		Short   string       `json:"short"`
		Answers int          `json:"answers"`
		Points  [][3]float64 `json:"points"`
	} `json:"clouds"`
} {
	t.Helper()
	body := getRommet(t)
	re := regexp.MustCompile(`(?s)<script id="rommet-data" type="application/json">(.*?)</script>`)
	m := re.FindStringSubmatch(body)
	if m == nil {
		t.Fatal("no rommet-data JSON block: the canvas would have nothing to draw")
	}
	var p struct {
		Explained float64 `json:"explained"`
		Dims      []struct {
			Name string `json:"name"`
		} `json:"dims"`
		Clouds []struct {
			ID      string       `json:"id"`
			Short   string       `json:"short"`
			Answers int          `json:"answers"`
			Points  [][3]float64 `json:"points"`
		} `json:"clouds"`
	}
	if err := json.Unmarshal([]byte(m[1]), &p); err != nil {
		t.Fatalf("rommet-data is not valid JSON: %v", err)
	}
	return p
}
