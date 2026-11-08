package server

import (
	"strings"
	"testing"
)

func TestResultPagePutsDecisionInsightsBeforeTheRanking(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/resultat")
	insights := strings.Index(body, `id="result-insights-root"`)
	ranking := strings.Index(body, `id="quiz-root"`)
	analysis := strings.Index(body, `id="analyse-root"`)
	if insights < 0 || ranking < 0 || analysis < 0 {
		t.Fatalf("missing result mounts: insights=%d ranking=%d analysis=%d", insights, ranking, analysis)
	}
	if !(insights < ranking && ranking < analysis) {
		t.Errorf("wrong result order: insights=%d ranking=%d analysis=%d", insights, ranking, analysis)
	}
}
