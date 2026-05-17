package swagger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewHandler(t *testing.T) {
	h := NewHandler()
	if h == nil {
		t.Fatal("NewHandler() returned nil")
	}
	if h.spec == nil {
		t.Error("NewHandler() spec is nil")
	}
}

func TestServeSpec(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	w := httptest.NewRecorder()

	h.ServeSpec(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ServeSpec() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("ServeSpec() Content-Type = %q, want application/json", ct)
	}

	var spec OpenAPISpec
	if err := json.NewDecoder(resp.Body).Decode(&spec); err != nil {
		t.Fatalf("ServeSpec() response is not valid JSON: %v", err)
	}

	if spec.OpenAPI != "3.0.0" {
		t.Errorf("spec.OpenAPI = %q, want 3.0.0", spec.OpenAPI)
	}
	if spec.Info.Title == "" {
		t.Error("spec.Info.Title is empty")
	}
	if len(spec.Paths) == 0 {
		t.Error("spec.Paths is empty")
	}
}

func TestServeUI(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/swagger-ui", nil)
	w := httptest.NewRecorder()

	h.ServeUI(w, req)

	resp := w.Result()
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("ServeUI() Content-Type = %q, want text/html", ct)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("ServeUI() response does not contain DOCTYPE declaration")
	}
}

func TestServeUI_ThemeQuery(t *testing.T) {
	h := NewHandler()

	for _, theme := range []string{"dark", "light"} {
		req := httptest.NewRequest(http.MethodGet, "/swagger-ui?theme="+theme, nil)
		w := httptest.NewRecorder()
		h.ServeUI(w, req)
		body := w.Body.String()
		if !strings.Contains(body, theme) {
			t.Errorf("ServeUI() with theme=%q: response does not mention theme", theme)
		}
	}
}

func TestGetThemePreference_QueryParam(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"?theme=dark", "dark"},
		{"?theme=light", "light"},
		{"?theme=invalid", "dark"},
		{"", "dark"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)
		got := getThemePreference(req)
		if got != tt.want {
			t.Errorf("getThemePreference with query %q = %q, want %q", tt.query, got, tt.want)
		}
	}
}

func TestGetThemePreference_Cookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "theme", Value: "light"})
	if got := getThemePreference(req); got != "light" {
		t.Errorf("getThemePreference with cookie light = %q, want light", got)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{Name: "theme", Value: "dark"})
	if got := getThemePreference(req2); got != "dark" {
		t.Errorf("getThemePreference with cookie dark = %q, want dark", got)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.AddCookie(&http.Cookie{Name: "theme", Value: "invalid"})
	if got := getThemePreference(req3); got != "dark" {
		t.Errorf("getThemePreference with invalid cookie = %q, want dark (default)", got)
	}
}

func TestGetThemePreference_QueryOverridesCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?theme=dark", nil)
	req.AddCookie(&http.Cookie{Name: "theme", Value: "light"})
	if got := getThemePreference(req); got != "dark" {
		t.Errorf("query param should override cookie: got %q, want dark", got)
	}
}

func TestRenderJSON(t *testing.T) {
	h := NewHandler()
	w := httptest.NewRecorder()
	h.renderJSON(w, http.StatusCreated, map[string]string{"key": "value"})

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("renderJSON status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	var got map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("renderJSON body is not valid JSON: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("renderJSON body[key] = %q, want value", got["key"])
	}
}

func TestRenderError(t *testing.T) {
	h := NewHandler()
	w := httptest.NewRecorder()
	h.renderError(w, http.StatusBadRequest, "something went wrong")

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("renderError status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	var got map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("renderError body is not valid JSON: %v", err)
	}
	if got["error"] != "something went wrong" {
		t.Errorf("renderError body[error] = %q, want %q", got["error"], "something went wrong")
	}
}

func TestGenerateOpenAPISpec(t *testing.T) {
	spec := generateOpenAPISpec()
	if spec == nil {
		t.Fatal("generateOpenAPISpec() returned nil")
	}
	if spec.OpenAPI == "" {
		t.Error("spec.OpenAPI is empty")
	}
	if len(spec.Paths) == 0 {
		t.Error("spec.Paths is empty")
	}
	if _, ok := spec.Components.SecuritySchemes["bearerAuth"]; !ok {
		t.Error("spec.Components.SecuritySchemes missing bearerAuth")
	}
}

func TestGeneratePaths(t *testing.T) {
	paths := generatePaths()
	requiredPaths := []string{"/healthz", "/profiles", "/profiles/{slug}", "/links"}
	for _, path := range requiredPaths {
		if _, ok := paths[path]; !ok {
			t.Errorf("generatePaths() missing path %q", path)
		}
	}
}

func TestOpenAPISpecSerialization(t *testing.T) {
	spec := generateOpenAPISpec()
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("json.Marshal(spec) error: %v", err)
	}
	if len(data) == 0 {
		t.Error("json.Marshal(spec) produced empty output")
	}

	var decoded OpenAPISpec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal round-trip error: %v", err)
	}
	if decoded.OpenAPI != spec.OpenAPI {
		t.Errorf("round-trip OpenAPI = %q, want %q", decoded.OpenAPI, spec.OpenAPI)
	}
}
