package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/skip2/go-qrcode"
)

// QRHandler handles QR code generation
type QRHandler struct{}

// NewQRHandler creates a new QR code handler
func NewQRHandler() *QRHandler {
	return &QRHandler{}
}

// HandleGenerateQR generates a QR code
func (h *QRHandler) HandleGenerateQR(w http.ResponseWriter, r *http.Request) {
	// Get URL to encode
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "URL parameter required", http.StatusBadRequest)
		return
	}

	// Get size (default 256)
	sizeStr := r.URL.Query().Get("size")
	size := 256
	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 1024 {
			size = s
		}
	}

	// Get format (default png)
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "png"
	}

	// Generate QR code
	switch format {
	case "png":
		h.generatePNG(w, url, size)
	case "svg":
		h.generateSVG(w, url, size)
	case "base64":
		h.generateBase64(w, url, size)
	default:
		http.Error(w, "Invalid format. Use png, svg, or base64", http.StatusBadRequest)
	}
}

// HandleProfileQR generates a QR code for a profile
func (h *QRHandler) HandleProfileQR(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Slug parameter required", http.StatusBadRequest)
		return
	}

	// Build profile URL using the request host
	profileURL := fmt.Sprintf("https://%s/%s", r.Host, slug)

	// Get size (default 256)
	sizeStr := r.URL.Query().Get("size")
	size := 256
	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 1024 {
			size = s
		}
	}

	// Generate and serve QR code as PNG
	h.generatePNG(w, profileURL, size)
}

// generatePNG generates and serves a PNG QR code
func (h *QRHandler) generatePNG(w http.ResponseWriter, content string, size int) {
	png, err := qrcode.Encode(content, qrcode.Medium, size)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(png)
}

// generateSVG generates and serves an SVG QR code
func (h *QRHandler) generateSVG(w http.ResponseWriter, content string, size int) {
	http.Error(w, "SVG format not yet implemented", http.StatusNotImplemented)
}

// generateBase64 generates and returns a base64-encoded QR code
func (h *QRHandler) generateBase64(w http.ResponseWriter, content string, size int) {
	png, err := qrcode.Encode(content, qrcode.Medium, size)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	encoded := base64.StdEncoding.EncodeToString(png)
	dataURL := fmt.Sprintf("data:image/png;base64,%s", encoded)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"qr_code": dataURL,
	})
}
