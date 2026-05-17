package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestPublicHandler creates a PublicHandler with an in-memory DB and default config.
func newTestPublicHandler(t *testing.T) *PublicHandler {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect returned error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	cfg := &config.Config{}
	cfg.Cassocial.SiteName = "TestSite"
	cfg.Cassocial.SiteDescription = "A test site"
	cfg.Cassocial.AllowRegistration = true

	return NewPublicHandler(cfg, db)
}

func TestNewPublicHandler_NotNil(t *testing.T) {
	h := newTestPublicHandler(t)
	if h == nil {
		t.Fatal("NewPublicHandler returned nil")
	}
}

// HandleHomepage for path "/" must return 200 with site_name in body.
func TestPublicHandler_HandleHomepage_Root(t *testing.T) {
	h := newTestPublicHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.HandleHomepage(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleHomepage / returned %d, want %d", rr.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["site_name"] != "TestSite" {
		t.Errorf("site_name = %v, want \"TestSite\"", body["site_name"])
	}
}

// HandleHomepage for a non-root path delegates to HandleProfilePage.
func TestPublicHandler_HandleHomepage_NonRoot_DelegatesToProfile(t *testing.T) {
	h := newTestPublicHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/someuser", nil)
	rr := httptest.NewRecorder()
	h.HandleHomepage(rr, req)

	// Profile page always returns 200 with a "slug" field.
	if rr.Code != http.StatusOK {
		t.Errorf("HandleHomepage /someuser returned %d, want %d", rr.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["slug"] != "someuser" {
		t.Errorf("slug = %v, want \"someuser\"", body["slug"])
	}
}

// HandleProfilePage for reserved slugs must return 404.
func TestPublicHandler_HandleProfilePage_ReservedSlugs(t *testing.T) {
	h := newTestPublicHandler(t)

	reserved := []string{"api", "admin", "healthz", ""}
	for _, slug := range reserved {
		path := "/" + slug
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		h.HandleProfilePage(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("HandleProfilePage %q returned %d, want %d", path, rr.Code, http.StatusNotFound)
		}
	}
}

// HandleProfilePage for a valid slug must return 200 with slug in body.
func TestPublicHandler_HandleProfilePage_ValidSlug(t *testing.T) {
	h := newTestPublicHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/myprofile", nil)
	rr := httptest.NewRecorder()
	h.HandleProfilePage(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleProfilePage /myprofile returned %d, want %d", rr.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["slug"] != "myprofile" {
		t.Errorf("slug = %v, want \"myprofile\"", body["slug"])
	}
}

// HandleSitemap must return 200 with XML content-type and valid urlset.
func TestPublicHandler_HandleSitemap(t *testing.T) {
	h := newTestPublicHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rr := httptest.NewRecorder()
	h.HandleSitemap(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleSitemap returned %d, want %d", rr.Code, http.StatusOK)
	}

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "xml") {
		t.Errorf("HandleSitemap Content-Type = %q, want xml", ct)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "urlset") {
		t.Errorf("HandleSitemap body missing urlset: %s", body)
	}
}

// HandleRobotsTxt must return 200 with plain text and standard directives.
func TestPublicHandler_HandleRobotsTxt(t *testing.T) {
	h := newTestPublicHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	req.Host = "example.com"
	rr := httptest.NewRecorder()
	h.HandleRobotsTxt(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleRobotsTxt returned %d, want %d", rr.Code, http.StatusOK)
	}

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("HandleRobotsTxt Content-Type = %q, want text/plain", ct)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "User-agent") {
		t.Errorf("HandleRobotsTxt body missing User-agent: %s", body)
	}
	if !strings.Contains(body, "Sitemap") {
		t.Errorf("HandleRobotsTxt body missing Sitemap: %s", body)
	}
}

// HandleSecurityTxt must return 200 with Contact field.
func TestPublicHandler_HandleSecurityTxt(t *testing.T) {
	h := newTestPublicHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/security.txt", nil)
	req.Host = "example.com"
	rr := httptest.NewRecorder()
	h.HandleSecurityTxt(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleSecurityTxt returned %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Contact:") {
		t.Errorf("HandleSecurityTxt body missing Contact field: %s", body)
	}
	if !strings.Contains(body, "Canonical:") {
		t.Errorf("HandleSecurityTxt body missing Canonical field: %s", body)
	}
}

// HandleRobotsTxt must include the request host in the Sitemap URL.
func TestPublicHandler_HandleRobotsTxt_IncludesHost(t *testing.T) {
	h := newTestPublicHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	req.Host = "mysite.example.org"
	rr := httptest.NewRecorder()
	h.HandleRobotsTxt(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "mysite.example.org") {
		t.Errorf("HandleRobotsTxt body does not contain host %q: %s", "mysite.example.org", body)
	}
}
