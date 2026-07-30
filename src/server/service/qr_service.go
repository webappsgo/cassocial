package service

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"

	"github.com/casapps/cassocial/src/server/store"
	"github.com/casapps/cassocial/src/server/model"
	"github.com/skip2/go-qrcode"
)

var (
	ErrInvalidQRSize   = errors.New("invalid QR code size")
	ErrInvalidQRFormat = errors.New("invalid QR code format")
	ErrQRGeneration    = errors.New("failed to generate QR code")
)

// QRService handles QR code generation
type QRService struct {
	db *store.DB
}

// NewQRService creates a new QR service
func NewQRService(db *store.DB) *QRService {
	return &QRService{db: db}
}

// GenerateQRCode generates a QR code for a profile URL
func (s *QRService) GenerateQRCode(profileURL string, settings *model.QRCodeSettings) ([]byte, error) {
	// Validate settings
	if err := settings.Validate(); err != nil {
		return nil, fmt.Errorf("invalid QR settings: %w", err)
	}

	// Map error correction level
	errorCorrection := s.mapErrorCorrection(settings.ErrorCorrection)

	// Generate QR code based on format
	switch settings.Format {
	case "png":
		return s.generatePNG(profileURL, settings, errorCorrection)
	case "svg":
		return s.generateSVG(profileURL, settings)
	case "pdf":
		return s.generatePDF(profileURL, settings)
	default:
		return nil, ErrInvalidQRFormat
	}
}

// GenerateWithLogo generates a QR code with an embedded logo
func (s *QRService) GenerateWithLogo(profileURL string, settings *model.QRCodeSettings, logoData []byte) ([]byte, error) {
	if !settings.LogoEnabled {
		return s.GenerateQRCode(profileURL, settings)
	}

	// Generate base QR code
	baseQR, err := s.GenerateQRCode(profileURL, settings)
	if err != nil {
		return nil, err
	}

	// Decode base QR image
	baseImg, err := png.Decode(bytes.NewReader(baseQR))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base QR: %w", err)
	}

	// Decode logo image
	logoImg, _, err := image.Decode(bytes.NewReader(logoData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode logo: %w", err)
	}

	// Embed logo in center
	result := s.embedLogo(baseImg, logoImg, settings.LogoSize)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, result); err != nil {
		return nil, fmt.Errorf("failed to encode result: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateBase64 generates a QR code and returns it as base64
func (s *QRService) GenerateBase64(profileURL string, settings *model.QRCodeSettings) (string, error) {
	data, err := s.GenerateQRCode(profileURL, settings)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// GenerateDataURI generates a QR code and returns it as a data URI
func (s *QRService) GenerateDataURI(profileURL string, settings *model.QRCodeSettings) (string, error) {
	data, err := s.GenerateQRCode(profileURL, settings)
	if err != nil {
		return "", err
	}

	base64Data := base64.StdEncoding.EncodeToString(data)

	mimeType := "image/png"
	if settings.Format == "svg" {
		mimeType = "image/svg+xml"
	} else if settings.Format == "pdf" {
		mimeType = "application/pdf"
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data), nil
}

// Private helper methods

func (s *QRService) generatePNG(url string, settings *model.QRCodeSettings, errorCorrection qrcode.RecoveryLevel) ([]byte, error) {
	// Create QR code
	qr, err := qrcode.New(url, errorCorrection)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	// Parse colors
	darkColor, err := s.parseColor(settings.DarkColor)
	if err != nil {
		darkColor = color.Black
	}

	lightColor, err := s.parseColor(settings.LightColor)
	if err != nil {
		lightColor = color.White
	}

	qr.ForegroundColor = darkColor
	qr.BackgroundColor = lightColor

	// Apply style
	if settings.Style == "rounded" {
		// Set rounded corners (approximation with DisableBorder)
		qr.DisableBorder = false
	} else if settings.Style == "dots" {
		// Use dots style (qrcode library doesn't support this natively)
		// In production, use a library that supports different styles
		qr.DisableBorder = true
	}

	// Generate PNG
	data, err := qr.PNG(settings.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PNG: %w", err)
	}

	return data, nil
}

func (s *QRService) generateSVG(url string, settings *model.QRCodeSettings) ([]byte, error) {
	errorCorrection := s.mapErrorCorrection(settings.ErrorCorrection)

	qr, err := qrcode.New(url, errorCorrection)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	// The module matrix is rendered directly as SVG vector rectangles so the
	// output is a scalable, scannable QR code with no raster or external deps.
	bitmap := qr.Bitmap()
	modules := len(bitmap)
	if modules == 0 {
		return nil, ErrQRGeneration
	}

	size := settings.Size
	if size <= 0 {
		size = 256
	}
	moduleSize := float64(size) / float64(modules)

	dark := settings.DarkColor
	if _, err := s.parseColor(dark); err != nil {
		dark = "#000000"
	}
	light := settings.LightColor
	if _, err := s.parseColor(light); err != nil {
		light = "#ffffff"
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" version="1.1" width="%d" height="%d" viewBox="0 0 %d %d" shape-rendering="crispEdges">
`, size, size, size, size)
	fmt.Fprintf(&b, "\t<rect width=\"%d\" height=\"%d\" fill=\"%s\"/>\n", size, size, light)
	fmt.Fprintf(&b, "\t<g fill=\"%s\">\n", dark)
	for row := 0; row < modules; row++ {
		for col := 0; col < len(bitmap[row]); col++ {
			if !bitmap[row][col] {
				continue
			}
			x := float64(col) * moduleSize
			y := float64(row) * moduleSize
			fmt.Fprintf(&b, "\t\t<rect x=\"%.3f\" y=\"%.3f\" width=\"%.3f\" height=\"%.3f\"/>\n", x, y, moduleSize, moduleSize)
		}
	}
	b.WriteString("\t</g>\n</svg>\n")

	return []byte(b.String()), nil
}

func (s *QRService) generatePDF(url string, settings *model.QRCodeSettings) ([]byte, error) {
	// Build the QR module matrix, then render it as PDF vector rectangles.
	// This keeps the output pure Go and CGO-free (no raster/PDF C libraries).
	errorCorrection := s.mapErrorCorrection(settings.ErrorCorrection)
	qr, err := qrcode.New(url, errorCorrection)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	bitmap := qr.Bitmap()
	modules := len(bitmap)
	if modules == 0 {
		return nil, ErrQRGeneration
	}

	size := float64(settings.Size)
	if size <= 0 {
		size = 256
	}
	moduleSize := size / float64(modules)

	darkColor, err := s.parseColor(settings.DarkColor)
	if err != nil {
		darkColor = color.Black
	}
	lightColor, err := s.parseColor(settings.LightColor)
	if err != nil {
		lightColor = color.White
	}

	doc := newPDFDoc(size, size)

	// Background
	lr, lg, lb := colorToUnit(lightColor)
	doc.setFillColor(lr, lg, lb)
	doc.fillRect(0, 0, size, size)

	// Dark modules. PDF's origin is bottom-left, so the top row is drawn last.
	dr, dg, db := colorToUnit(darkColor)
	doc.setFillColor(dr, dg, db)
	for row := 0; row < modules; row++ {
		for col := 0; col < len(bitmap[row]); col++ {
			if !bitmap[row][col] {
				continue
			}
			x := float64(col) * moduleSize
			y := size - float64(row+1)*moduleSize
			doc.fillRect(x, y, moduleSize, moduleSize)
		}
	}

	return doc.render(), nil
}

func (s *QRService) embedLogo(baseImg, logoImg image.Image, logoSize int) image.Image {
	// Get dimensions
	bounds := baseImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create new RGBA image
	result := image.NewRGBA(bounds)

	// Draw base QR code
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			result.Set(x, y, baseImg.At(x, y))
		}
	}

	// Calculate logo position (center)
	logoWidth := (width * logoSize) / 100
	logoHeight := (height * logoSize) / 100
	logoX := (width - logoWidth) / 2
	logoY := (height - logoHeight) / 2

	// Scale and draw logo
	logoBounds := logoImg.Bounds()
	for y := 0; y < logoHeight; y++ {
		for x := 0; x < logoWidth; x++ {
			// Map to original logo coordinates
			origX := (x * logoBounds.Dx()) / logoWidth
			origY := (y * logoBounds.Dy()) / logoHeight

			result.Set(logoX+x, logoY+y, logoImg.At(origX, origY))
		}
	}

	return result
}

func (s *QRService) mapErrorCorrection(level string) qrcode.RecoveryLevel {
	switch level {
	case "L":
		return qrcode.Low
	case "M":
		return qrcode.Medium
	case "Q":
		return qrcode.High
	case "H":
		return qrcode.Highest
	default:
		return qrcode.Medium
	}
}

func (s *QRService) parseColor(colorStr string) (color.Color, error) {
	// Parse hex color (#RRGGBB or #RGB)
	if len(colorStr) == 0 || colorStr[0] != '#' {
		return nil, errors.New("invalid color format")
	}

	colorStr = colorStr[1:] // Remove #

	var r, g, b uint8

	if len(colorStr) == 6 {
		// #RRGGBB
		fmt.Sscanf(colorStr[0:2], "%02x", &r)
		fmt.Sscanf(colorStr[2:4], "%02x", &g)
		fmt.Sscanf(colorStr[4:6], "%02x", &b)
	} else if len(colorStr) == 3 {
		// #RGB - expand to #RRGGBB
		fmt.Sscanf(colorStr[0:1], "%1x", &r)
		r = r*16 + r
		fmt.Sscanf(colorStr[1:2], "%1x", &g)
		g = g*16 + g
		fmt.Sscanf(colorStr[2:3], "%1x", &b)
		b = b*16 + b
	} else {
		return nil, errors.New("invalid color format")
	}

	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

// GetDefaultSettings returns default QR code settings
func (s *QRService) GetDefaultSettings() *model.QRCodeSettings {
	return &model.QRCodeSettings{
		Size:            256,
		ErrorCorrection: "M",
		Style:           "square",
		DarkColor:       "#000000",
		LightColor:      "#ffffff",
		LogoEnabled:     false,
		LogoSize:        30,
		Format:          "png",
	}
}
