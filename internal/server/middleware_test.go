package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityHeadersAllowOnlySameOriginForms(t *testing.T) {
	handler := withSecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/utforsk", nil))
	csp := recorder.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "form-action 'self'") {
		t.Errorf("CSP blocks the read-only statement finder: %s", csp)
	}
	if strings.Contains(csp, "form-action 'none'") || strings.Contains(csp, "form-action *") {
		t.Errorf("form-action must be same-origin only: %s", csp)
	}
}
