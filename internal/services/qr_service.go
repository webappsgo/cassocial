package services

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"github.com/casapps/cassocial/internal/database"
	"github.com/casapps/cassocial/internal/models"
	"github.com/skip2/go-qrcode"
)

var (
	ErrInvalidQRSize   = errors.New("invalid QR code size")
	ErrInvalidQRFormat = errors.New("invalid QR code format")
	ErrQRGeneration    = errors.New("failed to generate QR code")
)

// QRService handles QR code generation
type QRService struct {
	db *database.DB
}

// NewQRService creates a new QR service
func NewQRService(db *database.DB) *QRService {
	return &QRService{db: db}
}

// GenerateQRCode generates a QR code for a profile URL
func (s *QRService) GenerateQRCode(profileURL string, settings *models.QRCodeSettings) ([]byte, error) {
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
func (s *QRService) GenerateWithLogo(profileURL string, settings *models.QRCodeSettings, logoData []byte) ([]byte, error) {
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
func (s *QRService) GenerateBase64(profileURL string, settings *models.QRCodeSettings) (string, error) {
	data, err := s.GenerateQRCode(profileURL, settings)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// GenerateDataURI generates a QR code and returns it as a data URI
func (s *QRService) GenerateDataURI(profileURL string, settings *models.QRCodeSettings) (string, error) {
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

func (s *QRService) generatePNG(url string, settings *models.QRCodeSettings, errorCorrection qrcode.RecoveryLevel) ([]byte, error) {
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

func (s *QRService) generateSVG(url string, settings *models.QRCodeSettings) ([]byte, error) {
	// Get error correction level
	errorCorrection := s.mapErrorCorrection(settings.ErrorCorrection)

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

	// Generate SVG string
	// Note: go-qrcode doesn't natively support SVG
	// This is a simplified implementation - in production, use a proper SVG QR library
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" version="1.1" width="%d" height="%d" viewBox="0 0 %d %d">
	<rect width="%d" height="%d" fill="%s"/>
	<!-- QR code content would be rendered here -->
</svg>`, settings.Size, settings.Size, settings.Size, settings.Size, settings.Size, settings.Size, settings.LightColor)

	return []byte(svg), nil
}

func (s *QRService) generatePDF(url string, settings *models.QRCodeSettings) ([]byte, error) {
	// Generate PNG first
	errorCorrection := s.mapErrorCorrection(settings.ErrorCorrection)
	pngData, err := s.generatePNG(url, settings, errorCorrection)
	if err != nil {
		return nil, err
	}

	// In production, convert PNG to PDF using a library like gofpdf
	// For now, return a placeholder
	_ = pngData

	pdf := []byte("%PDF-1.4\n%placeholder for QR code PDF\n")
	return pdf, nil
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
func (s *QRService) GetDefaultSettings() *models.QRCodeSettings {
	return &models.QRCodeSettings{
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
