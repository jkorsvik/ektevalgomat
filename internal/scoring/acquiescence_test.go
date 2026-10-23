package scoring

import (
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// maxAcquiescenceSpread is a ratchet, not a target. A user who gives the
// SAME answer to every statement ("sterkt enig" everywhere, or "sterkt
// uenig" everywhere) has taken no real position, so no party should
// match them much better than another. The spread is the gap between the
// best and worst party's average saksmatch for such a user.
//
// Measured 2026-09-27 at 41 points in both directions: all-enig gave
// R 72 % / Ap 31 %, all-uenig gave Ap 68 % / R 27 %. Cause: most
// statements come from failed opposition forslag, so "enig" usually
// means "side with the proposer". A right-leaning tester found the
// symptom (he matched the parties he disagrees with most); another
// valgomat showed the same flaw when every answer was FOR.
//
// Fixed 2026-09-27 by flipping 13 statements to the opposite direction
// (utsagn-04, 07, 08, 12, 18, 19, 20, 21, 24, 25, 31, 37, 44): text
// rewritten to the position a real majority party actually argued in
// the source innstilling (never invented), every party position on
// that statement negated to match. Re-measured the same day at 14
// points in both directions: all-enig now gives Sp best / Ap worst,
// all-uenig gives Ap best / Sp worst — a much smaller, and no longer
// systematically left-vs-right, gap. See commit a9887cf's message for the
// candidate-selection method and the statements deliberately left
// unflipped (no natural positive phrasing backed by a real merknad).
//
// Lower this number whenever content changes reduce the spread. Never
// raise it to make a change pass: that would re-admit the bias this test
// exists to catch. Goal: below 15 — reached.
//
// Lowered to 13 on 2026-09-28: the verification pass coded 9 more cells
// on 47 and 51 (W4) and 2 on 52/53 (W3), and the spread fell to 13 in
// both directions.
//
// Lowered to 12 on 2026-10-02: utsagn-57 (skatt på store arv, agreed by
// the left) and utsagn-58 (utflytterskatten lempes, agreed by the right)
// were added as a direction-balanced pair (57 at 8/9, 58 at 9/9).
const maxAcquiescenceSpread = 12

func TestAcquiescenceSpreadDoesNotGrow(t *testing.T) {
	ds := data.Store
	if len(ds.Parties) < 2 || len(ds.Statements) == 0 {
		t.Fatalf("dataset too small to measure: %d parties, %d statements", len(ds.Parties), len(ds.Statements))
	}
	for _, answer := range []int{2, -2} {
		spread, best, worst := AcquiescenceSpread(ds, answer)
		t.Logf("same answer %+d on all %d statements: spread %d points (best %s, worst %s)", answer, len(ds.Statements), spread, best, worst)
		if spread > maxAcquiescenceSpread {
			t.Errorf("answer %+d everywhere: spread %d > ratchet %d (best %s, worst %s). New content made the quiz reward always-agree/always-disagree more; add or rephrase statements in the opposite direction instead of raising the ratchet", answer, spread, maxAcquiescenceSpread, best, worst)
		}
	}
}
