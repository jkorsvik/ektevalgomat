package server

import (
	"strings"
	"testing"
)

func TestStatementKeepsAuditNotesBehindTechnicalDisclosure(t *testing.T) {
	body := bodyOf(t, newTestServer(), "/utsagn/1")
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
