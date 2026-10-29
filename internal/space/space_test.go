package space

import (
	"math"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

func TestProjectionCoversEveryPartyAndDimension(t *testing.T) {
	sp := Project(data.Store)
	if len(sp.Clouds) != 9 {
		t.Fatalf("clouds = %d, want 9 parties", len(sp.Clouds))
	}
	if len(sp.Dimensions) != 12 {
		t.Fatalf("dimensions = %d, want 12", len(sp.Dimensions))
	}
	if len(sp.Axes) != Axes {
		t.Fatalf("axes = %d, want %d", len(sp.Axes), Axes)
	}
	points := 0
	for _, c := range sp.Clouds {
		points += len(c.Points)
		if c.Answers == 0 {
			t.Errorf("%s has no coded answers: it would appear as an empty cloud", c.PartyID)
		}
	}
	if want := len(sp.Clouds) * len(sp.Dimensions); points != want {
		t.Errorf("points = %d, want %d (one per party per dimension)", points, want)
	}
	if got, want := len(sp.Springs), len(sp.Clouds)*(len(sp.Clouds)-1)/2; got != want {
		t.Errorf("springs = %d, want %d pairwise", got, want)
	}
}

// The variance shares are the projection's headline claim, so they are pinned
// against an independent implementation (numpy's SVD on the same matrix gave
// 50.0 % / 20.3 % / 12.3 %, total 82.6 %). A Jacobi bug that still produced
// three plausible numbers would pass a "sums to something" check and fail this.
func TestThreeAxesHoldTheMeasuredShare(t *testing.T) {
	sp := Project(data.Store)
	if math.Abs(sp.Explained-0.826) > 0.005 {
		t.Errorf("explained = %.4f, want 0.826 (numpy SVD cross-check)", sp.Explained)
	}
	want := []float64{0.500, 0.203, 0.123}
	sum := 0.0
	for i, a := range sp.Axes {
		if math.Abs(a.Variance-want[i]) > 0.005 {
			t.Errorf("axis %d variance = %.4f, want %.3f", i+1, a.Variance, want[i])
		}
		if i > 0 && a.Variance > sp.Axes[i-1].Variance {
			t.Errorf("axis %d is larger than axis %d: the order is not descending", i+1, i)
		}
		sum += a.Variance
	}
	if math.Abs(sum-sp.Explained) > 1e-9 {
		t.Errorf("axis shares sum to %.6f but Explained = %.6f", sum, sp.Explained)
	}
	if !sp.Standardised {
		t.Error("Standardised is false: the page would be claiming the raw scale")
	}
}

// This is the guard against the trap the raw data sets. Unstandardised, the first
// axis loads 0.65 on "Teknologisk kapasitet" purely because that dimension has
// three extreme statements and a variance of 1.75 against 0.18 for "Norsk
// handlefrihet", and the page would confidently name the master axis of Norwegian
// politics after a measurement artefact.
func TestNoSingleDimensionDominatesAnAxis(t *testing.T) {
	sp := Project(data.Store)
	for i, a := range sp.Axes {
		if a.Loadings[0].Dimension == "" {
			t.Fatalf("axis %d has no loadings", i+1)
		}
		if a.Loadings[0].Loading < 0 {
			t.Errorf("axis %d's heaviest loading is negative: the eigenvectors are not "+
				"sign-normalised, so an arbitrary flip renames the axis", i+1)
		}
	}

	// Only axis 1 is held to the "no single dimension" bar, and deliberately. It is
	// the axis the raw scale corrupts, because the first component chases the
	// largest variance. Later components are EXPECTED to concentrate once the
	// columns are standardised: a 0.56 loading there is a real correlation between
	// one dimension and that direction, not a scale artefact.
	if top := math.Abs(sp.Axes[0].Loadings[0].Loading); top >= 0.5 {
		t.Errorf("axis 1 loads %.3f on %q alone: the columns are not standardised, so the "+
			"axis is naming whichever dimension happens to hold the most extreme statements",
			top, sp.Axes[0].Loadings[0].Dimension)
	}
	// And the specific artefact: the raw-heaviest dimension must not be what axis 1
	// is named after. Unstandardised, "Teknologisk kapasitet" is exactly that.
	heaviestRaw := sp.Dimensions[0]
	for _, d := range sp.Dimensions {
		if d.Variance > heaviestRaw.Variance {
			heaviestRaw = d
		}
	}
	if sp.Axes[0].Loadings[0].Dimension == heaviestRaw.Name {
		t.Errorf("axis 1 is named after %q, which is also the raw-heaviest dimension "+
			"(variance %.2f): the projection is reporting scale, not substance",
			heaviestRaw.Name, heaviestRaw.Variance)
	}
	// And the raw spread between dimensions must still be reported, because it is
	// the reason the standardising is not optional.
	lo, hi := math.Inf(1), 0.0
	for _, d := range sp.Dimensions {
		if d.Variance < lo {
			lo = d.Variance
		}
		if d.Variance > hi {
			hi = d.Variance
		}
	}
	if lo <= 0 || hi/lo < 5 {
		t.Errorf("raw dimension variance ranges %.3f..%.3f (ratio %.1f): the scale problem "+
			"this projection exists to correct is no longer present, so the test's own "+
			"premise is stale", lo, hi, hi/lo)
	}
	t.Logf("raw variance ratio %.1fx is why the columns are standardised", hi/lo)
}

// The decomposition has to be exact, not decorative: a party's per-dimension
// contributions must add up to the centre it is drawn at, with the star centred on
// it. If the offsets did not sum to zero the cloud would be drawn somewhere the
// party is not.
func TestCloudOffsetsAreCentredOnTheParty(t *testing.T) {
	sp := Project(data.Store)
	for _, c := range sp.Clouds {
		for a := 0; a < Axes; a++ {
			sum := 0.0
			for _, p := range c.Points {
				sum += p.Offset[a]
			}
			if math.Abs(sum) > 1e-9 {
				t.Errorf("%s offsets on axis %d sum to %.3g, want 0: the cloud is not "+
					"centred on the party it belongs to", c.PartyID, a+1, sum)
			}
		}
		// Spread must equal the RMS length it claims to be.
		var sumSq float64
		for _, p := range c.Points {
			for a := 0; a < Axes; a++ {
				sumSq += p.Offset[a] * p.Offset[a]
			}
		}
		if want := math.Sqrt(sumSq / float64(len(c.Points))); math.Abs(c.Spread-want) > 1e-9 {
			t.Errorf("%s spread = %.6f, want %.6f from its own offsets", c.PartyID, c.Spread, want)
		}
	}
}

// Spread is easy to fake as a constant, so this pins that it is measured and states
// the honest consequence: the parties are all about equally spread. Cloud SIZE is
// therefore not the story; orientation is.
func TestSpreadIsMeasuredAndNarrow(t *testing.T) {
	sp := Project(data.Store)
	lo, hi := math.Inf(1), 0.0
	for _, c := range sp.Clouds {
		if c.Spread < lo {
			lo = c.Spread
		}
		if c.Spread > hi {
			hi = c.Spread
		}
	}
	if lo == hi {
		t.Fatalf("every spread is %.4f: the value is being set, not measured", lo)
	}
	if hi/lo > 2 {
		t.Errorf("spread ranges %.2f..%.2f: larger than the measured data supports", lo, hi)
	}
	t.Logf("spread ranges %.2f..%.2f across the parties (ratio %.2f): size alone "+
		"separates them weakly, so the shape has to carry the meaning", lo, hi, hi/lo)
}

// The springs are the second measurement in the picture and must come from real
// divisions. No pair agrees always, which is the finding the visual exists to show.
func TestSpringsComeFromRealDivisions(t *testing.T) {
	sp := Project(data.Store)
	busiest, maxAg := 0, 0.0
	for _, s := range sp.Springs {
		if s.Comparable <= 0 {
			t.Errorf("%s/%s spring has %d comparable divisions: it is not evidence", s.A, s.B, s.Comparable)
		}
		if s.Agreement < 0 || s.Agreement > 1 {
			t.Errorf("%s/%s agreement = %.3f, outside 0..1", s.A, s.B, s.Agreement)
		}
		if s.Comparable > busiest {
			busiest = s.Comparable
		}
		if s.Agreement > maxAg {
			maxAg = s.Agreement
		}
	}
	if busiest < 10000 {
		t.Errorf("busiest pair is only %d divisions: the vote index has shrunk", busiest)
	}
	if maxAg >= 0.999 {
		t.Errorf("some pair agrees %.3f of the time: either the data or the tie is wrong, "+
			"and 'there are no perfect allies' would stop being true", maxAg)
	}
	t.Logf("closest pair agrees %.1f%% over %d divisions", maxAg*100, busiest)
}

// A dimension with no coded statement for anybody would be a column of zeroes that
// silently took a share of the variance.
func TestEveryDimensionIsCovered(t *testing.T) {
	sp := Project(data.Store)
	for _, d := range sp.Dimensions {
		if d.Statements == 0 {
			t.Errorf("%q has no statements", d.Name)
		}
		if d.Coded == 0 {
			t.Errorf("%q has no coded party position: it is an empty column in the projection", d.Name)
		}
	}
}

// An eigenvector's sign is arbitrary, so the projection fixes it by making the
// heaviest component positive. The real corpus happens to yield positive-heavy
// vectors already, which means a mutation that disables the flip would pass every
// other test in this file — so it is tested here on a vector that needs flipping.
func TestAnAxisSignIsFixedNotArbitrary(t *testing.T) {
	in := []float64{-0.1, -0.7, 0.2}
	got := signNormalise(in)
	if got[1] < 0 {
		t.Errorf("heaviest component is still %.2f: an arbitrary flip would rename the axis", got[1])
	}
	if in[1] != -0.7 {
		t.Errorf("signNormalise mutated its input (%.2f), so a caller's vector changes under it", in[1])
	}
	// A vector that is already positive-heavy must come back unchanged.
	same := signNormalise([]float64{0.3, 0.9, -0.2})
	if same[1] != 0.9 {
		t.Errorf("signNormalise flipped an already-positive vector: %v", same)
	}
}
