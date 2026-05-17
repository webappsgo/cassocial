package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/casapps/cassocial/src/server/model"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestExportDB creates an in-memory DB with a user and profile seeded.
func newTestExportDB(t *testing.T) (*store.DB, string, string) {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	userID := "export-user-001"
	_, err = db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, created_at, updated_at)
		 VALUES (?, 'exportowner', 'export@example.com',
		         '$argon2id$v=19$m=65536,t=3,p=4$s$h', 'user', 'active', 1, 0,
		         CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		userID,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if err := db.CreateProfile(&store.Profile{
		ID:       "export-profile-001",
		UserID:   userID,
		Slug:     "export-test-profile",
		IsPublic: true,
	}); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	return db, userID, "export-profile-001"
}

// newExportServiceWithData returns an ExportService with one profile and two links.
func newExportServiceWithData(t *testing.T) (*ExportService, string, string) {
	t.Helper()

	db, userID, profileID := newTestExportDB(t)

	ps := NewProfileService(db)
	ls := NewLinkService(db)
	es := NewExportService(db, ps, ls)

	// Insert two links
	for i, title := range []string{"GitHub", "Twitter"} {
		l := &model.Link{
			ProfileID: profileID,
			Title:     title,
			URL:       "https://example.com/" + title,
			IsActive:  true,
			Position:  i + 1,
		}
		if err := ls.Create(l); err != nil {
			t.Fatalf("Create link %s: %v", title, err)
		}
	}

	return es, userID, profileID
}

// ---------------------------------------------------------------------------
// ExportProfile – format dispatch
// ---------------------------------------------------------------------------

func TestExportService_ExportProfile_JSON(t *testing.T) {
	es, userID, profileID := newExportServiceWithData(t)

	data, filename, err := es.ExportProfile(profileID, userID, "json")
	if err != nil {
		t.Fatalf("ExportProfile json: %v", err)
	}
	if len(data) == 0 {
		t.Error("ExportProfile json: empty data")
	}
	if !strings.HasSuffix(filename, ".json") {
		t.Errorf("ExportProfile json: filename %q should end with .json", filename)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("ExportProfile json: invalid JSON: %v", err)
	}
}

func TestExportService_ExportProfile_CSV(t *testing.T) {
	es, userID, profileID := newExportServiceWithData(t)

	data, filename, err := es.ExportProfile(profileID, userID, "csv")
	if err != nil {
		t.Fatalf("ExportProfile csv: %v", err)
	}
	if len(data) == 0 {
		t.Error("ExportProfile csv: empty data")
	}
	if !strings.HasSuffix(filename, ".csv") {
		t.Errorf("ExportProfile csv: filename %q should end with .csv", filename)
	}

	// Parse CSV to validate
	r := csv.NewReader(bytes.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		t.Errorf("ExportProfile csv: invalid CSV: %v", err)
	}
	if len(records) < 2 { // header + at least 1 row
		t.Errorf("ExportProfile csv: got %d rows, want at least 2", len(records))
	}
}

func TestExportService_ExportProfile_HTML(t *testing.T) {
	es, userID, profileID := newExportServiceWithData(t)

	data, filename, err := es.ExportProfile(profileID, userID, "html")
	if err != nil {
		t.Fatalf("ExportProfile html: %v", err)
	}
	if len(data) == 0 {
		t.Error("ExportProfile html: empty data")
	}
	if !strings.HasSuffix(filename, ".html") {
		t.Errorf("ExportProfile html: filename %q should end with .html", filename)
	}
	if !strings.Contains(string(data), "<!DOCTYPE html>") {
		t.Error("ExportProfile html: output missing DOCTYPE declaration")
	}
}

func TestExportService_ExportProfile_PDF(t *testing.T) {
	es, userID, profileID := newExportServiceWithData(t)

	data, filename, err := es.ExportProfile(profileID, userID, "pdf")
	if err != nil {
		t.Fatalf("ExportProfile pdf: %v", err)
	}
	if len(data) == 0 {
		t.Error("ExportProfile pdf: empty data")
	}
	if !strings.HasSuffix(filename, ".pdf") {
		t.Errorf("ExportProfile pdf: filename %q should end with .pdf", filename)
	}
}

func TestExportService_ExportProfile_VCard(t *testing.T) {
	es, userID, profileID := newExportServiceWithData(t)

	data, filename, err := es.ExportProfile(profileID, userID, "vcard")
	if err != nil {
		t.Fatalf("ExportProfile vcard: %v", err)
	}
	if len(data) == 0 {
		t.Error("ExportProfile vcard: empty data")
	}
	if !strings.HasSuffix(filename, ".vcf") {
		t.Errorf("ExportProfile vcard: filename %q should end with .vcf", filename)
	}
	if !strings.Contains(string(data), "BEGIN:VCARD") {
		t.Error("ExportProfile vcard: output missing BEGIN:VCARD")
	}
}

func TestExportService_ExportProfile_UnsupportedFormat(t *testing.T) {
	es, userID, profileID := newExportServiceWithData(t)

	_, _, err := es.ExportProfile(profileID, userID, "xlsx")
	if err != ErrUnsupportedExportFormat {
		t.Errorf("ExportProfile unsupported = %v, want ErrUnsupportedExportFormat", err)
	}
}

func TestExportService_ExportProfile_WrongOwner(t *testing.T) {
	es, _, profileID := newExportServiceWithData(t)

	_, _, err := es.ExportProfile(profileID, "wrong-user-id", "json")
	if err != ErrProfileAccessDenied {
		t.Errorf("ExportProfile wrong owner = %v, want ErrProfileAccessDenied", err)
	}
}

func TestExportService_ExportProfile_ProfileNotFound(t *testing.T) {
	es, userID, _ := newExportServiceWithData(t)

	_, _, err := es.ExportProfile("no-such-profile", userID, "json")
	if err == nil {
		t.Error("ExportProfile nonexistent profile should return error")
	}
}

// ---------------------------------------------------------------------------
// exportToCSV (direct)
// ---------------------------------------------------------------------------

func TestExportService_ExportToCSV_ActiveLinks(t *testing.T) {
	es, _, _ := newExportServiceWithData(t)

	profile := &model.Profile{
		ID:          "csv-test-id",
		Slug:        "csv-test",
		DisplayName: "CSV Test",
	}
	links := []*model.Link{
		{Title: "Active", URL: "https://active.com", IsActive: true, Position: 1},
		{Title: "Inactive", URL: "https://inactive.com", IsActive: false, Position: 2},
	}

	data, filename, err := es.exportToCSV(profile, links)
	if err != nil {
		t.Fatalf("exportToCSV: %v", err)
	}
	if !strings.HasSuffix(filename, ".csv") {
		t.Errorf("filename %q should end with .csv", filename)
	}

	r := csv.NewReader(bytes.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	// Header + 2 link rows (both active and inactive are exported)
	if len(records) != 3 {
		t.Errorf("CSV rows = %d, want 3 (1 header + 2 links)", len(records))
	}
}

// ---------------------------------------------------------------------------
// exportToHTML (direct)
// ---------------------------------------------------------------------------

func TestExportService_ExportToHTML_ContainsLinks(t *testing.T) {
	es, _, _ := newExportServiceWithData(t)

	profile := &model.Profile{
		Slug:        "html-test",
		DisplayName: "HTML Test User",
		Bio:         "My bio",
	}
	links := []*model.Link{
		{Title: "Site", URL: "https://site.com", IsActive: true},
		{Title: "Blog", URL: "https://blog.com", IsActive: false},
	}

	data, _, err := es.exportToHTML(profile, links)
	if err != nil {
		t.Fatalf("exportToHTML: %v", err)
	}

	htmlStr := string(data)
	if !strings.Contains(htmlStr, "HTML Test User") {
		t.Error("HTML missing display name")
	}
	if !strings.Contains(htmlStr, "https://site.com") {
		t.Error("HTML missing active link URL")
	}
	// Inactive link should not appear
	if strings.Contains(htmlStr, "https://blog.com") {
		t.Error("HTML should not contain inactive link")
	}
}

// ---------------------------------------------------------------------------
// exportToVCard (direct)
// ---------------------------------------------------------------------------

func TestExportService_ExportToVCard(t *testing.T) {
	es, _, _ := newExportServiceWithData(t)

	profile := &model.Profile{
		Slug:        "vcard-test",
		DisplayName: "VCard Test",
		Bio:         "Bio text",
		AvatarURL:   "https://avatar.example.com/pic.jpg",
	}
	links := []*model.Link{
		{Title: "GitHub", URL: "https://github.com/test", IsActive: true},
	}

	data, filename, err := es.exportToVCard(profile, links)
	if err != nil {
		t.Fatalf("exportToVCard: %v", err)
	}
	if !strings.HasSuffix(filename, ".vcf") {
		t.Errorf("vcard filename %q should end with .vcf", filename)
	}

	content := string(data)
	if !strings.Contains(content, "BEGIN:VCARD") {
		t.Error("vcard missing BEGIN:VCARD")
	}
	if !strings.Contains(content, "END:VCARD") {
		t.Error("vcard missing END:VCARD")
	}
	if !strings.Contains(content, "VCard Test") {
		t.Error("vcard missing display name")
	}
}

// ---------------------------------------------------------------------------
// ExportAnalytics
// ---------------------------------------------------------------------------

func TestExportService_ExportAnalytics_JSON(t *testing.T) {
	db, userID, profileID := newTestExportDB(t)
	ps := NewProfileService(db)
	ls := NewLinkService(db)
	es := NewExportService(db, ps, ls)

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(time.Hour)

	data, filename, err := es.ExportAnalytics(profileID, userID, start, end, "json")
	if err != nil {
		t.Fatalf("ExportAnalytics json: %v", err)
	}
	if len(data) == 0 {
		t.Error("ExportAnalytics json: empty data")
	}
	if !strings.HasSuffix(filename, ".json") {
		t.Errorf("filename %q should end with .json", filename)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("ExportAnalytics json: invalid JSON: %v", err)
	}
}

func TestExportService_ExportAnalytics_CSV(t *testing.T) {
	db, userID, profileID := newTestExportDB(t)
	ps := NewProfileService(db)
	ls := NewLinkService(db)
	es := NewExportService(db, ps, ls)

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(time.Hour)

	data, filename, err := es.ExportAnalytics(profileID, userID, start, end, "csv")
	if err != nil {
		t.Fatalf("ExportAnalytics csv: %v", err)
	}
	if !strings.HasSuffix(filename, ".csv") {
		t.Errorf("filename %q should end with .csv", filename)
	}
	_ = data
}

func TestExportService_ExportAnalytics_UnsupportedFormat(t *testing.T) {
	db, userID, profileID := newTestExportDB(t)
	ps := NewProfileService(db)
	ls := NewLinkService(db)
	es := NewExportService(db, ps, ls)

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(time.Hour)

	_, _, err := es.ExportAnalytics(profileID, userID, start, end, "pdf")
	if err != ErrUnsupportedExportFormat {
		t.Errorf("ExportAnalytics unsupported = %v, want ErrUnsupportedExportFormat", err)
	}
}

func TestExportService_ExportAnalytics_WrongOwner(t *testing.T) {
	db, _, profileID := newTestExportDB(t)
	ps := NewProfileService(db)
	ls := NewLinkService(db)
	es := NewExportService(db, ps, ls)

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(time.Hour)

	_, _, err := es.ExportAnalytics(profileID, "wrong-user", start, end, "json")
	if err != ErrProfileAccessDenied {
		t.Errorf("ExportAnalytics wrong owner = %v, want ErrProfileAccessDenied", err)
	}
}

func TestExportService_ExportAnalytics_ProfileNotFound(t *testing.T) {
	db, userID, _ := newTestExportDB(t)
	ps := NewProfileService(db)
	ls := NewLinkService(db)
	es := NewExportService(db, ps, ls)

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(time.Hour)

	_, _, err := es.ExportAnalytics("no-such-profile", userID, start, end, "json")
	if err == nil {
		t.Error("ExportAnalytics with nonexistent profile should return error")
	}
}

// ---------------------------------------------------------------------------
// exportAnalyticsJSON / exportAnalyticsCSV — with actual records
// ---------------------------------------------------------------------------

func TestExportService_ExportAnalyticsJSON_WithRecords(t *testing.T) {
	db, userID, profileID := newTestExportDB(t)

	// Insert analytics events for the profile.
	_, err := db.Exec(
		`INSERT INTO analytics (id, profile_id, event_type, created_at) VALUES
		 ('a1', ?, 'view', CURRENT_TIMESTAMP),
		 ('a2', ?, 'view', CURRENT_TIMESTAMP),
		 ('a3', ?, 'click', CURRENT_TIMESTAMP)`,
		profileID, profileID, profileID,
	)
	if err != nil {
		t.Fatalf("insert analytics: %v", err)
	}

	ps := NewProfileService(db)
	ls := NewLinkService(db)
	es := NewExportService(db, ps, ls)

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(time.Hour)

	data, filename, err := es.ExportAnalytics(profileID, userID, start, end, "json")
	if err != nil {
		t.Fatalf("ExportAnalytics json with records: %v", err)
	}
	if len(data) == 0 {
		t.Error("ExportAnalytics json: empty data")
	}
	if !strings.HasSuffix(filename, ".json") {
		t.Errorf("filename %q should end with .json", filename)
	}
}

func TestExportService_ExportAnalyticsCSV_WithRecords(t *testing.T) {
	db, userID, profileID := newTestExportDB(t)

	_, err := db.Exec(
		`INSERT INTO analytics (id, profile_id, event_type, created_at) VALUES
		 ('b1', ?, 'view', CURRENT_TIMESTAMP),
		 ('b2', ?, 'click', CURRENT_TIMESTAMP)`,
		profileID, profileID,
	)
	if err != nil {
		t.Fatalf("insert analytics: %v", err)
	}

	ps := NewProfileService(db)
	ls := NewLinkService(db)
	es := NewExportService(db, ps, ls)

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(time.Hour)

	data, filename, err := es.ExportAnalytics(profileID, userID, start, end, "csv")
	if err != nil {
		t.Fatalf("ExportAnalytics csv with records: %v", err)
	}
	if !strings.HasSuffix(filename, ".csv") {
		t.Errorf("filename %q should end with .csv", filename)
	}
	_ = data
}

// ---------------------------------------------------------------------------
// exportToCSV — additional edge cases
// ---------------------------------------------------------------------------

func TestExportService_ExportToCSV_EmptyLinks(t *testing.T) {
	es, _, _ := newExportServiceWithData(t)

	profile := &model.Profile{
		ID:   "empty-links-csv",
		Slug: "empty-csv",
	}

	data, filename, err := es.exportToCSV(profile, []*model.Link{})
	if err != nil {
		t.Fatalf("exportToCSV empty links: %v", err)
	}
	if !strings.HasSuffix(filename, ".csv") {
		t.Errorf("filename %q should end with .csv", filename)
	}
	// Should at least contain the header row.
	if !strings.Contains(string(data), "Title") {
		t.Error("CSV should contain header row")
	}
}

// ---------------------------------------------------------------------------
// exportToJSON — profile with all optional fields set
// ---------------------------------------------------------------------------

func TestExportService_ExportToJSON_FullProfile(t *testing.T) {
	es, _, _ := newExportServiceWithData(t)

	profile := &model.Profile{
		ID:               "full-profile-id",
		Slug:             "full-profile",
		DisplayName:      "Full Profile",
		Bio:              "Bio text",
		AvatarURL:        "https://example.com/avatar.png",
		HeaderImageURL:   "https://example.com/header.png",
		MetaTitle:        "Meta Title",
		MetaDescription:  "Meta description here",
		ShowUsernames:    true,
		IsPublic:         true,
		AnalyticsEnabled: true,
	}
	links := []*model.Link{
		{Title: "Link1", URL: "https://link1.example.com", IsActive: true, Position: 1, ClickCount: 42},
		{Title: "Link2", URL: "https://link2.example.com", IsActive: false, Position: 2, Username: "user2"},
	}

	data, filename, err := es.exportToJSON(profile, links)
	if err != nil {
		t.Fatalf("exportToJSON full profile: %v", err)
	}
	if !strings.HasSuffix(filename, ".json") {
		t.Errorf("filename %q should end with .json", filename)
	}
	if !strings.Contains(string(data), "full-profile") {
		t.Error("JSON should contain profile slug")
	}
	if !strings.Contains(string(data), "1.0.0") {
		t.Error("JSON should contain version field")
	}
}

// ---------------------------------------------------------------------------
// exportToHTML — profile with avatar and meta description
// ---------------------------------------------------------------------------

func TestExportService_ExportToHTML_WithAvatarAndMeta(t *testing.T) {
	es, _, _ := newExportServiceWithData(t)

	profile := &model.Profile{
		Slug:            "avatar-test",
		DisplayName:     "Avatar User",
		Bio:             "Has avatar",
		AvatarURL:       "https://example.com/avatar.png",
		MetaDescription: "A test profile with avatar",
	}
	links := []*model.Link{
		{Title: "Active Link", URL: "https://active.example.com", IsActive: true},
	}

	data, filename, err := es.exportToHTML(profile, links)
	if err != nil {
		t.Fatalf("exportToHTML with avatar: %v", err)
	}
	if !strings.HasSuffix(filename, ".html") {
		t.Errorf("filename %q should end with .html", filename)
	}
	htmlStr := string(data)
	if !strings.Contains(htmlStr, "avatar.png") {
		t.Error("HTML should contain avatar URL")
	}
	if !strings.Contains(htmlStr, "A test profile with avatar") {
		t.Error("HTML should contain meta description")
	}
}

// ---------------------------------------------------------------------------
// exportToPDF — direct test
// ---------------------------------------------------------------------------

func TestExportService_ExportToPDF_Direct(t *testing.T) {
	es, _, _ := newExportServiceWithData(t)

	profile := &model.Profile{
		Slug:        "pdf-direct",
		DisplayName: "PDF User",
		Bio:         "PDF bio",
	}
	links := []*model.Link{
		{Title: "Link", URL: "https://pdf.example.com", IsActive: true},
		{Title: "Inactive", URL: "https://inactive.example.com", IsActive: false},
	}

	data, filename, err := es.exportToPDF(profile, links)
	if err != nil {
		t.Fatalf("exportToPDF: %v", err)
	}
	if !strings.HasSuffix(filename, ".pdf") {
		t.Errorf("filename %q should end with .pdf", filename)
	}
	if !strings.Contains(string(data), "%PDF") {
		t.Error("PDF output should start with %PDF")
	}
}

// ---------------------------------------------------------------------------
// exportToVCard — profile with no bio, no avatar
// ---------------------------------------------------------------------------

func TestExportService_ExportToVCard_MinimalProfile(t *testing.T) {
	es, _, _ := newExportServiceWithData(t)

	profile := &model.Profile{
		Slug:        "minimal-vcard",
		DisplayName: "Minimal",
	}

	data, filename, err := es.exportToVCard(profile, []*model.Link{})
	if err != nil {
		t.Fatalf("exportToVCard minimal: %v", err)
	}
	if !strings.HasSuffix(filename, ".vcf") {
		t.Errorf("filename %q should end with .vcf", filename)
	}
	content := string(data)
	if !strings.Contains(content, "BEGIN:VCARD") {
		t.Error("vcard missing BEGIN:VCARD")
	}
	if !strings.Contains(content, "END:VCARD") {
		t.Error("vcard missing END:VCARD")
	}
}

// ---------------------------------------------------------------------------
// ExportProfile — get links error (linkService for deleted profile data)
// ---------------------------------------------------------------------------

func TestExportService_ExportProfile_AllFormats_NoLinks(t *testing.T) {
	es, userID, profileID := newExportServiceWithData(t)

	for _, format := range []string{"json", "csv", "html", "pdf", "vcard"} {
		data, filename, err := es.ExportProfile(profileID, userID, format)
		if err != nil {
			t.Errorf("ExportProfile format=%s: unexpected error: %v", format, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("ExportProfile format=%s: empty data", format)
		}
		if filename == "" {
			t.Errorf("ExportProfile format=%s: empty filename", format)
		}
	}
}
