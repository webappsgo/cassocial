package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewPWA(t *testing.T) {
	p := NewPWA("My Site", "A great social site")
	if p == nil {
		t.Fatal("NewPWA returned nil")
	}
	if p.manifest == nil {
		t.Fatal("manifest is nil")
	}
	if p.manifest.Name != "My Site" {
		t.Errorf("Name = %q, want %q", p.manifest.Name, "My Site")
	}
	if p.manifest.Description != "A great social site" {
		t.Errorf("Description = %q, want %q", p.manifest.Description, "A great social site")
	}
}

func TestNewPWA_DefaultFields(t *testing.T) {
	p := NewPWA("Test", "Desc")
	if p.manifest.Display != "standalone" {
		t.Errorf("Display = %q, want standalone", p.manifest.Display)
	}
	if p.manifest.StartURL != "/" {
		t.Errorf("StartURL = %q, want /", p.manifest.StartURL)
	}
	if len(p.manifest.Icons) == 0 {
		t.Error("Icons slice is empty")
	}
}

func TestServeManifest(t *testing.T) {
	p := NewPWA("Cassocial", "Social profiles")
	req := httptest.NewRequest("GET", "/manifest.json", nil)
	rr := httptest.NewRecorder()
	p.ServeManifest(rr, req)

	if rr.Code != 200 {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "manifest") {
		t.Errorf("Content-Type = %q, want application/manifest+json", ct)
	}

	var manifest WebAppManifest
	if err := json.NewDecoder(rr.Body).Decode(&manifest); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if manifest.Name != "Cassocial" {
		t.Errorf("manifest name = %q, want %q", manifest.Name, "Cassocial")
	}
}

func TestServeServiceWorker(t *testing.T) {
	p := NewPWA("Cassocial", "Social profiles")
	req := httptest.NewRequest("GET", "/sw.js", nil)
	rr := httptest.NewRecorder()
	p.ServeServiceWorker(rr, req)

	if rr.Code != 200 {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "javascript") {
		t.Errorf("Content-Type = %q, want application/javascript", ct)
	}
	if swa := rr.Header().Get("Service-Worker-Allowed"); swa != "/" {
		t.Errorf("Service-Worker-Allowed = %q, want /", swa)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "CACHE_NAME") {
		t.Error("service worker body missing CACHE_NAME")
	}
}

func TestGetInstallPromptHTML(t *testing.T) {
	p := NewPWA("Cassocial", "Social profiles")
	html := p.GetInstallPromptHTML()
	if !strings.Contains(html, "pwa-install-prompt") {
		t.Error("install prompt HTML missing pwa-install-prompt id")
	}
	if !strings.Contains(html, "Install Cassocial") {
		t.Error("install prompt HTML missing title")
	}
}

func TestGetOfflinePage(t *testing.T) {
	page := GetOfflinePage()
	if !strings.Contains(page, "You're Offline") {
		t.Error("offline page missing 'You're Offline' heading")
	}
	if !strings.Contains(page, "<!DOCTYPE html>") {
		t.Error("offline page missing DOCTYPE")
	}
}

func TestGetPWAMetaTags_Default(t *testing.T) {
	tags := GetPWAMetaTags("")
	if !strings.Contains(tags, "#bd93f9") {
		t.Error("meta tags should contain default theme color #bd93f9")
	}
	if !strings.Contains(tags, "manifest") {
		t.Error("meta tags missing manifest link")
	}
}

func TestGetPWAMetaTags_Custom(t *testing.T) {
	tags := GetPWAMetaTags("#ff0000")
	if !strings.Contains(tags, "#ff0000") {
		t.Error("meta tags should contain custom theme color")
	}
}
