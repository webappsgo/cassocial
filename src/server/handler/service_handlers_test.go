package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/casapps/cassocial/src/server/store"
)

// newTestServiceHandlers creates a ServiceHandlers backed by an in-memory SQLite database.
func newTestServiceHandlers(t *testing.T) (*ServiceHandlers, *store.DB) {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect returned error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	return NewServiceHandlers(db), db
}

// insertTestService inserts a service row directly into the DB and returns its ID.
// All nullable text columns are given empty-string values so that scanning into
// model.Service (which uses string, not *string) does not produce NULL errors.
func insertTestService(t *testing.T, db *store.DB, name, category string, popularity int, isActive bool) string {
	t.Helper()

	id := generateUUID()
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	_, err := db.Exec(
		`INSERT INTO services
		 (id, name, category, icon_url, icon_svg, url_pattern, background_color, text_color,
		  popularity, is_active, requires_username, placeholder_text, validation_pattern,
		  created_at, updated_at)
		 VALUES (?, ?, ?, '', '', '', '', '', ?, ?, 1, '', '', ?, ?)`,
		id, name, category, popularity, isActive, now, now,
	)
	if err != nil {
		t.Fatalf("insertTestService: INSERT returned error: %v", err)
	}
	return id
}

// ---- NewServiceHandlers ----

func TestNewServiceHandlers(t *testing.T) {
	h, _ := newTestServiceHandlers(t)
	if h == nil {
		t.Fatal("NewServiceHandlers returned nil")
	}
}

// ---- ListServices ----

func TestServiceHandlers_ListServices_Empty(t *testing.T) {
	h, _ := newTestServiceHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	rr := httptest.NewRecorder()
	h.ListServices(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListServices returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d entries", len(resp))
	}
}

func TestServiceHandlers_ListServices_WithServices(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	insertTestService(t, db, "GitHub", "development", 100, true)
	insertTestService(t, db, "Twitter", "social", 90, true)

	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	rr := httptest.NewRecorder()
	h.ListServices(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListServices returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 services, got %d", len(resp))
	}
}

func TestServiceHandlers_ListServices_InactiveExcluded(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	insertTestService(t, db, "Active", "social", 10, true)
	insertTestService(t, db, "Inactive", "social", 10, false)

	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	rr := httptest.NewRecorder()
	h.ListServices(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListServices returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 active service, got %d", len(resp))
	}
}

func TestServiceHandlers_ListServices_FilterByCategory(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	insertTestService(t, db, "GitHub", "development", 100, true)
	insertTestService(t, db, "Twitter", "social", 90, true)
	insertTestService(t, db, "GitLab", "development", 80, true)

	req := httptest.NewRequest(http.MethodGet, "/api/services?category=development", nil)
	rr := httptest.NewRecorder()
	h.ListServices(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListServices with category filter returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 development services, got %d", len(resp))
	}
}

func TestServiceHandlers_ListServices_WithLimit(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	for i, name := range []string{"SvcA", "SvcB", "SvcC", "SvcD", "SvcE"} {
		insertTestService(t, db, name, "social", i*10, true)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/services?limit=3", nil)
	rr := httptest.NewRecorder()
	h.ListServices(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListServices with limit returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 3 {
		t.Errorf("expected 3 services with limit=3, got %d", len(resp))
	}
}

// TestServiceHandlers_ListServices_WithOffset verifies that supplying offset
// without limit results in a 500 because SQLite does not allow OFFSET without
// a preceding LIMIT clause.  The handler propagates the DB error as a 500.
func TestServiceHandlers_ListServices_WithOffset(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	for i, name := range []string{"OffA", "OffB", "OffC"} {
		insertTestService(t, db, name, "social", i*10, true)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/services?offset=1", nil)
	rr := httptest.NewRecorder()
	h.ListServices(rr, req)

	// SQLite does not support OFFSET without LIMIT; the handler returns 500.
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("ListServices with offset-only returned %d, want %d; body: %s",
			rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

func TestServiceHandlers_ListServices_WithLimitAndOffset(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	for i, name := range []string{"LoA", "LoB", "LoC", "LoD"} {
		insertTestService(t, db, name, "social", i*10, true)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/services?limit=2&offset=1", nil)
	rr := httptest.NewRecorder()
	h.ListServices(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListServices with limit+offset returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 services with limit=2&offset=1, got %d", len(resp))
	}
}

// ---- SearchServices ----

func TestServiceHandlers_SearchServices_MissingQuery(t *testing.T) {
	h, _ := newTestServiceHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/services/search", nil)
	rr := httptest.NewRecorder()
	h.SearchServices(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("SearchServices with no query returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestServiceHandlers_SearchServices_NoResults(t *testing.T) {
	h, _ := newTestServiceHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/services/search?q=nonexistentxyz", nil)
	rr := httptest.NewRecorder()
	h.SearchServices(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("SearchServices no-results returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp))
	}
}

func TestServiceHandlers_SearchServices_WithResults(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	insertTestService(t, db, "GitHub", "development", 100, true)
	insertTestService(t, db, "GitLab", "development", 90, true)
	insertTestService(t, db, "Twitter", "social", 80, true)

	req := httptest.NewRequest(http.MethodGet, "/api/services/search?q=git", nil)
	rr := httptest.NewRecorder()
	h.SearchServices(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("SearchServices returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 results for 'git', got %d", len(resp))
	}
}

func TestServiceHandlers_SearchServices_InactiveExcluded(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	insertTestService(t, db, "FindMe", "social", 10, true)
	insertTestService(t, db, "FindMeInactive", "social", 10, false)

	req := httptest.NewRequest(http.MethodGet, "/api/services/search?q=findme", nil)
	rr := httptest.NewRecorder()
	h.SearchServices(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("SearchServices returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 result (inactive excluded), got %d", len(resp))
	}
}

// ---- ListCategories ----

func TestServiceHandlers_ListCategories_Empty(t *testing.T) {
	h, _ := newTestServiceHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/services/categories", nil)
	rr := httptest.NewRecorder()
	h.ListCategories(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListCategories returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty categories list, got %d", len(resp))
	}
}

func TestServiceHandlers_ListCategories_WithServices(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	insertTestService(t, db, "GitHub", "development", 100, true)
	insertTestService(t, db, "GitLab", "development", 90, true)
	insertTestService(t, db, "Twitter", "social", 80, true)

	req := httptest.NewRequest(http.MethodGet, "/api/services/categories", nil)
	rr := httptest.NewRecorder()
	h.ListCategories(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListCategories returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 categories, got %d", len(resp))
	}

	// Verify structure: each entry has "category" and "count".
	for _, cat := range resp {
		if _, ok := cat["category"]; !ok {
			t.Error("category entry missing 'category' field")
		}
		if _, ok := cat["count"]; !ok {
			t.Error("category entry missing 'count' field")
		}
	}
}

func TestServiceHandlers_ListCategories_InactiveExcluded(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	insertTestService(t, db, "ActiveSvc", "social", 10, true)
	insertTestService(t, db, "InactiveSvc", "gaming", 10, false)

	req := httptest.NewRequest(http.MethodGet, "/api/services/categories", nil)
	rr := httptest.NewRecorder()
	h.ListCategories(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListCategories returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	// Only "social" should appear; "gaming" is inactive.
	if len(resp) != 1 {
		t.Errorf("expected 1 category (inactive excluded), got %d", len(resp))
	}
}

// ---- ListPopularServices ----

func TestServiceHandlers_ListPopularServices_Empty(t *testing.T) {
	h, _ := newTestServiceHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/services/popular", nil)
	rr := httptest.NewRecorder()
	h.ListPopularServices(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListPopularServices returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d", len(resp))
	}
}

func TestServiceHandlers_ListPopularServices_DefaultLimit(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	for i, name := range []string{"Pop1", "Pop2", "Pop3"} {
		insertTestService(t, db, name, "social", (3-i)*10, true)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/services/popular", nil)
	rr := httptest.NewRecorder()
	h.ListPopularServices(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListPopularServices returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 3 {
		t.Errorf("expected 3 popular services, got %d", len(resp))
	}
}

func TestServiceHandlers_ListPopularServices_CustomLimit(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	for i, name := range []string{"PopA", "PopB", "PopC", "PopD", "PopE"} {
		insertTestService(t, db, name, "social", (5-i)*10, true)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/services/popular?limit=2", nil)
	rr := httptest.NewRecorder()
	h.ListPopularServices(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListPopularServices with limit returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 popular services with limit=2, got %d", len(resp))
	}
}

// ---- GetService ----

func TestServiceHandlers_GetService_MissingID(t *testing.T) {
	h, _ := newTestServiceHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/services/", nil)
	rr := httptest.NewRecorder()
	h.GetService(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("GetService with empty ID returned %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestServiceHandlers_GetService_NotFound(t *testing.T) {
	h, _ := newTestServiceHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/services/nonexistent-id", nil)
	req.SetPathValue("id", "nonexistent-id")
	rr := httptest.NewRecorder()
	h.GetService(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("GetService with nonexistent ID returned %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestServiceHandlers_GetService_Valid(t *testing.T) {
	h, db := newTestServiceHandlers(t)

	id := insertTestService(t, db, "TestSvc", "social", 50, true)

	req := httptest.NewRequest(http.MethodGet, "/api/services/"+id, nil)
	req.SetPathValue("id", id)
	rr := httptest.NewRecorder()
	h.GetService(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GetService returned %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["id"] != id {
		t.Errorf("expected id=%s, got %v", id, resp["id"])
	}
	if resp["name"] != "TestSvc" {
		t.Errorf("expected name=TestSvc, got %v", resp["name"])
	}
}
