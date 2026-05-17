package graphql

import (
	"bytes"
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
	if h.schema == nil {
		t.Error("NewHandler() schema is nil")
	}
}

func TestServeGraphiQL(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/graphiql", nil)
	w := httptest.NewRecorder()

	h.ServeGraphiQL(w, req)

	resp := w.Result()
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("ServeGraphiQL() Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(w.Body.String(), "<!DOCTYPE html>") {
		t.Error("ServeGraphiQL() did not return HTML")
	}
}

func TestServeGraphiQL_DarkTheme(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/graphiql?theme=dark", nil)
	w := httptest.NewRecorder()
	h.ServeGraphiQL(w, req)
	if !strings.Contains(w.Body.String(), "dark") {
		t.Error("ServeGraphiQL() with theme=dark did not include dark theme")
	}
}

func TestServeGraphiQL_LightTheme(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/graphiql?theme=light", nil)
	w := httptest.NewRecorder()
	h.ServeGraphiQL(w, req)
	if !strings.Contains(w.Body.String(), "light") {
		t.Error("ServeGraphiQL() with theme=light did not include light theme")
	}
}

func TestServeGraphQL_GET_MethodNotAllowed(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/graphql", nil)
	w := httptest.NewRecorder()

	h.ServeGraphQL(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("ServeGraphQL(GET) status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestServeGraphQL_POST_InvalidJSON(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader("not-json"))
	w := httptest.NewRecorder()

	h.ServeGraphQL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ServeGraphQL(invalid JSON) status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestServeGraphQL_POST_ValidRequest(t *testing.T) {
	h := NewHandler()

	gqlReq := GraphQLRequest{
		Query: "{ profiles { id } }",
	}
	body, _ := json.Marshal(gqlReq)
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeGraphQL(w, req)

	resp := w.Result()
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("ServeGraphQL() Content-Type = %q, want application/json", ct)
	}

	var gqlResp GraphQLResponse
	if err := json.NewDecoder(w.Body).Decode(&gqlResp); err != nil {
		t.Fatalf("ServeGraphQL() response is not valid JSON: %v", err)
	}
}

func TestBuildSchema(t *testing.T) {
	s := buildSchema()
	if s == nil {
		t.Fatal("buildSchema() returned nil")
	}
	if s.Query == nil {
		t.Error("schema.Query is nil")
	}
	if s.Mutation == nil {
		t.Error("schema.Mutation is nil")
	}
	if _, ok := s.Query.Fields["profiles"]; !ok {
		t.Error("schema.Query missing 'profiles' field")
	}
	if _, ok := s.Mutation.Fields["createProfile"]; !ok {
		t.Error("schema.Mutation missing 'createProfile' field")
	}
}

func TestGetThemePreference(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"/graphiql?theme=dark", "dark"},
		{"/graphiql?theme=light", "light"},
		{"/graphiql?theme=invalid", "dark"},
		{"/graphiql", "dark"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.url, nil)
		got := getThemePreference(req)
		if got != tt.want {
			t.Errorf("getThemePreference(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestGetThemePreference_Cookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "theme", Value: "light"})
	if got := getThemePreference(req); got != "light" {
		t.Errorf("getThemePreference(cookie=light) = %q, want light", got)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{Name: "theme", Value: "bad"})
	if got := getThemePreference(req2); got != "dark" {
		t.Errorf("getThemePreference(cookie=bad) = %q, want dark", got)
	}
}

func TestGenerateGraphiQLHTML_Dark(t *testing.T) {
	h := NewHandler()
	html := h.generateGraphiQLHTML("dark")
	if !strings.Contains(html, "graphiql-dark") {
		t.Error("generateGraphiQLHTML(dark) should contain graphiql-dark class")
	}
}

func TestGenerateGraphiQLHTML_Light(t *testing.T) {
	h := NewHandler()
	html := h.generateGraphiQLHTML("light")
	if !strings.Contains(html, "graphiql-light") {
		t.Error("generateGraphiQLHTML(light) should contain graphiql-light class")
	}
}
