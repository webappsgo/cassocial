package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

func newTestImportExportHandler(t *testing.T) (*ImportExportHandler, *store.DB) {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	return NewImportExportHandler(&config.Config{}, db), db
}

// TestNewImportExportHandler verifies the constructor.
func TestNewImportExportHandler(t *testing.T) {
	h, _ := newTestImportExportHandler(t)
	if h == nil {
		t.Fatal("NewImportExportHandler returned nil")
	}
}

// TestHandleImport covers: wrong method, bad JSON, missing auth, unsupported source,
// and each supported import source.
func TestHandleImport(t *testing.T) {
	h, db := newTestImportExportHandler(t)
	userID := createTestUser(t, db, "importuser", "importuser@example.com")

	t.Run("wrong method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/import", nil)
		rr := httptest.NewRecorder()
		h.HandleImport(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("bad JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/import", bytes.NewBufferString("{bad"))
		rr := httptest.NewRecorder()
		h.HandleImport(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing auth returns 401", func(t *testing.T) {
		body, _ := json.Marshal(ImportRequest{Source: "json", Data: json.RawMessage(`{}`)})
		req := httptest.NewRequest(http.MethodPost, "/api/import", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		h.HandleImport(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("unsupported source returns 400", func(t *testing.T) {
		body, _ := json.Marshal(ImportRequest{Source: "unknown", Data: json.RawMessage(`{}`)})
		req := httptest.NewRequest(http.MethodPost, "/api/import", bytes.NewReader(body))
		req = withUserID(req, userID)
		rr := httptest.NewRecorder()
		h.HandleImport(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	sources := []struct {
		source string
		data   json.RawMessage
	}{
		{"linktree", json.RawMessage(`{}`)},
		{"linkstack", json.RawMessage(`{}`)},
		{"carrd", json.RawMessage(`{}`)},
		{"aboutme", json.RawMessage(`{}`)},
		{"csv", json.RawMessage(`{}`)},
		{"json", json.RawMessage(`{"profile":{"title":"T","description":"D"},"links":[]}`)},
	}

	for _, s := range sources {
		s := s
		t.Run("source="+s.source+" returns 200", func(t *testing.T) {
			body, _ := json.Marshal(ImportRequest{Source: s.source, Data: s.data})
			req := httptest.NewRequest(http.MethodPost, "/api/import", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = withUserID(req, userID)
			rr := httptest.NewRecorder()
			h.HandleImport(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("source=%s got %d, want %d; body: %s", s.source, rr.Code, http.StatusOK, rr.Body.String())
			}
			var resp map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("source=%s: response not valid JSON: %v", s.source, err)
			}
			if resp["status"] != "success" {
				t.Errorf("source=%s status = %q, want success", s.source, resp["status"])
			}
		})
	}
}

// TestImportFromJSON verifies link count is returned correctly.
func TestImportFromJSON(t *testing.T) {
	h, _ := newTestImportExportHandler(t)

	t.Run("valid JSON with links counts them", func(t *testing.T) {
		data := json.RawMessage(`{
			"profile":{"title":"T","description":"D"},
			"links":[
				{"service":"github","url":"https://github.com/u","title":"GitHub"},
				{"service":"twitter","url":"https://twitter.com/u","title":"Twitter"}
			]
		}`)
		n, err := h.importFromJSON("user-1", data)
		if err != nil {
			t.Fatalf("importFromJSON returned error: %v", err)
		}
		if n != 2 {
			t.Errorf("imported = %d, want 2", n)
		}
	})

	t.Run("empty links returns 0", func(t *testing.T) {
		data := json.RawMessage(`{"profile":{},"links":[]}`)
		n, err := h.importFromJSON("user-1", data)
		if err != nil {
			t.Fatalf("importFromJSON returned error: %v", err)
		}
		if n != 0 {
			t.Errorf("imported = %d, want 0", n)
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		data := json.RawMessage(`{bad json}`)
		_, err := h.importFromJSON("user-1", data)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})
}

// TestImportFromLinktree verifies it accepts any data without error.
func TestImportFromLinktree(t *testing.T) {
	h, _ := newTestImportExportHandler(t)
	n, err := h.importFromLinktree("user-1", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("importFromLinktree returned error: %v", err)
	}
	if n != 0 {
		t.Errorf("imported = %d, want 0", n)
	}
}

// TestImportFromLinkstack verifies it accepts any data without error.
func TestImportFromLinkstack(t *testing.T) {
	h, _ := newTestImportExportHandler(t)
	n, err := h.importFromLinkstack("user-1", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("importFromLinkstack returned error: %v", err)
	}
	if n != 0 {
		t.Errorf("imported = %d, want 0", n)
	}
}

// TestImportFromCSV verifies it accepts any data without error.
func TestImportFromCSV(t *testing.T) {
	h, _ := newTestImportExportHandler(t)
	n, err := h.importFromCSV("user-1", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("importFromCSV returned error: %v", err)
	}
	if n != 0 {
		t.Errorf("imported = %d, want 0", n)
	}
}

// TestHandleExport covers: missing profile_id, each export format, and invalid format.
func TestHandleExport(t *testing.T) {
	h, _ := newTestImportExportHandler(t)

	t.Run("missing profile_id returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/export?format=json", nil)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid format returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/export?profile_id=p1&format=xml", nil)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("json export sets correct headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/export?profile_id=p1&format=json", nil)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		cd := rr.Header().Get("Content-Disposition")
		if !strings.Contains(cd, "profile.json") {
			t.Errorf("Content-Disposition = %q, want to contain profile.json", cd)
		}
	})

	t.Run("csv export sets correct headers and has header row", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/export?profile_id=p1&format=csv", nil)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "text/csv" {
			t.Errorf("Content-Type = %q, want text/csv", ct)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "Service") {
			t.Errorf("CSV body missing header row, got: %q", body)
		}
	})

	t.Run("html export returns HTML document", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/export?profile_id=p1&format=html", nil)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("html export body missing DOCTYPE, got: %q", body[:50])
		}
	})

	t.Run("vcard export returns vCard data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/export?profile_id=p1&format=vcard", nil)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "text/vcard" {
			t.Errorf("Content-Type = %q, want text/vcard", ct)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "BEGIN:VCARD") {
			t.Errorf("vcard body missing BEGIN:VCARD, got: %q", body)
		}
	})

	t.Run("pdf export returns 501", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/export?profile_id=p1&format=pdf", nil)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusNotImplemented {
			t.Errorf("got %d, want %d", rr.Code, http.StatusNotImplemented)
		}
	})
}

// TestExportJSON verifies JSON export content directly.
func TestExportJSON(t *testing.T) {
	h, _ := newTestImportExportHandler(t)

	rr := httptest.NewRecorder()
	h.exportJSON(rr, "profile-abc")

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
	var data map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&data); err != nil {
		t.Fatalf("exportJSON output not valid JSON: %v", err)
	}
	profile, ok := data["profile"].(map[string]interface{})
	if !ok {
		t.Fatal("missing 'profile' object in JSON export")
	}
	if profile["id"] != "profile-abc" {
		t.Errorf("profile.id = %q, want profile-abc", profile["id"])
	}
}

// TestExportCSV verifies CSV export writes a header row.
func TestExportCSV(t *testing.T) {
	h, _ := newTestImportExportHandler(t)

	rr := httptest.NewRecorder()
	h.exportCSV(rr, "profile-abc")

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Service") || !strings.Contains(body, "URL") {
		t.Errorf("CSV missing expected header columns, got: %q", body)
	}
}

// TestExportHTML verifies HTML export is a valid document.
func TestExportHTML(t *testing.T) {
	h, _ := newTestImportExportHandler(t)

	rr := httptest.NewRecorder()
	h.exportHTML(rr, "profile-abc")

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("exportHTML output missing DOCTYPE declaration")
	}
	if !strings.Contains(body, "</html>") {
		t.Error("exportHTML output missing closing html tag")
	}
}

// TestExportVCard verifies vCard export format.
func TestExportVCard(t *testing.T) {
	h, _ := newTestImportExportHandler(t)

	rr := httptest.NewRecorder()
	h.exportVCard(rr, "profile-abc")

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "BEGIN:VCARD") {
		t.Error("exportVCard output missing BEGIN:VCARD")
	}
	if !strings.Contains(body, "END:VCARD") {
		t.Error("exportVCard output missing END:VCARD")
	}
	if !strings.Contains(body, "VERSION:3.0") {
		t.Error("exportVCard output missing VERSION:3.0")
	}
}

// TestGenerateGraphiQLHTML verifies the generated HTML is a complete document
// and embeds the supplied theme value.
func TestGenerateGraphiQLHTML(t *testing.T) {
	tests := []struct {
		theme string
	}{
		{"dark"},
		{"light"},
		{""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("theme="+tt.theme, func(t *testing.T) {
			html := generateGraphiQLHTML(tt.theme)

			if !strings.Contains(html, "<!DOCTYPE html>") {
				t.Error("generateGraphiQLHTML: missing DOCTYPE")
			}
			if !strings.Contains(html, "</html>") {
				t.Error("generateGraphiQLHTML: missing closing html tag")
			}
			if !strings.Contains(html, "GraphiQL") {
				t.Error("generateGraphiQLHTML: missing GraphiQL reference")
			}
			if !strings.Contains(html, "/graphql") {
				t.Error("generateGraphiQLHTML: missing /graphql endpoint reference")
			}
			if !strings.Contains(html, tt.theme) && tt.theme != "" {
				t.Errorf("generateGraphiQLHTML: theme %q not embedded in output", tt.theme)
			}
		})
	}
}

// TestGetThemePreference verifies theme resolution from query param and cookie.
func TestGetThemePreference(t *testing.T) {
	t.Run("query param dark", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/graphql?theme=dark", nil)
		if got := getThemePreference(req); got != "dark" {
			t.Errorf("got %q, want dark", got)
		}
	})

	t.Run("query param light", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/graphql?theme=light", nil)
		if got := getThemePreference(req); got != "light" {
			t.Errorf("got %q, want light", got)
		}
	})

	t.Run("invalid query param falls through to cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/graphql?theme=blue", nil)
		req.AddCookie(&http.Cookie{Name: "theme", Value: "light"})
		if got := getThemePreference(req); got != "light" {
			t.Errorf("got %q, want light", got)
		}
	})

	t.Run("invalid query param invalid cookie returns dark default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/graphql?theme=invalid", nil)
		req.AddCookie(&http.Cookie{Name: "theme", Value: "custom"})
		if got := getThemePreference(req); got != "dark" {
			t.Errorf("got %q, want dark", got)
		}
	})

	t.Run("no param no cookie returns dark default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/graphql", nil)
		if got := getThemePreference(req); got != "dark" {
			t.Errorf("got %q, want dark", got)
		}
	})

	t.Run("cookie dark no query param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/graphql", nil)
		req.AddCookie(&http.Cookie{Name: "theme", Value: "dark"})
		if got := getThemePreference(req); got != "dark" {
			t.Errorf("got %q, want dark", got)
		}
	})
}
