package server

import (
	"strings"
	"testing"
)

func TestStatementKeepsAuditNotesBehindTechnicalDisclosure(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/utsagn/1")
	if !strings.Contains(body, "Hva betyr dette i praksis?") {
		t.Error("/utsagn/1 does not render PracticalStakes in the hero")
	}
	party := strings.Index(body, `id="party-ap"`)
	internal := strings.Index(body, "Batchen")
	if party < 0 || internal < 0 {
		t.Fatalf("missing fixture: party=%d internal-note=%d", party, internal)
	}
	segment := body[party:internal]
	if !strings.Contains(segment, "Dette er en redaksjonell koding") {
		t.Error("the public interpretation is missing before the audit note")
	}
	if !strings.Contains(segment, "Teknisk kodebegrunnelse") {
		t.Error("the internal context note is not behind the technical disclosure")
	}
}

func TestPartyPageShowsOrdOgHandlingSummary(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/parti/ap")
	for _, want := range []string{
		"Ord og handling på Stortinget",
		"konsistente",
		"Vis alle",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/parti/ap missing %q", want)
		}
	}
}
