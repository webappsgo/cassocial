package service

import (
	"encoding/json"
	"testing"

	"github.com/casapps/cassocial/src/server/model"
	"github.com/casapps/cassocial/src/server/store"
)

func newTestExportService(t *testing.T) (*ExportService, *store.DB) {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	ps := NewProfileService(db)
	ls := NewLinkService(db)
	es := NewExportService(db, ps, ls)

	return es, db
}

func TestEscapeHTML_Ampersand(t *testing.T) {
	got := escapeHTML("a & b")
	want := "a &amp; b"
	if got != want {
		t.Errorf("escapeHTML(%q) = %q, want %q", "a & b", got, want)
	}
}

func TestEscapeHTML_LtGt(t *testing.T) {
	got := escapeHTML("<script>")
	want := "&lt;script&gt;"
	if got != want {
		t.Errorf("escapeHTML(%q) = %q, want %q", "<script>", got, want)
	}
}

func TestEscapeHTML_Quotes(t *testing.T) {
	got := escapeHTML(`say "hello" & 'world'`)
	expected := `say &quot;hello&quot; &amp; &#39;world&#39;`
	if got != expected {
		t.Errorf("escapeHTML quotes: got %q, want %q", got, expected)
	}
}

func TestEscapeHTML_NoSpecialChars(t *testing.T) {
	plain := "Hello World 123"
	got := escapeHTML(plain)
	if got != plain {
		t.Errorf("escapeHTML(%q) = %q, want identity", plain, got)
	}
}

func TestEscapeHTML_Empty(t *testing.T) {
	if got := escapeHTML(""); got != "" {
		t.Errorf("escapeHTML('') = %q, want empty", got)
	}
}

func TestNewExportService(t *testing.T) {
	es, _ := newTestExportService(t)
	if es == nil {
		t.Fatal("NewExportService returned nil")
	}
}

func TestExportToJSON_ValidProfile(t *testing.T) {
	es, _ := newTestExportService(t)

	profile := &model.Profile{
		ID:          "test-id",
		Slug:        "testslug",
		DisplayName: "Test User",
		Bio:         "Test bio",
		IsPublic:    true,
	}
	links := []*model.Link{
		{
			ID:       "link-1",
			Title:    "GitHub",
			URL:      "https://github.com",
			IsActive: true,
		},
	}

	data, contentType, err := es.exportToJSON(profile, links)
	if err != nil {
		t.Fatalf("exportToJSON returned error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("exportToJSON returned empty data")
	}
	if contentType == "" {
		t.Error("exportToJSON returned empty content type")
	}

	// Validate the JSON is parseable.
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("exportToJSON returned invalid JSON: %v", err)
	}

	if _, ok := result["profile"]; !ok {
		t.Error("exportToJSON JSON missing 'profile' key")
	}
	if _, ok := result["links"]; !ok {
		t.Error("exportToJSON JSON missing 'links' key")
	}
}

func TestExportToJSON_EmptyLinks(t *testing.T) {
	es, _ := newTestExportService(t)

	profile := &model.Profile{
		ID:          "no-links-id",
		Slug:        "nolinkslug",
		DisplayName: "No Links",
	}

	data, _, err := es.exportToJSON(profile, []*model.Link{})
	if err != nil {
		t.Fatalf("exportToJSON returned error for empty links: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("exportToJSON returned empty data for empty links")
	}
}
