package scoring

import "github.com/jkorsvik/ektevalgomat/internal/data"

// AcquiescenceSpread is the gap in average saksmatch between the best and
// worst matching party for a user who gives the same answer to every
// statement. Such a user has taken no position, so the gap measures bias
// in the statement set, not in the user. The ratchet test in
// acquiescence_test.go guards it, and /metode shows it, from this one
// function.
func AcquiescenceSpread(ds *data.Dataset, answer int) (spread int, best, worst string) {
	hi, lo := -1, 101
	for _, p := range ds.Parties {
		sum, n := 0, 0
		for _, s := range ds.Statements {
			if pct, ok := ComputeCell(ds, s.ID, p.ID).SaksmatchPercent(answer); ok {
				sum += pct
				n++
			}
		}
		if n == 0 {
			continue
		}
		avg := sum / n
		if avg > hi {
			hi, best = avg, p.Short
		}
		if avg < lo {
			lo, worst = avg, p.Short
		}
	}
	return hi - lo, best, worst
}

// MaxAcquiescenceSpread is the larger spread of the two same-answer
// extremes (+2 everywhere, -2 everywhere).
func MaxAcquiescenceSpread(ds *data.Dataset) int {
	a, _, _ := AcquiescenceSpread(ds, 2)
	b, _, _ := AcquiescenceSpread(ds, -2)
	return max(a, b)
}
