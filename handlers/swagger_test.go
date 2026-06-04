package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSwaggerUI_ReturnsHTML(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	w := httptest.NewRecorder()

	SwaggerUI(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("expected text/html Content-Type, got %q", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "SwaggerUIBundle") {
		t.Fatalf("expected swagger-ui shell in body, got %q", body)
	}
	if !strings.Contains(body, "/openapi.yaml") {
		t.Fatalf("expected the shell to fetch /openapi.yaml, got %q", body)
	}
}

func TestOpenAPISpec_ReturnsEmbeddedYAML(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	w := httptest.NewRecorder()

	OpenAPISpec(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "application/yaml" {
		t.Fatalf("expected application/yaml, got %q", got)
	}
	body := w.Body.String()
	if !strings.HasPrefix(body, "openapi:") {
		t.Fatalf("expected body to start with 'openapi:', got %.40q...", body)
	}
	if !strings.Contains(body, "/subscriptions") {
		t.Fatalf("expected the spec to describe /subscriptions, got %q", body)
	}
}
