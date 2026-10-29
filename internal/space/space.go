// Package space projects the party-by-dimension positions into three axes a
// reader can look at, and measures how much of the data the projection keeps.
//
// The one decision worth knowing before reading a picture built from this: the
// dimensions are standardised before the projection. They are not on a common
// scale. Across the nine parties "Teknologisk kapasitet" has a variance of 1.75
// from three statements while "Norsk handlefrihet" has 0.18 from seven, a factor
// of ten. A projection that used the raw values would make the first axis
// "technology capacity" and would be reporting which dimension happens to hold
// the most extreme statements, not how the parties divide. Standardising makes
// each dimension count once; the raw variance is still reported per dimension so
// the choice is visible rather than buried.
package space

import (
	"math"
	"sort"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// Axes is how many dimensions survive the projection. Three, because it is what
// a person can rotate in their head and what the page draws.
const Axes = 3

// Weight is one dimension's signed contribution to an axis.
type Weight struct {
	Dimension string
	Loading   float64
}

// Axis is one principal axis of the standardised space: how much of the total
// variation it carries, and how the dimensions load on it, heaviest first.
type Axis struct {
	Variance float64
	Loadings []Weight
}

// Point is one dimension's contribution to a party's place in the space, drawn
// from the party's centre. The Axes coordinates are offsets, so a point carries
// which dimension pulls a party where, not just how far.
type Point struct {
	Dimension string
	Offset    [Axes]float64
}

// Cloud is one party: where it sits, how widely its own dimensions scatter
// around that place, and the per-dimension offsets that make the shape.
type Cloud struct {
	PartyID string
	Centre  [Axes]float64
	// Spread is the RMS length of the party's dimension offsets. A party that
	// answers every dimension the same way collapses to a dot; one pulled in
	// opposite directions by different dimensions is large. It is measured from
	// the data, never chosen.
	Spread  float64
	Points  []Point
	Answers int
}

// Spring is how two parties actually voted together, which is a different
// question from how close their written positions are.
type Spring struct {
	A, B       string
	Agreement  float64 // 0..1 over the divisions both took a side in
	Comparable int     // divisions where both took a side; 0 never appears
}

// Dimension reports one dimension's shape before standardisation, so a reader
// can see whether a thin dimension is being given weight it has not earned.
type Dimension struct {
	Name       string
	Statements int
	Coded      int     // parties with at least one coded statement here
	Variance   float64 // raw spread across the parties, before standardising
}

// Space is the whole projection.
type Space struct {
	Axes       []Axis
	Clouds     []Cloud
	Springs    []Spring
	Dimensions []Dimension
	// Explained is the share of the standardised variance the three axes keep
	// together. The page prints it, because the rest is off-screen.
	Explained float64
	// Standardised is always true; it is a field so the page cannot claim the
	// raw scale without the projection disagreeing.
	Standardised bool
}

// Project builds the space from the loaded dataset and the committed vote index.
func Project(ds *data.Dataset) Space {
	return project(ds, data.Votes())
}

// cellPosition returns the position the cell reports: its newest dated row, ties
// broken by evidence id. It mirrors scoring.ComputeCell deliberately rather than
// importing it, because scoring imports data and this package must stay usable
// from both.
func cellPosition(rows []data.Evidence) (int, bool) {
	var best *data.Evidence
	for i := range rows {
		e := &rows[i]
		if e.Position == nil {
			continue
		}
		if best == nil || e.Date > best.Date || (e.Date == best.Date && e.ID > best.ID) {
			best = e
		}
	}
	if best == nil {
		return 0, false
	}
	return *best.Position, true
}

func project(ds *data.Dataset, votes []data.Vote) Space {
	// Dimensions in a stable order, taken from the statements themselves so a new
	// dimension appears without a code change.
	dimSet := map[string]bool{}
	stmtsByDim := map[string][]string{}
	for _, s := range ds.Statements {
		dimSet[s.Dimension] = true
		stmtsByDim[s.Dimension] = append(stmtsByDim[s.Dimension], s.ID)
	}
	var dimNames []string
	for d := range dimSet {
		dimNames = append(dimNames, d)
	}
	sort.Strings(dimNames)

	parties := ds.Parties
	nP, nD := len(parties), len(dimNames)
	dimIdx := map[string]int{}
	for j, d := range dimNames {
		dimIdx[d] = j
	}

	// Matrix of mean coded positions per (party, dimension), plus coverage.
	raw := make([][]float64, nP)
	have := make([][]bool, nP)
	answers := make([]int, nP)
	for i := range parties {
		raw[i] = make([]float64, nD)
		have[i] = make([]bool, nD)
	}

	for i, p := range parties {
		sum := make([]float64, nD)
		cnt := make([]int, nD)
		for _, s := range ds.Statements {
			if _, ok := cellPosition(ds.EvidenceForCell(s.ID, p.ID)); !ok {
				continue
			}
			pos, _ := cellPosition(ds.EvidenceForCell(s.ID, p.ID))
			j := dimIdx[s.Dimension]
			sum[j] += float64(pos)
			cnt[j]++
			answers[i]++
		}
		for j := 0; j < nD; j++ {
			if cnt[j] > 0 {
				raw[i][j] = sum[j] / float64(cnt[j])
				have[i][j] = true
			}
		}
	}

	// Fill a missing cell with the dimension's mean so one absent party does not
	// silently pull the axis toward zero, and record the raw variance before any
	// standardising, which is the number that shows the scale problem.
	dims := make([]Dimension, nD)
	for j, name := range dimNames {
		col := make([]float64, 0, nP)
		coded := 0
		for i := 0; i < nP; i++ {
			if have[i][j] {
				col = append(col, raw[i][j])
				coded++
			}
		}
		mean := 0.0
		for _, v := range col {
			mean += v
		}
		if len(col) > 0 {
			mean /= float64(len(col))
		}
		for i := 0; i < nP; i++ {
			if !have[i][j] {
				raw[i][j] = mean
			}
		}
		v := 0.0
		for i := 0; i < nP; i++ {
			v += (raw[i][j] - mean) * (raw[i][j] - mean)
		}
		if nP > 0 {
			v /= float64(nP)
		}
		dims[j] = Dimension{Name: name, Statements: len(stmtsByDim[name]), Coded: coded, Variance: v}
	}

	// Standardise each dimension: mean 0, unit spread, so every dimension counts
	// once regardless of how many statements or how extreme its statements are.
	z := make([][]float64, nP)
	for i := range z {
		z[i] = make([]float64, nD)
	}
	for j := 0; j < nD; j++ {
		mean, sd := 0.0, 0.0
		for i := 0; i < nP; i++ {
			mean += raw[i][j]
		}
		if nP > 0 {
			mean /= float64(nP)
		}
		for i := 0; i < nP; i++ {
			sd += (raw[i][j] - mean) * (raw[i][j] - mean)
		}
		sd = math.Sqrt(sd / float64(nP))
		for i := 0; i < nP; i++ {
			if sd > 0 {
				z[i][j] = (raw[i][j] - mean) / sd
			}
		}
	}

	// Correlation matrix over the dimensions, then its eigen-decomposition.
	corr := make([][]float64, nD)
	for j := range corr {
		corr[j] = make([]float64, nD)
	}
	for a := 0; a < nD; a++ {
		for b := 0; b < nD; b++ {
			s := 0.0
			for i := 0; i < nP; i++ {
				s += z[i][a] * z[i][b]
			}
			if nP > 0 {
				s /= float64(nP)
			}
			corr[a][b] = s
		}
	}
	vals, vecs := jacobiEigen(corr)

	total := 0.0
	for _, l := range vals {
		total += l
	}
	axes := make([]Axis, 0, Axes)
	// proj keeps each axis's vector in DIMENSION order. Axis.Loadings is sorted by
	// weight for display, so the two are not interchangeable and using the sorted
	// one for the projection would multiply the wrong dimension on every party.
	proj := make([][]float64, 0, Axes)
	for a := 0; a < Axes && a < nD; a++ {
		// An eigenvector's sign is arbitrary, and flipping it must not rename the
		// axis, so the heaviest component is made positive before labelling.
		vec := make([]float64, nD)
		for j := 0; j < nD; j++ {
			vec[j] = vecs[j][a]
		}
		vec = signNormalise(vec)
		proj = append(proj, vec)
		w := make([]Weight, nD)
		for j := 0; j < nD; j++ {
			w[j] = Weight{Dimension: dimNames[j], Loading: vec[j]}
		}
		sort.SliceStable(w, func(x, y int) bool { return math.Abs(w[x].Loading) > math.Abs(w[y].Loading) })
		v := 0.0
		if total > 0 {
			v = vals[a] / total
		}
		axes = append(axes, Axis{Variance: v, Loadings: w})
	}

	// Party centres and clouds: the party's row projected onto the axes, and the
	// per-dimension contributions that add up to it.
	clouds := make([]Cloud, nP)
	for i, p := range parties {
		contrib := make([][Axes]float64, nD)
		var centre [Axes]float64
		for j := 0; j < nD; j++ {
			for a := 0; a < len(proj); a++ {
				c := z[i][j] * proj[a][j]
				contrib[j][a] = c
				centre[a] += c
			}
		}
		// Centre the star on the party: an offset is its dimension's contribution
		// relative to the party's average pull. Order-free, because dimension order
		// is arbitrary and a trajectory through an invented order would invent a
		// shape too.
		var meanOff [Axes]float64
		for a := range meanOff {
			meanOff[a] = centre[a] / float64(nD)
		}
		pts := make([]Point, nD)
		var sumSq float64
		for j := 0; j < nD; j++ {
			for a := 0; a < len(proj); a++ {
				pts[j].Offset[a] = contrib[j][a] - meanOff[a]
				sumSq += pts[j].Offset[a] * pts[j].Offset[a]
			}
			pts[j].Dimension = dimNames[j]
		}
		clouds[i] = Cloud{
			PartyID: p.ID,
			Centre:  centre,
			Spread:  math.Sqrt(sumSq / float64(nD)),
			Points:  pts,
			Answers: answers[i],
		}
	}

	explained := 0.0
	for _, a := range axes {
		explained += a.Variance
	}

	return Space{
		Axes:         axes,
		Clouds:       clouds,
		Springs:      springs(parties, votes),
		Dimensions:   dims,
		Explained:    explained,
		Standardised: true,
	}
}

// signNormalise returns vec with its heaviest component made positive. An
// eigenvector is defined only up to sign, so without this an arbitrary flip would
// rename an axis and change what the page says about Norwegian politics.
func signNormalise(vec []float64) []float64 {
	heaviest, hi := 0.0, 0
	for j, v := range vec {
		if math.Abs(v) > heaviest {
			heaviest, hi = math.Abs(v), j
		}
	}
	if vec[hi] >= 0 {
		return vec
	}
	out := make([]float64, len(vec))
	for j, v := range vec {
		out[j] = -v
	}
	return out
}

// springs measures how often each pair of parties took the same side, over the
// divisions where both took a side at all. Unanimous divisions and parties
// recorded as absent or split contribute nothing, because neither is agreement.
func springs(parties []data.Party, votes []data.Vote) []Spring {
	type acc struct{ agree, n int }
	pairs := map[[2]int]*acc{}
	for _, v := range votes {
		for i := 0; i < len(parties); i++ {
			a := v.Parties[parties[i].ID]
			if a != "for" && a != "mot" {
				continue
			}
			for j := i + 1; j < len(parties); j++ {
				b := v.Parties[parties[j].ID]
				if b != "for" && b != "mot" {
					continue
				}
				k := [2]int{i, j}
				if pairs[k] == nil {
					pairs[k] = &acc{}
				}
				pairs[k].n++
				if a == b {
					pairs[k].agree++
				}
			}
		}
	}
	keys := make([][2]int, 0, len(pairs))
	for k := range pairs {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(x, y int) bool {
		if keys[x][0] != keys[y][0] {
			return keys[x][0] < keys[y][0]
		}
		return keys[x][1] < keys[y][1]
	})
	out := make([]Spring, 0, len(keys))
	for _, k := range keys {
		a := pairs[k]
		if a.n == 0 {
			continue
		}
		out = append(out, Spring{
			A:          parties[k[0]].ID,
			B:          parties[k[1]].ID,
			Agreement:  float64(a.agree) / float64(a.n),
			Comparable: a.n,
		})
	}
	return out
}

// jacobiEigen returns the eigenvalues of a symmetric matrix in descending order,
// with the matching eigenvectors as columns. Cyclic Jacobi is used because the
// matrices here are at most a few dozen wide and it needs no dependency; the
// test cross-checks the variance shares against an independent implementation.
func jacobiEigen(a [][]float64) ([]float64, [][]float64) {
	n := len(a)
	m := make([][]float64, n)
	v := make([][]float64, n)
	for i := 0; i < n; i++ {
		m[i] = append([]float64(nil), a[i]...)
		v[i] = make([]float64, n)
		v[i][i] = 1
	}
	for sweep := 0; sweep < 100; sweep++ {
		off := 0.0
		for p := 0; p < n; p++ {
			for q := p + 1; q < n; q++ {
				off += m[p][q] * m[p][q]
			}
		}
		if off < 1e-20 {
			break
		}
		for p := 0; p < n; p++ {
			for q := p + 1; q < n; q++ {
				if math.Abs(m[p][q]) < 1e-18 {
					continue
				}
				theta := (m[q][q] - m[p][p]) / (2 * m[p][q])
				t := 1.0
				if theta != 0 {
					t = math.Copysign(1, theta) / (math.Abs(theta) + math.Sqrt(theta*theta+1))
				}
				c := 1 / math.Sqrt(t*t+1)
				s := t * c
				for k := 0; k < n; k++ {
					mkp, mkq := m[k][p], m[k][q]
					m[k][p] = c*mkp - s*mkq
					m[k][q] = s*mkp + c*mkq
				}
				for k := 0; k < n; k++ {
					mpk, mqk := m[p][k], m[q][k]
					m[p][k] = c*mpk - s*mqk
					m[q][k] = s*mpk + c*mqk
				}
				for k := 0; k < n; k++ {
					vkp, vkq := v[k][p], v[k][q]
					v[k][p] = c*vkp - s*vkq
					v[k][q] = s*vkp + c*vkq
				}
			}
		}
	}

	vals := make([]float64, n)
	idx := make([]int, n)
	for i := 0; i < n; i++ {
		vals[i] = m[i][i]
		idx[i] = i
	}
	sort.Slice(idx, func(x, y int) bool { return vals[idx[x]] > vals[idx[y]] })
	outVals := make([]float64, n)
	outVecs := make([][]float64, n)
	for i := range outVecs {
		outVecs[i] = make([]float64, n)
	}
	for newPos, old := range idx {
		outVals[newPos] = vals[old]
		for row := 0; row < n; row++ {
			outVecs[row][newPos] = v[row][old]
		}
	}
	return outVals, outVecs
}
