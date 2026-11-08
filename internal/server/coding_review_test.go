package server

import (
	"strings"
	"testing"

	"github.com/jkorsvik/ektevalgomat/internal/data"
)

func TestMetodePublishesTheWholeCodingReview(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/metode")
	review := data.CodingReviewData()
	for _, want := range []string{
		`id="kodingskontroll"`,
		"18 av 517 celler",
		"15 tilfeller",
		"8 rettelser",
		"ikke uavhengig menneskelig kontroll",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/metode missing %q", want)
		}
	}
	if got := strings.Count(body, "Begrunnelse og kilder for packet-"); got != len(review.Packets) {
		t.Errorf("/metode shows %d audit rationales, want %d", got, len(review.Packets))
	}
	if got := strings.Count(body, " · rettet"); got != review.Meta.ChangedPositions {
		t.Errorf("/metode shows %d corrections, want %d", got, review.Meta.ChangedPositions)
	}
}
