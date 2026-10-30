package server

import (
	"encoding/json"
	"html/template"
	"math"
	"strconv"
	"strings"

	"github.com/jkorsvik/ektevalgomat/internal/data"
	"github.com/jkorsvik/ektevalgomat/internal/space"
)

// The space view: the twelve dimensions projected onto three axes, so a party is a
// volume instead of a dot on a line.
//
// Why this page exists beside /kart: the map answers one question at a time, one
// number per cell. That makes every cell look equally certain and hides that the
// dimensions are not on a common scale at all — "Teknologisk kapasitet" varies 9.6
// times as much as "Norsk handlefrihet" across the nine parties. Projected raw, the
// first axis would be that dimension and the page would name the master axis of
// Norwegian politics after a measurement artefact.
//
// So the projection standardises, and the page says so, prints the share of the
// variation the three axes keep, and prints the spread that each party's cloud is
// drawn from. A reader must be able to see the 17 % that is off-screen.

type spaceDimensionView struct {
	Name       string
	Statements int
	Coded      int
	Variance   string // already formatted, so the table cannot print a dot
}

type spacePartyView struct {
	ID      string
	Name    string
	Short   string
	Centre  string // formatted coordinates, for the accessible table
	Spread  string
	Answers int
}

// spaceWeightView is a loading already formatted for a Norwegian reader: comma
// decimals and a typographic minus. Formatting it in the template turned the
// separators into their own flex items and printed "0.37" on a Norwegian page.
type spaceWeightView struct {
	Dimension string
	Loading   string
}

type spaceAxisView struct {
	Number  string
	Percent string
	TopText string            // heaviest three, already joined for one text flow
	Rest    []spaceWeightView // the remainder, kept available but not shouted
}

type spaceView struct {
	Meta         pageMeta
	SnapshotDate string
	Explained    string // share kept by three axes, one decimal
	Rest         string // 100 minus Explained, so the page cannot round it away
	Axes         []spaceAxisView
	Parties      []spacePartyView
	Dimensions   []spaceDimensionView
	Springs      int
	Busiest      int
	Closest      string
	ClosestPair  string
	// SpreadLo and SpreadHi are the smallest and largest measured spread, so the
	// page can say how much of the shape is size. Measured, never typed.
	SpreadLo    string
	SpreadHi    string
	PayloadJSON template.JS
}

// decimal formats a share of 1 as a percentage with one decimal and a Norwegian
// comma, so the page never mixes "82.6" and "82,6".
func decimal(share float64) string {
	s := strconv.FormatFloat(share*100, 'f', 1, 64)
	return strings.Replace(s, ".", ",", -1)
}

// comma formats with a Norwegian decimal comma and a typographic minus sign.
func comma(v float64, dp int) string {
	s := strconv.FormatFloat(v, 'f', dp, 64)
	s = strings.Replace(s, ".", ",", -1)
	return strings.Replace(s, "-", "\u2212", 1)
}

func signedDecimal(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	s = strings.Replace(s, ".", ",", -1)
	return s
}

// spacePayload is the geometry the canvas draws, with dimension names replaced by
// their index. The names appear once in Dims instead of once per point, which
// keeps the inlined payload about a tenth of its naive size.
type spacePayload struct {
	Explained float64     `json:"explained"`
	Axes      []axisOut   `json:"axes"`
	Dims      []dimOut    `json:"dims"`
	Clouds    []cloudOut  `json:"clouds"`
	Springs   []springOut `json:"springs"`
}

type axisOut struct {
	Variance float64   `json:"variance"`
	Loadings []float64 `json:"loadings"` // aligned with Dims
}

type dimOut struct {
	Name       string  `json:"name"`
	Statements int     `json:"statements"`
	Coded      int     `json:"coded"`
	Variance   float64 `json:"variance"`
}

type cloudOut struct {
	ID      string                `json:"id"`
	Short   string                `json:"short"`   // the name the reader sees, from Party.Short
	Answers int                   `json:"answers"` // coded cells the centre rests on
	Centre  [space.Axes]float64   `json:"centre"`
	Spread  float64               `json:"spread"`
	Points  [][space.Axes]float64 `json:"points"` // aligned with Dims
}

type springOut struct {
	A          string  `json:"a"`
	B          string  `json:"b"`
	Agreement  float64 `json:"agreement"`
	Comparable int     `json:"comparable"`
}

func buildSpaceView(ds *data.Dataset) spaceView {
	sp := space.Project(ds)

	dimIndex := map[string]int{}
	for j, d := range sp.Dimensions {
		dimIndex[d.Name] = j
	}

	axes := make([]spaceAxisView, 0, len(sp.Axes))
	axesOut := make([]axisOut, 0, len(sp.Axes))
	for i, a := range sp.Axes {
		top := a.Loadings
		if len(top) > 3 {
			top = top[:3]
		}
		var parts []string
		for _, w := range top {
			parts = append(parts, w.Dimension+" "+comma(w.Loading, 2))
		}
		rest := make([]spaceWeightView, 0, len(a.Loadings)-len(top))
		for _, w := range a.Loadings[len(top):] {
			rest = append(rest, spaceWeightView{Dimension: w.Dimension, Loading: comma(w.Loading, 2)})
		}
		axes = append(axes, spaceAxisView{
			Number:  strconv.Itoa(i + 1),
			Percent: decimal(a.Variance),
			TopText: strings.Join(parts, " \u00b7 "),
			Rest:    rest,
		})
		loadings := make([]float64, len(sp.Dimensions))
		for _, w := range a.Loadings {
			loadings[dimIndex[w.Dimension]] = w.Loading
		}
		axesOut = append(axesOut, axisOut{Variance: a.Variance, Loadings: loadings})
	}

	dimsView := make([]spaceDimensionView, 0, len(sp.Dimensions))
	for _, d := range sp.Dimensions {
		dimsView = append(dimsView, spaceDimensionView{
			Name: d.Name, Statements: d.Statements, Coded: d.Coded, Variance: comma(d.Variance, 2),
		})
	}
	dimsOut := make([]dimOut, 0, len(sp.Dimensions))
	for _, d := range sp.Dimensions {
		dimsOut = append(dimsOut, dimOut{
			Name: d.Name, Statements: d.Statements, Coded: d.Coded, Variance: d.Variance,
		})
	}

	byID := map[string]data.Party{}
	for _, p := range ds.Parties {
		byID[p.ID] = p
	}

	spreadLo, spreadHi := math.Inf(1), 0.0
	parties := make([]spacePartyView, 0, len(sp.Clouds))
	cloudsOut := make([]cloudOut, 0, len(sp.Clouds))
	busiest := 0
	var closest, closestPair string
	maxAg := -1.0
	for _, c := range sp.Clouds {
		p := byID[c.PartyID]
		pts := make([][space.Axes]float64, len(c.Points))
		for j, pt := range c.Points {
			pts[j] = pt.Offset
		}
		short := c.PartyID
		if p.Short != "" {
			short = p.Short
		}
		cloudsOut = append(cloudsOut, cloudOut{
			ID: c.PartyID, Short: short, Answers: c.Answers,
			Centre: c.Centre, Spread: c.Spread, Points: pts,
		})
		if c.Spread < spreadLo {
			spreadLo = c.Spread
		}
		if c.Spread > spreadHi {
			spreadHi = c.Spread
		}
		parties = append(parties, spacePartyView{
			ID:      c.PartyID,
			Name:    p.Name,
			Short:   p.Short,
			Centre:  signedDecimal(c.Centre[0]) + " · " + signedDecimal(c.Centre[1]) + " · " + signedDecimal(c.Centre[2]),
			Spread:  signedDecimal(c.Spread),
			Answers: c.Answers,
		})
	}
	springsOut := make([]springOut, 0, len(sp.Springs))
	for _, s := range sp.Springs {
		springsOut = append(springsOut, springOut{
			A: s.A, B: s.B, Agreement: s.Agreement, Comparable: s.Comparable,
		})
		if s.Comparable > busiest {
			busiest = s.Comparable
		}
		if s.Agreement > maxAg {
			maxAg = s.Agreement
			closest = decimal(s.Agreement) + " %"
			closestPair = strings.ToUpper(s.A) + "–" + strings.ToUpper(s.B)
		}
	}

	payload := spacePayload{
		Explained: sp.Explained,
		Axes:      axesOut,
		Dims:      dimsOut,
		Clouds:    cloudsOut,
		Springs:   springsOut,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		// The payload is built from committed data with no interface values, so a
		// failure here is a build defect, not a runtime condition.
		panic("server: marshaling space payload: " + err.Error())
	}

	return spaceView{
		Meta: meta("Rommet · Ekte valgomat",
			"Tolv politiske dimensjoner projisert ned til tre akser, så hvert parti blir et volum og ikke et punkt. "+
				"Siden viser hvor mye av variasjonen aksene beholder, og hva som faller utenfor.", "/rommet"),
		SnapshotDate: SnapshotDate,
		Explained:    decimal(sp.Explained),
		Rest:         decimal(1 - sp.Explained),
		Axes:         axes,
		Parties:      parties,
		Dimensions:   dimsView,
		Springs:      len(sp.Springs),
		Busiest:      busiest,
		Closest:      closest,
		ClosestPair:  closestPair,
		SpreadLo:     signedDecimal(spreadLo),
		SpreadHi:     signedDecimal(spreadHi),
		PayloadJSON:  template.JS(b), //nolint:gosec // our own generated JSON, never user input
	}
}
