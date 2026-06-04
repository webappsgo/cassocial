package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewQRHandler verifies the constructor returns a non-nil handler.
func TestNewQRHandler(t *testing.T) {
	h := NewQRHandler()
	if h == nil {
		t.Fatal("NewQRHandler returned nil")
	}
}

// TestHandleGenerateQR covers: missing URL, invalid format, size clamping, each valid format.
func TestHandleGenerateQR(t *testing.T) {
	h := NewQRHandler()

	t.Run("missing url returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/qr/generate", nil)
		rr := httptest.NewRecorder()
		h.HandleGenerateQR(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid format returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/qr/generate?url=https://example.com&format=gif", nil)
		rr := httptest.NewRecorder()
		h.HandleGenerateQR(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("png format returns PNG image", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/qr/generate?url=https://example.com&format=png", nil)
		rr := httptest.NewRecorder()
		h.HandleGenerateQR(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "image/png" {
			t.Errorf("Content-Type = %q, want image/png", ct)
		}
		if rr.Body.Len() == 0 {
			t.Error("expected non-empty PNG body")
		}
	})

	t.Run("default format is png", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/qr/generate?url=https://example.com", nil)
		rr := httptest.NewRecorder()
		h.HandleGenerateQR(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "image/png" {
			t.Errorf("Content-Type = %q, want image/png", ct)
		}
	})

	t.Run("base64 format returns JSON data URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/qr/generate?url=https://example.com&format=base64", nil)
		rr := httptest.NewRecorder()
		h.HandleGenerateQR(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		var resp map[string]string
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("response not valid JSON: %v", err)
		}
		if !strings.HasPrefix(resp["qr_code"], "data:image/png;base64,") {
			t.Errorf("qr_code = %q, want data:image/png;base64,... prefix", resp["qr_code"])
		}
	})

	t.Run("svg format returns 200 with image/svg+xml", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/qr/generate?url=https://example.com&format=svg", nil)
		rr := httptest.NewRecorder()
		h.HandleGenerateQR(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "image/svg+xml" {
			t.Errorf("Content-Type = %q, want image/svg+xml", ct)
		}
	})

	t.Run("custom valid size is accepted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/qr/generate?url=https://example.com&size=512&format=png", nil)
		rr := httptest.NewRecorder()
		h.HandleGenerateQR(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
	})

	t.Run("size above 1024 falls back to 256", func(t *testing.T) {
		// size > 1024 is rejected; the handler falls back to 256; still generates OK
		req := httptest.NewRequest(http.MethodGet, "/api/qr/generate?url=https://example.com&size=9999&format=png", nil)
		rr := httptest.NewRecorder()
		h.HandleGenerateQR(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
	})

	t.Run("size zero falls back to 256", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/qr/generate?url=https://example.com&size=0&format=png", nil)
		rr := httptest.NewRecorder()
		h.HandleGenerateQR(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
	})

	t.Run("non-numeric size falls back to 256", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/qr/generate?url=https://example.com&size=big&format=png", nil)
		rr := httptest.NewRecorder()
		h.HandleGenerateQR(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
	})
}

// TestHandleProfileQR covers: missing slug, valid slug, custom size.
func TestHandleProfileQR(t *testing.T) {
	h := NewQRHandler()

	t.Run("missing slug returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/qr/profile", nil)
		rr := httptest.NewRecorder()
		h.HandleProfileQR(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid slug generates PNG", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/qr/profile?slug=alice", nil)
		req.Host = "cassocial.example.com"
		rr := httptest.NewRecorder()
		h.HandleProfileQR(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "image/png" {
			t.Errorf("Content-Type = %q, want image/png", ct)
		}
		if rr.Body.Len() == 0 {
			t.Error("expected non-empty PNG body")
		}
	})

	t.Run("custom size accepted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/qr/profile?slug=bob&size=128", nil)
		req.Host = "cassocial.example.com"
		rr := httptest.NewRecorder()
		h.HandleProfileQR(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
	})
}

// TestGeneratePNG verifies the internal helper directly.
func TestGeneratePNG(t *testing.T) {
	h := NewQRHandler()

	t.Run("short content generates valid PNG", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.generatePNG(rr, "https://example.com", 128)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "image/png" {
			t.Errorf("Content-Type = %q, want image/png", ct)
		}
		body := rr.Body.Bytes()
		// PNG magic bytes: 0x89 50 4E 47
		if len(body) < 4 || body[0] != 0x89 || body[1] != 0x50 || body[2] != 0x4E || body[3] != 0x47 {
			t.Error("response body does not begin with PNG magic bytes")
		}
	})
}

// TestGenerateSVG verifies the SVG generator returns 200 with valid SVG content.
func TestGenerateSVG(t *testing.T) {
	h := NewQRHandler()
	rr := httptest.NewRecorder()
	h.generateSVG(rr, "https://example.com", 256)
	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "<svg") {
		t.Error("generateSVG: response does not contain <svg element")
	}
	if ct := rr.Header().Get("Content-Type"); ct != "image/svg+xml" {
		t.Errorf("Content-Type = %q, want image/svg+xml", ct)
	}
}

// TestGenerateBase64 verifies the base64 helper returns a valid data URL.
func TestGenerateBase64(t *testing.T) {
	h := NewQRHandler()
	rr := httptest.NewRecorder()
	h.generateBase64(rr, "https://example.com", 128)
	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("response not valid JSON: %v", err)
	}
	if !strings.HasPrefix(resp["qr_code"], "data:image/png;base64,") {
		t.Errorf("qr_code does not have expected prefix, got: %s", resp["qr_code"][:50])
	}
}

// TestGeneratePNG_TooLongContent verifies that generatePNG returns 500 when
// the content is too long for qrcode.Encode to handle.
func TestGeneratePNG_TooLongContent(t *testing.T) {
	h := NewQRHandler()
	// QR codes at Medium recovery level can hold at most ~2953 bytes.
	// A 5000-byte string forces qrcode.Encode to return an error.
	content := strings.Repeat("x", 5000)
	rr := httptest.NewRecorder()
	h.generatePNG(rr, content, 256)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("generatePNG(oversized content) got %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

// TestGenerateBase64_TooLongContent verifies that generateBase64 returns 500
// when the content is too long for qrcode.Encode to handle.
func TestGenerateBase64_TooLongContent(t *testing.T) {
	h := NewQRHandler()
	content := strings.Repeat("x", 5000)
	rr := httptest.NewRecorder()
	h.generateBase64(rr, content, 256)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("generateBase64(oversized content) got %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}
