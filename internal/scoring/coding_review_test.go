package scoring

import (
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

// The audit is historical. A newer evidence row may legitimately change a live
// cell without invalidating the old review. Only the corrections themselves must
// remain present in the exact source rows the reviewers saw.
func TestCodingReviewCorrectionsRemainInTheirSourceRows(t *testing.T) {
	review := data.CodingReviewData()
	byID := map[string]data.Evidence{}
	for _, row := range data.Store.Evidence {
		byID[row.ID] = row
	}
	changed := 0
	for _, packet := range review.Packets {
		if packet.Resolution != "changed" {
			continue
		}
		changed++
		if packet.FinalPosition == nil || len(packet.ReviewedEvidence) == 0 {
			t.Errorf("%s: correction lacks final position or reviewed evidence", packet.PacketID)
			continue
		}
		newest := packet.ReviewedEvidence[0]
		for _, row := range packet.ReviewedEvidence[1:] {
			if row.Date > newest.Date || (row.Date == newest.Date && row.ID > newest.ID) {
				newest = row
			}
		}
		current, ok := byID[newest.ID]
		if !ok {
			t.Errorf("%s: corrected source row %s no longer exists", packet.PacketID, newest.ID)
			continue
		}
		if current.Position == nil || *current.Position != *packet.FinalPosition {
			t.Errorf("%s: source row %s has %v, audit correction has %d",
				packet.PacketID, newest.ID, current.Position, *packet.FinalPosition)
		}
	}
	if changed != review.Meta.ChangedPositions || changed == 0 {
		t.Fatalf("checked %d corrections, meta says %d", changed, review.Meta.ChangedPositions)
	}
	t.Logf("checked %d corrections against their immutable source rows; %d other packets left historical", changed, len(review.Packets)-changed)
}
