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

// ---------------------------------------------------------------------------
// importFromLinkstack (direct test)
// ---------------------------------------------------------------------------

func TestImportService_ImportFromLinkstack(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"profile": map[string]interface{}{
			"username": "lsuser",
			"name":     "Linkstack User",
			"bio":      "Bio from linkstack",
			"avatar":   "https://example.com/avatar.png",
		},
		"links": []map[string]interface{}{
			{"title": "Website", "url": "https://site.example.com", "order": 1},
			{"title": "Blog", "url": "https://blog.example.com", "order": 2},
		},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromLinkstack(userID, data)
	if err != nil {
		t.Fatalf("importFromLinkstack: %v", err)
	}
	if result == nil {
		t.Fatal("importFromLinkstack returned nil result")
	}
	if _, ok := result["profile_id"]; !ok {
		t.Error("result missing profile_id key")
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 2 {
		t.Errorf("links_imported = %v, want 2", result["links_imported"])
	}
	totalLinks, ok := result["total_links"].(int)
	if !ok || totalLinks != 2 {
		t.Errorf("total_links = %v, want 2", result["total_links"])
	}
}

func TestImportService_ImportFromLinkstack_InvalidJSON(t *testing.T) {
	svc, userID := newTestImportService(t)

	_, err := svc.importFromLinkstack(userID, []byte(`{not valid}`))
	if err == nil {
		t.Error("importFromLinkstack with invalid JSON should return error")
	}
}

func TestImportService_ImportFromLinkstack_NoLinks(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"profile": map[string]interface{}{
			"username": "lsnolinks",
			"name":     "No Links User",
			"bio":      "",
			"avatar":   "",
		},
		"links": []interface{}{},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromLinkstack(userID, data)
	if err != nil {
		t.Fatalf("importFromLinkstack (no links): %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 0 {
		t.Errorf("links_imported = %v, want 0", result["links_imported"])
	}
}

func TestImportService_ImportData_Linkstack(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"profile": map[string]interface{}{
			"username": "lspublicuser",
			"name":     "Public User",
			"bio":      "Test",
			"avatar":   "",
		},
		"links": []map[string]interface{}{
			{"title": "Link", "url": "https://example.com", "order": 1},
		},
	}

	data, _ := json.Marshal(input)
	jobID, err := svc.ImportData(userID, "linkstack", data)
	if err != nil {
		t.Fatalf("ImportData(linkstack): %v", err)
	}
	if jobID == "" {
		t.Error("ImportData(linkstack) returned empty jobID")
	}
}

// ---------------------------------------------------------------------------
// importFromCarrd (direct test)
// ---------------------------------------------------------------------------

func TestImportService_ImportFromCarrd(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"title": "My Carrd Site",
		"bio":   "Short bio from carrd",
		"links": []string{
			"https://twitter.com/user",
			"https://github.com/user",
		},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromCarrd(userID, data)
	if err != nil {
		t.Fatalf("importFromCarrd: %v", err)
	}
	if result == nil {
		t.Fatal("importFromCarrd returned nil result")
	}
	if _, ok := result["profile_id"]; !ok {
		t.Error("result missing profile_id key")
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 2 {
		t.Errorf("links_imported = %v, want 2", result["links_imported"])
	}
}

func TestImportService_ImportFromCarrd_InvalidJSON(t *testing.T) {
	svc, userID := newTestImportService(t)

	_, err := svc.importFromCarrd(userID, []byte(`{bad json`))
	if err == nil {
		t.Error("importFromCarrd with invalid JSON should return error")
	}
}

func TestImportService_ImportFromCarrd_NoLinks(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"title": "Empty Carrd",
		"bio":   "No links here",
		"links": []string{},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromCarrd(userID, data)
	if err != nil {
		t.Fatalf("importFromCarrd (no links): %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 0 {
		t.Errorf("links_imported = %v, want 0", result["links_imported"])
	}
}

func TestImportService_ImportData_Carrd(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"title": "Carrd Public Test",
		"bio":   "Testing carrd import",
		"links": []string{"https://example.com"},
	}

	data, _ := json.Marshal(input)
	jobID, err := svc.ImportData(userID, "carrd", data)
	if err != nil {
		t.Fatalf("ImportData(carrd): %v", err)
	}
	if jobID == "" {
		t.Error("ImportData(carrd) returned empty jobID")
	}
}

// ---------------------------------------------------------------------------
// importFromAboutMe (direct test)
// ---------------------------------------------------------------------------

func TestImportService_ImportFromAboutMe(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"username": "aboutmeuser",
		"name":     "About Me User",
		"headline": "Professional headline",
		"avatar":   "https://example.com/avatar.png",
		"links": []map[string]interface{}{
			{"label": "LinkedIn", "url": "https://linkedin.com/in/user"},
			{"label": "Twitter", "url": "https://twitter.com/user"},
		},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromAboutMe(userID, data)
	if err != nil {
		t.Fatalf("importFromAboutMe: %v", err)
	}
	if result == nil {
		t.Fatal("importFromAboutMe returned nil result")
	}
	if _, ok := result["profile_id"]; !ok {
		t.Error("result missing profile_id key")
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 2 {
		t.Errorf("links_imported = %v, want 2", result["links_imported"])
	}
	totalLinks, ok := result["total_links"].(int)
	if !ok || totalLinks != 2 {
		t.Errorf("total_links = %v, want 2", result["total_links"])
	}
}

func TestImportService_ImportFromAboutMe_InvalidJSON(t *testing.T) {
	svc, userID := newTestImportService(t)

	_, err := svc.importFromAboutMe(userID, []byte(`[not an object]`))
	if err == nil {
		t.Error("importFromAboutMe with invalid JSON should return error")
	}
}

func TestImportService_ImportFromAboutMe_NoLinks(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"username": "aboutmequiet",
		"name":     "Quiet User",
		"headline": "",
		"avatar":   "",
		"links":    []interface{}{},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromAboutMe(userID, data)
	if err != nil {
		t.Fatalf("importFromAboutMe (no links): %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 0 {
		t.Errorf("links_imported = %v, want 0", result["links_imported"])
	}
}

func TestImportService_ImportData_AboutMe(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"username": "aboutmepublic",
		"name":     "Public About Me",
		"headline": "A headline",
		"avatar":   "",
		"links": []map[string]interface{}{
			{"label": "Website", "url": "https://example.com"},
		},
	}

	data, _ := json.Marshal(input)
	jobID, err := svc.ImportData(userID, "aboutme", data)
	if err != nil {
		t.Fatalf("ImportData(aboutme): %v", err)
	}
	if jobID == "" {
		t.Error("ImportData(aboutme) returned empty jobID")
	}
}

// ---------------------------------------------------------------------------
// importFromLinktree — error paths
// ---------------------------------------------------------------------------

func TestImportService_ImportFromLinktree_InvalidJSON(t *testing.T) {
	svc, userID := newTestImportService(t)

	_, err := svc.importFromLinktree(userID, []byte(`{not valid json`))
	if err == nil {
		t.Error("importFromLinktree with invalid JSON should return error")
	}
}

func TestImportService_ImportFromLinktree_NoLinks(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"accountData": map[string]interface{}{
			"username":          "ltnolinks",
			"displayName":       "No Links",
			"bio":               "",
			"profilePictureUrl": "",
		},
		"links": []interface{}{},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromLinktree(userID, data)
	if err != nil {
		t.Fatalf("importFromLinktree (no links): %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 0 {
		t.Errorf("links_imported = %v, want 0", result["links_imported"])
	}
}

// ---------------------------------------------------------------------------
// importFromAboutMe — InvalidJSON already tested; exercise full path with links
// ---------------------------------------------------------------------------

func TestImportService_ImportFromAboutMe_WithMultipleLinks(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"username": "aboutmemulti",
		"name":     "Multi Link User",
		"headline": "Three links",
		"avatar":   "https://example.com/avatar.png",
		"links": []map[string]interface{}{
			{"label": "GitHub", "url": "https://github.com/user"},
			{"label": "Twitter", "url": "https://twitter.com/user"},
			{"label": "Blog", "url": "https://blog.example.com"},
		},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromAboutMe(userID, data)
	if err != nil {
		t.Fatalf("importFromAboutMe (multi links): %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 3 {
		t.Errorf("links_imported = %v, want 3", result["links_imported"])
	}
}

// ---------------------------------------------------------------------------
// importFromCSV — additional edge cases
// ---------------------------------------------------------------------------

func TestImportService_ImportFromCSV_WithUsername(t *testing.T) {
	svc, userID := newTestImportService(t)

	csvData := []byte("title,url,username\nGitHub,https://github.com/user,ghuser\nBlog,https://blog.example.com,")

	result, err := svc.importFromCSV(userID, csvData)
	if err != nil {
		t.Fatalf("importFromCSV with username: %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 2 {
		t.Errorf("links_imported = %v, want 2", result["links_imported"])
	}
}

func TestImportService_ImportFromCSV_EmptyRows(t *testing.T) {
	svc, userID := newTestImportService(t)

	// Header only, no data rows.
	csvData := []byte("title,url\n")

	result, err := svc.importFromCSV(userID, csvData)
	if err != nil {
		t.Fatalf("importFromCSV empty rows: %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 0 {
		t.Errorf("links_imported = %v, want 0", result["links_imported"])
	}
}

func TestImportService_ImportFromCSV_MissingHeader(t *testing.T) {
	svc, userID := newTestImportService(t)

	// Empty data — reader.Read() will return EOF.
	_, err := svc.importFromCSV(userID, []byte(""))
	if err == nil {
		t.Error("importFromCSV with empty data should return error")
	}
}

// ---------------------------------------------------------------------------
// importFromJSON — error paths
// ---------------------------------------------------------------------------

func TestImportService_ImportFromJSON_InvalidJSON(t *testing.T) {
	svc, userID := newTestImportService(t)

	_, err := svc.importFromJSON(userID, []byte(`{bad json`))
	if err == nil {
		t.Error("importFromJSON with invalid JSON should return error")
	}
}

func TestImportService_ImportFromJSON_WithLinks(t *testing.T) {
	svc, userID := newTestImportService(t)

	data := []byte(`{
		"profile": {
			"slug": "json-with-links",
			"display_name": "Link User",
			"bio": "Bio",
			"avatar_url": "https://example.com/avatar.png"
		},
		"links": [
			{"title": "GitHub", "url": "https://github.com/u", "username": "ghuser"},
			{"title": "Twitter", "url": "https://twitter.com/u"}
		]
	}`)

	result, err := svc.importFromJSON(userID, data)
	if err != nil {
		t.Fatalf("importFromJSON with links: %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 2 {
		t.Errorf("links_imported = %v, want 2", result["links_imported"])
	}
	totalLinks, ok := result["total_links"].(int)
	if !ok || totalLinks != 2 {
		t.Errorf("total_links = %v, want 2", result["total_links"])
	}
}

// ---------------------------------------------------------------------------
// ImportData — default case (already tested via ErrUnsupportedImportSource)
// and the "failed to update job result" path covered indirectly
// ---------------------------------------------------------------------------

func TestImportService_ImportData_LinktreeInvalidData(t *testing.T) {
	svc, userID := newTestImportService(t)

	// Invalid JSON for linktree source — ImportData will record the failed job
	// and return the import error.
	jobID, err := svc.ImportData(userID, "linktree", []byte(`{invalid}`))
	if err == nil {
		t.Error("ImportData linktree invalid data should return error")
	}
	// A job ID is still returned even on failure.
	if jobID == "" {
		t.Error("ImportData should return a jobID even on failure")
	}
}

func TestImportService_ImportData_CarrdInvalidData(t *testing.T) {
	svc, userID := newTestImportService(t)

	jobID, err := svc.ImportData(userID, "carrd", []byte(`{invalid json`))
	if err == nil {
		t.Error("ImportData carrd invalid data should return error")
	}
	if jobID == "" {
		t.Error("ImportData should return a jobID even on failure")
	}
}

func TestImportService_ImportData_AboutMeInvalidData(t *testing.T) {
	svc, userID := newTestImportService(t)

	jobID, err := svc.ImportData(userID, "aboutme", []byte(`[not object]`))
	if err == nil {
		t.Error("ImportData aboutme invalid data should return error")
	}
	if jobID == "" {
		t.Error("ImportData should return a jobID even on failure")
	}
}

func TestImportService_ImportData_LinkstackInvalidData(t *testing.T) {
	svc, userID := newTestImportService(t)

	jobID, err := svc.ImportData(userID, "linkstack", []byte(`{invalid`))
	if err == nil {
		t.Error("ImportData linkstack invalid data should return error")
	}
	if jobID == "" {
		t.Error("ImportData should return a jobID even on failure")
	}
}

// ---------------------------------------------------------------------------
// linkService.Create failure — invalid URL triggers the "continue" path
// ---------------------------------------------------------------------------

func TestImportService_ImportFromLinktree_InvalidLinkURL(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"accountData": map[string]interface{}{
			"username":          "ltinvalidurl",
			"displayName":       "Invalid URL User",
			"bio":               "",
			"profilePictureUrl": "",
		},
		"links": []map[string]interface{}{
			{"title": "Bad Link", "url": "not-a-valid-url"},
			{"title": "Good Link", "url": "https://good.example.com"},
		},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromLinktree(userID, data)
	if err != nil {
		t.Fatalf("importFromLinktree with invalid URL: %v", err)
	}
	// One link skipped (invalid URL), one imported.
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 1 {
		t.Errorf("links_imported = %v, want 1 (bad URL should be skipped)", result["links_imported"])
	}
}

func TestImportService_ImportFromLinkstack_InvalidLinkURL(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"profile": map[string]interface{}{
			"username": "lsinvalidurl",
			"name":     "Invalid URL Linkstack",
			"bio":      "",
			"avatar":   "",
		},
		"links": []map[string]interface{}{
			{"title": "Bad", "url": "not-valid", "order": 1},
			{"title": "Good", "url": "https://valid.example.com", "order": 2},
		},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromLinkstack(userID, data)
	if err != nil {
		t.Fatalf("importFromLinkstack with invalid URL: %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 1 {
		t.Errorf("links_imported = %v, want 1", result["links_imported"])
	}
}

func TestImportService_ImportFromCarrd_InvalidLinkURL(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"title": "Carrd Invalid URL Test",
		"bio":   "testing",
		"links": []string{"not-valid-url", "https://valid.example.com"},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromCarrd(userID, data)
	if err != nil {
		t.Fatalf("importFromCarrd with invalid URL: %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 1 {
		t.Errorf("links_imported = %v, want 1", result["links_imported"])
	}
}

func TestImportService_ImportFromAboutMe_InvalidLinkURL(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"username": "aboutmeinvalidurl",
		"name":     "Invalid URL About Me",
		"headline": "",
		"avatar":   "",
		"links": []map[string]interface{}{
			{"label": "Bad", "url": "not-a-url"},
			{"label": "Good", "url": "https://good.example.com"},
		},
	}

	data, _ := json.Marshal(input)
	result, err := svc.importFromAboutMe(userID, data)
	if err != nil {
		t.Fatalf("importFromAboutMe with invalid URL: %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 1 {
		t.Errorf("links_imported = %v, want 1", result["links_imported"])
	}
}

func TestImportService_ImportFromJSON_InvalidLinkURL(t *testing.T) {
	svc, userID := newTestImportService(t)

	data := []byte(`{
		"profile": {"slug": "json-invalid-url", "display_name": "Invalid URL"},
		"links": [
			{"title": "Bad", "url": "not-valid"},
			{"title": "Good", "url": "https://valid.example.com"}
		]
	}`)

	result, err := svc.importFromJSON(userID, data)
	if err != nil {
		t.Fatalf("importFromJSON with invalid URL: %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 1 {
		t.Errorf("links_imported = %v, want 1", result["links_imported"])
	}
}

// ---------------------------------------------------------------------------
// importFromCSV — record with fewer than 2 columns (len(record) < 2 path)
// ---------------------------------------------------------------------------

func TestImportService_ImportFromCSV_ShortRecord(t *testing.T) {
	svc, userID := newTestImportService(t)

	// One good row, one short row (only 1 column), one more good row.
	csvData := []byte("title,url\nGoodLink,https://good.example.com\nbadrow\nGoodLink2,https://good2.example.com")

	result, err := svc.importFromCSV(userID, csvData)
	if err != nil {
		t.Fatalf("importFromCSV short record: %v", err)
	}
	// The short row should be skipped, 2 good rows imported.
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 2 {
		t.Errorf("links_imported = %v, want 2 (short row skipped)", result["links_imported"])
	}
}

func TestImportService_ImportFromCSV_InvalidLinkURL(t *testing.T) {
	svc, userID := newTestImportService(t)

	csvData := []byte("title,url\nBadLink,not-a-valid-url\nGoodLink,https://good.example.com")

	result, err := svc.importFromCSV(userID, csvData)
	if err != nil {
		t.Fatalf("importFromCSV invalid URL: %v", err)
	}
	linksImported, ok := result["links_imported"].(int)
	if !ok || linksImported != 1 {
		t.Errorf("links_imported = %v, want 1 (invalid URL skipped)", result["links_imported"])
	}
}

// ---------------------------------------------------------------------------
// profileService.Create failure — duplicate slug triggers error path
// ---------------------------------------------------------------------------

func TestImportService_ImportFromLinktree_DuplicateSlug(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"accountData": map[string]interface{}{
			"username":          "duplinktree",
			"displayName":       "Dup Linktree",
			"bio":               "",
			"profilePictureUrl": "",
		},
		"links": []interface{}{},
	}
	data, _ := json.Marshal(input)

	// First import succeeds.
	if _, err := svc.importFromLinktree(userID, data); err != nil {
		t.Fatalf("first import: %v", err)
	}

	// Second import with same slug should fail (duplicate slug).
	_, err := svc.importFromLinktree(userID, data)
	if err == nil {
		t.Error("importFromLinktree with duplicate slug should return error")
	}
}

func TestImportService_ImportFromLinkstack_DuplicateSlug(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"profile": map[string]interface{}{
			"username": "duplinkstack",
			"name":     "Dup Linkstack",
			"bio":      "",
			"avatar":   "",
		},
		"links": []interface{}{},
	}
	data, _ := json.Marshal(input)

	if _, err := svc.importFromLinkstack(userID, data); err != nil {
		t.Fatalf("first import: %v", err)
	}
	_, err := svc.importFromLinkstack(userID, data)
	if err == nil {
		t.Error("importFromLinkstack with duplicate slug should return error")
	}
}

func TestImportService_ImportFromCarrd_DuplicateSlug(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"title": "Dup Carrd",
		"bio":   "",
		"links": []string{},
	}
	data, _ := json.Marshal(input)

	if _, err := svc.importFromCarrd(userID, data); err != nil {
		t.Fatalf("first import: %v", err)
	}
	_, err := svc.importFromCarrd(userID, data)
	if err == nil {
		t.Error("importFromCarrd with duplicate slug should return error")
	}
}

func TestImportService_ImportFromAboutMe_DuplicateSlug(t *testing.T) {
	svc, userID := newTestImportService(t)

	input := map[string]interface{}{
		"username": "dupaboutme",
		"name":     "Dup About Me",
		"headline": "",
		"avatar":   "",
		"links":    []interface{}{},
	}
	data, _ := json.Marshal(input)

	if _, err := svc.importFromAboutMe(userID, data); err != nil {
		t.Fatalf("first import: %v", err)
	}
	_, err := svc.importFromAboutMe(userID, data)
	if err == nil {
		t.Error("importFromAboutMe with duplicate slug should return error")
	}
}

func TestImportService_ImportFromJSON_DuplicateSlug(t *testing.T) {
	svc, userID := newTestImportService(t)

	data := []byte(`{
		"profile": {"slug": "dup-json-slug", "display_name": "Dup JSON"},
		"links": []
	}`)

	if _, err := svc.importFromJSON(userID, data); err != nil {
		t.Fatalf("first import: %v", err)
	}
	_, err := svc.importFromJSON(userID, data)
	if err == nil {
		t.Error("importFromJSON with duplicate slug should return error")
	}
}
