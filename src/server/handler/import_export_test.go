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

	// mustJSONString wraps a plain string (used for the CSV source, which is
	// plain text rather than a JSON object) as a JSON string literal so it
	// round-trips correctly through the ImportRequest.Data json.RawMessage field.
	mustJSONString := func(s string) json.RawMessage {
		b, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		return json.RawMessage(b)
	}

	sources := []struct {
		source       string
		data         json.RawMessage
		wantMinLinks int
	}{
		{
			source: "linktree",
			data: json.RawMessage(`{"accountData":{"username":"linktreeuser","displayName":"LT User","bio":"bio"},` +
				`"links":[{"title":"GH","url":"https://github.com/linktreeuser"}]}`),
			wantMinLinks: 1,
		},
		{
			source: "linkstack",
			data: json.RawMessage(`{"profile":{"username":"linkstackuser","name":"LS User","bio":"bio"},` +
				`"links":[{"title":"GH","url":"https://github.com/linkstackuser","order":1}]}`),
			wantMinLinks: 1,
		},
		{
			source:       "carrd",
			data:         json.RawMessage(`{"title":"Carrd User","bio":"bio","links":["https://github.com/carrduser"]}`),
			wantMinLinks: 1,
		},
		{
			source: "aboutme",
			data: json.RawMessage(`{"username":"aboutmeuser","name":"AM User","headline":"bio",` +
				`"links":[{"label":"GH","url":"https://github.com/aboutmeuser"}]}`),
			wantMinLinks: 1,
		},
		{
			source:       "csv",
			data:         mustJSONString("title,url\nGitHub,https://github.com/csvuser\n"),
			wantMinLinks: 1,
		},
		{
			source: "json",
			data: json.RawMessage(`{"profile":{"slug":"jsonimportuser","display_name":"JSON User"},` +
				`"links":[{"title":"GH","url":"https://github.com/jsonuser"}]}`),
			wantMinLinks: 1,
		},
	}

	for _, s := range sources {
		s := s
		t.Run("source="+s.source+" returns 200", func(t *testing.T) {
			sourceUserID := createTestUser(t, db, "importuser-"+s.source, s.source+"@example.com")
			body, _ := json.Marshal(ImportRequest{Source: s.source, Data: s.data})
			req := httptest.NewRequest(http.MethodPost, "/api/import", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = withUserID(req, sourceUserID)
			rr := httptest.NewRecorder()
			h.HandleImport(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("source=%s got %d, want %d; body: %s", s.source, rr.Code, http.StatusOK, rr.Body.String())
			}
			var resp map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("source=%s: response not valid JSON: %v", s.source, err)
			}
			if resp["status"] != "success" {
				t.Errorf("source=%s status = %q, want success", s.source, resp["status"])
			}
			if resp["job_id"] == "" || resp["job_id"] == nil {
				t.Errorf("source=%s: missing job_id in response", s.source)
			}
			result, ok := resp["result"].(map[string]interface{})
			if !ok {
				t.Fatalf("source=%s: missing result object; got %v", s.source, resp)
			}
			if _, ok := result["profile_id"]; !ok {
				t.Errorf("source=%s: result missing profile_id", s.source)
			}
			linksImported, _ := result["links_imported"].(float64)
			if int(linksImported) < s.wantMinLinks {
				t.Errorf("source=%s: links_imported = %v, want >= %d", s.source, result["links_imported"], s.wantMinLinks)
			}
		})
	}
}

// TestHandleImportStatus covers job-status polling: missing job ID, missing
// auth, unknown job, non-owner access, and a successful owner lookup.
func TestHandleImportStatus(t *testing.T) {
	h, db := newTestImportExportHandler(t)
	userID := createTestUser(t, db, "importstatususer", "importstatus@example.com")
	otherID := createTestUser(t, db, "importstatusother", "importstatusother@example.com")

	body, _ := json.Marshal(ImportRequest{
		Source: "json",
		Data: json.RawMessage(`{"profile":{"slug":"statususer","display_name":"Status User"},` +
			`"links":[]}`),
	})
	importReq := httptest.NewRequest(http.MethodPost, "/api/import", bytes.NewReader(body))
	importReq.Header.Set("Content-Type", "application/json")
	importReq = withUserID(importReq, userID)
	importRR := httptest.NewRecorder()
	h.HandleImport(importRR, importReq)
	if importRR.Code != http.StatusOK {
		t.Fatalf("setup HandleImport returned %d, want %d; body: %s",
			importRR.Code, http.StatusOK, importRR.Body.String())
	}
	var importResp map[string]interface{}
	if err := json.NewDecoder(importRR.Body).Decode(&importResp); err != nil {
		t.Fatalf("failed to decode setup import response: %v", err)
	}
	jobID, _ := importResp["job_id"].(string)
	if jobID == "" {
		t.Fatal("setup import response missing job_id")
	}

	t.Run("missing job_id returns 400", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, "/api/import", nil), userID)
		rr := httptest.NewRecorder()
		h.HandleImportStatus(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing auth returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/import?job_id="+jobID, nil)
		rr := httptest.NewRecorder()
		h.HandleImportStatus(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("unknown job returns 404", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, "/api/import?job_id=nope", nil), userID)
		rr := httptest.NewRecorder()
		h.HandleImportStatus(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("non-owner returns 403", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, "/api/import?job_id="+jobID, nil), otherID)
		rr := httptest.NewRecorder()
		h.HandleImportStatus(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Errorf("got %d, want %d", rr.Code, http.StatusForbidden)
		}
	})

	t.Run("owner sees completed job", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, "/api/import?job_id="+jobID, nil), userID)
		rr := httptest.NewRecorder()
		h.HandleImportStatus(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("got %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		var job map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&job); err != nil {
			t.Fatalf("job response not valid JSON: %v", err)
		}
		if job["status"] != "completed" {
			t.Errorf("job status = %v, want completed", job["status"])
		}
	})
}

// TestHandleExport covers input validation, auth, ownership, and every real
// export format produced by the delegated ExportService.
func TestHandleExport(t *testing.T) {
	h, db := newTestImportExportHandler(t)
	userID := createTestUser(t, db, "exportuser", "export@example.com")
	otherID := createTestUser(t, db, "exportother", "exportother@example.com")
	profileID := "export-handler-profile-1"
	if err := db.CreateProfile(&store.Profile{
		ID:       profileID,
		UserID:   userID,
		Slug:     "exporthandlerslug",
		IsPublic: true,
	}); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	exportURL := func(format string) string {
		return "/api/export?profile_id=" + profileID + "&format=" + format
	}

	t.Run("missing profile_id returns 400", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, "/api/export?format=json", nil), userID)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("no auth returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, exportURL("json"), nil)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("non-owner returns 403", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, exportURL("json"), nil), otherID)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Errorf("got %d, want %d", rr.Code, http.StatusForbidden)
		}
	})

	t.Run("invalid format returns 400", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, exportURL("xml"), nil), userID)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("json export returns real profile data", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, exportURL("json"), nil), userID)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("got %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		var data map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&data); err != nil {
			t.Fatalf("export output not valid JSON: %v", err)
		}
		if _, ok := data["profile"].(map[string]interface{}); !ok {
			t.Error("missing 'profile' object in JSON export")
		}
	})

	t.Run("csv export sets correct content type", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, exportURL("csv"), nil), userID)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("got %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		if ct := rr.Header().Get("Content-Type"); ct != "text/csv" {
			t.Errorf("Content-Type = %q, want text/csv", ct)
		}
	})

	t.Run("html export returns HTML document", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, exportURL("html"), nil), userID)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("got %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), "<!DOCTYPE html>") {
			t.Error("html export body missing DOCTYPE")
		}
	})

	t.Run("vcard export returns vCard data", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, exportURL("vcard"), nil), userID)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("got %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		if ct := rr.Header().Get("Content-Type"); ct != "text/vcard" {
			t.Errorf("Content-Type = %q, want text/vcard", ct)
		}
		if !strings.Contains(rr.Body.String(), "BEGIN:VCARD") {
			t.Error("vcard body missing BEGIN:VCARD")
		}
	})

	t.Run("pdf export returns a real PDF", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, exportURL("pdf"), nil), userID)
		rr := httptest.NewRecorder()
		h.HandleExport(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("got %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		if ct := rr.Header().Get("Content-Type"); ct != "application/pdf" {
			t.Errorf("Content-Type = %q, want application/pdf", ct)
		}
		if !strings.HasPrefix(rr.Body.String(), "%PDF") {
			t.Error("pdf export body does not start with %PDF")
		}
	})
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
