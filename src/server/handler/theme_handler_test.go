package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

func newTestThemeHandler(t *testing.T) *ThemeHandler {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	return NewThemeHandler(&config.Config{}, db)
}

// TestNewThemeHandler verifies constructor.
func TestNewThemeHandler(t *testing.T) {
	h := newTestThemeHandler(t)
	if h == nil {
		t.Fatal("NewThemeHandler returned nil")
	}
}

// TestHandleGetThemes verifies the endpoint returns the built-in theme list.
func TestHandleGetThemes(t *testing.T) {
	h := newTestThemeHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/themes", nil)
	rr := httptest.NewRecorder()
	h.HandleGetThemes(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("response not valid JSON: %v", err)
	}
	themes, ok := resp["themes"].([]interface{})
	if !ok {
		t.Fatal("response 'themes' field is not a slice")
	}
	if len(themes) == 0 {
		t.Error("expected at least one built-in theme")
	}
	total, ok := resp["total"].(float64)
	if !ok {
		t.Error("response 'total' field missing or wrong type")
	}
	if int(total) != len(themes) {
		t.Errorf("total = %d, want %d (matching len of themes)", int(total), len(themes))
	}
}

// TestHandleGetTheme covers: missing ID, valid ID.
func TestHandleGetTheme(t *testing.T) {
	h := newTestThemeHandler(t)

	t.Run("missing id returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/themes/get", nil)
		rr := httptest.NewRecorder()
		h.HandleGetTheme(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
		var resp map[string]string
		json.NewDecoder(rr.Body).Decode(&resp)
		if resp["error"] == "" {
			t.Error("expected error field in response")
		}
	})

	t.Run("valid id returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/themes/get?id=dark-dracula", nil)
		rr := httptest.NewRecorder()
		h.HandleGetTheme(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		var resp map[string]string
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("response not valid JSON: %v", err)
		}
		if resp["id"] != "dark-dracula" {
			t.Errorf("id = %q, want dark-dracula", resp["id"])
		}
	})
}

// TestHandleSaveCustomTheme covers: wrong method, bad JSON, missing ID, valid theme.
func TestHandleSaveCustomTheme(t *testing.T) {
	h := newTestThemeHandler(t)

	t.Run("wrong method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/themes/custom", nil)
		rr := httptest.NewRecorder()
		h.HandleSaveCustomTheme(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("bad JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/themes/custom", bytes.NewBufferString("{bad"))
		rr := httptest.NewRecorder()
		h.HandleSaveCustomTheme(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing theme ID returns 400", func(t *testing.T) {
		theme := Theme{Name: "My Theme", Type: "dark"}
		body, _ := json.Marshal(theme)
		req := httptest.NewRequest(http.MethodPost, "/api/themes/custom", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.HandleSaveCustomTheme(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	for _, method := range []string{http.MethodPost, http.MethodPut} {
		method := method
		t.Run("valid theme via "+method+" returns 200", func(t *testing.T) {
			theme := Theme{
				ID:   "my-custom-theme",
				Name: "My Theme",
				Type: "dark",
				Colors: ColorConfig{
					Primary: "#ff0000",
				},
			}
			body, _ := json.Marshal(theme)
			req := httptest.NewRequest(method, "/api/themes/custom", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h.HandleSaveCustomTheme(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("method=%s got %d, want %d", method, rr.Code, http.StatusOK)
			}
			var resp map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("response not valid JSON: %v", err)
			}
			if resp["status"] != "success" {
				t.Errorf("status = %q, want success", resp["status"])
			}
			if resp["theme_id"] != "my-custom-theme" {
				t.Errorf("theme_id = %q, want my-custom-theme", resp["theme_id"])
			}
		})
	}
}

// TestHandleGetGradientPresets verifies the endpoint returns gradient presets.
func TestHandleGetGradientPresets(t *testing.T) {
	h := newTestThemeHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/themes/gradients", nil)
	rr := httptest.NewRecorder()
	h.HandleGetGradientPresets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("response not valid JSON: %v", err)
	}
	presets, ok := resp["presets"].([]interface{})
	if !ok {
		t.Fatal("response 'presets' field is not a slice")
	}
	if len(presets) == 0 {
		t.Error("expected at least one gradient preset")
	}
}
