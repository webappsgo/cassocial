package service

import (
	"encoding/json"
	"testing"

	"github.com/casapps/cassocial/src/server/store"
)

// newTestImportDB creates an in-memory DB with migrations run and a seeded user.
func newTestImportDB(t *testing.T) (*store.DB, string) {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	userID := "import-user-001"
	_, err = db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, created_at, updated_at)
		 VALUES (?, 'importuser', 'import@example.com',
		         '$argon2id$v=19$m=65536,t=3,p=4$s$h', 'user', 'active', 1, 0,
		         CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		userID,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	return db, userID
}

// newTestImportService creates an ImportService with all dependencies.
func newTestImportService(t *testing.T) (*ImportService, string) {
	t.Helper()

	db, userID := newTestImportDB(t)
	ps := NewProfileService(db)
	ls := NewLinkService(db)
	is := NewImportService(db, ps, ls)
	return is, userID
}

// ---------------------------------------------------------------------------
// NewImportService
// ---------------------------------------------------------------------------

func TestNewImportService(t *testing.T) {
	db, _ := newTestImportDB(t)
	ps := NewProfileService(db)
	ls := NewLinkService(db)
	is := NewImportService(db, ps, ls)
	if is == nil {
		t.Fatal("NewImportService returned nil")
	}
}

// ---------------------------------------------------------------------------
// ImportData – JSON source
// ---------------------------------------------------------------------------

func TestImportService_ImportData_JSON(t *testing.T) {
	svc, userID := newTestImportService(t)

	jsonData := []byte(`{
		"profile": {
			"slug": "imported-json-profile",
			"display_name": "Imported Profile",
			"bio": "Test bio"
		},
		"links": [
			{"title": "GitHub", "url": "https://github.com/test"},
			{"title": "Twitter", "url": "https://twitter.com/test"}
		]
	}`)

	jobID, err := svc.ImportData(userID, "json", jsonData)
	if err != nil {
		t.Fatalf("ImportData json: %v", err)
	}
	if jobID == "" {
		t.Error("ImportData returned empty jobID")
	}
}

// ---------------------------------------------------------------------------
// ImportData – CSV source
// ---------------------------------------------------------------------------

func TestImportService_ImportData_CSV(t *testing.T) {
	svc, userID := newTestImportService(t)

	csvData := []byte("title,url,username\nGitHub,https://github.com/test,testuser\nBlog,https://blog.example.com,")

	jobID, err := svc.ImportData(userID, "csv", csvData)
	if err != nil {
		t.Fatalf("ImportData csv: %v", err)
	}
	if jobID == "" {
		t.Error("ImportData csv returned empty jobID")
	}
}

// ---------------------------------------------------------------------------
// ImportData – Linktree source
// ---------------------------------------------------------------------------

func TestImportService_ImportData_Linktree(t *testing.T) {
	svc, userID := newTestImportService(t)

	linktreeData := []byte(`{
		"accountData": {
			"username": "linktreeuser",
			"displayName": "Linktree User",
			"bio": "From linktree",
			"profilePictureUrl": "https://example.com/avatar.png"
		},
		"links": [
			{"title": "Website", "url": "https://example.com"},
			{"title": "YouTube", "url": "https://youtube.com/c/test"}
		]
	}`)

	jobID, err := svc.ImportData(userID, "linktree", linktreeData)
	if err != nil {
		t.Fatalf("ImportData linktree: %v", err)
	}
	if jobID == "" {
		t.Error("ImportData linktree returned empty jobID")
	}
}

// ---------------------------------------------------------------------------
// ImportData – unsupported source
// ---------------------------------------------------------------------------

func TestImportService_ImportData_UnsupportedSource(t *testing.T) {
	svc, userID := newTestImportService(t)

	_, err := svc.ImportData(userID, "unsupported-platform", []byte(`{}`))
	if err != ErrUnsupportedImportSource {
		t.Errorf("ImportData unsupported = %v, want ErrUnsupportedImportSource", err)
	}
}

// ---------------------------------------------------------------------------
// ImportData – invalid data
// ---------------------------------------------------------------------------

func TestImportService_ImportData_InvalidJSON(t *testing.T) {
	svc, userID := newTestImportService(t)

	_, err := svc.ImportData(userID, "json", []byte(`{not valid json`))
	if err == nil {
		t.Error("ImportData with invalid JSON should return error")
	}
}

func TestImportService_ImportData_InvalidCSV(t *testing.T) {
	svc, userID := newTestImportService(t)

	// CSV with wrong header columns
	_, err := svc.ImportData(userID, "csv", []byte("wrong,header\nval1,val2"))
	if err == nil {
		t.Error("ImportData with invalid CSV header should return error")
	}
}

// ---------------------------------------------------------------------------
// GetImportJob
// ---------------------------------------------------------------------------

func TestImportService_GetImportJob_Found(t *testing.T) {
	svc, userID := newTestImportService(t)

	jsonData := []byte(`{
		"profile": {"slug": "job-test-profile", "display_name": "Job Test"},
		"links": []
	}`)

	jobID, err := svc.ImportData(userID, "json", jsonData)
	if err != nil {
		t.Fatalf("ImportData: %v", err)
	}

	job, err := svc.GetImportJob(jobID)
	if err != nil {
		t.Fatalf("GetImportJob: %v", err)
	}
	if job == nil {
		t.Fatal("GetImportJob returned nil")
	}
	if job["id"] != jobID {
		t.Errorf("GetImportJob id = %v, want %q", job["id"], jobID)
	}
}

func TestImportService_GetImportJob_NotFound(t *testing.T) {
	svc, _ := newTestImportService(t)

	_, err := svc.GetImportJob("no-such-job-id")
	if err != ErrImportJobNotFound {
		t.Errorf("GetImportJob(nonexistent) = %v, want ErrImportJobNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// importFromJSON (direct test)
// ---------------------------------------------------------------------------

func TestImportService_ImportFromJSON(t *testing.T) {
	svc, userID := newTestImportService(t)

	data := []byte(`{
		"profile": {
			"slug": "direct-json-import",
			"display_name": "Direct JSON Import",
			"bio": "Test"
		},
		"links": [
			{"title": "Link1", "url": "https://link1.example.com"},
			{"title": "Link2", "url": "https://link2.example.com"}
		]
	}`)

	result, err := svc.importFromJSON(userID, data)
	if err != nil {
		t.Fatalf("importFromJSON: %v", err)
	}
	if result == nil {
		t.Fatal("importFromJSON returned nil result")
	}
	if _, ok := result["profile_id"]; !ok {
		t.Error("result missing profile_id key")
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 2 {
		t.Errorf("links_imported = %v, want 2", result["links_imported"])
	}
}

func TestImportService_ImportFromJSON_EmptySlug(t *testing.T) {
	svc, userID := newTestImportService(t)

	data := []byte(`{
		"profile": {"display_name": "No Slug Profile"},
		"links": []
	}`)

	result, err := svc.importFromJSON(userID, data)
	if err != nil {
		t.Fatalf("importFromJSON (no slug): %v", err)
	}
	if _, ok := result["profile_id"]; !ok {
		t.Error("result missing profile_id")
	}
}

// ---------------------------------------------------------------------------
// importFromCSV (direct test)
// ---------------------------------------------------------------------------

func TestImportService_ImportFromCSV(t *testing.T) {
	svc, userID := newTestImportService(t)

	csvData := []byte("title,url\nMyBlog,https://blog.example.com\nPortfolio,https://portfolio.example.com")

	result, err := svc.importFromCSV(userID, csvData)
	if err != nil {
		t.Fatalf("importFromCSV: %v", err)
	}
	if result == nil {
		t.Fatal("importFromCSV returned nil")
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 2 {
		t.Errorf("links_imported = %v, want 2", result["links_imported"])
	}
}

// ---------------------------------------------------------------------------
// importFromLinktree (direct test)
// ---------------------------------------------------------------------------

func TestImportService_ImportFromLinktree(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"accountData": map[string]interface{}{
			"username":          "ltuser",
			"displayName":       "Linktree User",
			"bio":               "Bio text",
			"profilePictureUrl": "",
		},
		"links": []map[string]interface{}{
			{"title": "Site", "url": "https://site.example.com"},
		},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromLinktree(userID, data)
	if err != nil {
		t.Fatalf("importFromLinktree: %v", err)
	}
	if result == nil {
		t.Fatal("importFromLinktree returned nil")
	}
	if _, ok := result["profile_id"]; !ok {
		t.Error("result missing profile_id")
	}
}

// ---------------------------------------------------------------------------
// generateID (import service)
// ---------------------------------------------------------------------------

func TestImportService_GenerateID(t *testing.T) {
	db, _ := newTestImportDB(t)
	ps := NewProfileService(db)
	ls := NewLinkService(db)
	is := NewImportService(db, ps, ls)

	id := is.generateID()
	if id == "" {
		t.Error("generateID returned empty string")
	}
	if len(id) < 5 {
		t.Errorf("generateID = %q seems too short", id)
	}
}
