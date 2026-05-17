package service

import (
	"testing"

	"github.com/casapps/cassocial/src/server/model"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestProfileDB creates an in-memory DB with all migrations run.
func newTestProfileDB(t *testing.T) *store.DB {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	return db
}

// insertUser inserts a bare user row needed to satisfy FK constraints.
func insertUser(t *testing.T, db *store.DB, id, username string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, '$argon2id$v=19$m=65536,t=3,p=4$s$h', 'user', 'active', 1, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		id, username, username+"@example.com",
	)
	if err != nil {
		t.Fatalf("insertUser(%s): %v", id, err)
	}
}

// newTestProfileService returns a service and a user ID ready for use.
func newTestProfileService(t *testing.T) (*ProfileService, string) {
	t.Helper()

	db := newTestProfileDB(t)
	userID := "profile-svc-user-001"
	insertUser(t, db, userID, "profilesvcowner")

	return NewProfileService(db), userID
}

// validProfile returns a minimal valid profile.
func validProfile(userID string) *model.Profile {
	return &model.Profile{
		Slug:        "test-profile",
		DisplayName: "Test Profile",
		IsPublic:    true,
	}
}

// ---------------------------------------------------------------------------
// NewProfileService
// ---------------------------------------------------------------------------

func TestNewProfileService(t *testing.T) {
	db := newTestProfileDB(t)
	svc := NewProfileService(db)
	if svc == nil {
		t.Fatal("NewProfileService returned nil")
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestProfileService_Create_Valid(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.ID == "" {
		t.Error("Create did not populate ID")
	}
}

func TestProfileService_Create_GeneratesSlug(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := &model.Profile{
		DisplayName: "My Awesome Profile",
		IsPublic:    true,
	}
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create without slug: %v", err)
	}
	if p.Slug == "" {
		t.Error("Create did not generate a slug")
	}
}

func TestProfileService_Create_DuplicateSlug(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p1 := validProfile(userID)
	if err := svc.Create(userID, p1); err != nil {
		t.Fatalf("Create p1: %v", err)
	}

	// Use same slug — test duplicate slug detection within same service
	p2b := &model.Profile{Slug: "test-profile", DisplayName: "Dup", IsPublic: true}
	if err := svc.Create(userID, p2b); err != ErrSlugAlreadyExists {
		t.Errorf("Create with duplicate slug = %v, want ErrSlugAlreadyExists", err)
	}
}

func TestProfileService_Create_InvalidSlug(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := &model.Profile{
		Slug:        "bad slug!",
		DisplayName: "Bad",
	}
	if err := svc.Create(userID, p); err != ErrInvalidSlugFormat {
		t.Errorf("Create invalid slug = %v, want ErrInvalidSlugFormat", err)
	}
}

func TestProfileService_Create_SlugTooShort(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := &model.Profile{
		Slug:        "ab",
		DisplayName: "Short",
	}
	if err := svc.Create(userID, p); err != ErrInvalidSlugFormat {
		t.Errorf("Create with 2-char slug = %v, want ErrInvalidSlugFormat", err)
	}
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestProfileService_GetByID_Found(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := svc.GetByID(p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Slug != p.Slug {
		t.Errorf("GetByID Slug = %q, want %q", got.Slug, p.Slug)
	}
}

func TestProfileService_GetByID_NotFound(t *testing.T) {
	svc, _ := newTestProfileService(t)

	_, err := svc.GetByID("no-such-profile")
	if err != ErrProfileNotFound {
		t.Errorf("GetByID(nonexistent) = %v, want ErrProfileNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// GetBySlug
// ---------------------------------------------------------------------------

func TestProfileService_GetBySlug_Found(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := svc.GetBySlug(p.Slug)
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("GetBySlug ID = %q, want %q", got.ID, p.ID)
	}
}

func TestProfileService_GetBySlug_NotFound(t *testing.T) {
	svc, _ := newTestProfileService(t)

	_, err := svc.GetBySlug("no-such-slug-xyz")
	if err != ErrProfileNotFound {
		t.Errorf("GetBySlug(nonexistent) = %v, want ErrProfileNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// GetByUserID
// ---------------------------------------------------------------------------

func TestProfileService_GetByUserID_Multiple(t *testing.T) {
	svc, userID := newTestProfileService(t)

	for i, slug := range []string{"slug-a", "slug-b", "slug-c"} {
		p := &model.Profile{Slug: slug, DisplayName: "P" + slug, IsPublic: true}
		if err := svc.Create(userID, p); err != nil {
			t.Fatalf("Create[%d]: %v", i, err)
		}
	}

	profiles, err := svc.GetByUserID(userID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if len(profiles) != 3 {
		t.Errorf("GetByUserID = %d profiles, want 3", len(profiles))
	}
}

func TestProfileService_GetByUserID_Empty(t *testing.T) {
	svc, _ := newTestProfileService(t)

	profiles, err := svc.GetByUserID("nonexistent-user")
	if err != nil {
		t.Fatalf("GetByUserID(nonexistent user): %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("GetByUserID(nonexistent) = %d, want 0", len(profiles))
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestProfileService_Update_Valid(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	p.DisplayName = "Updated Name"
	p.Bio = "Updated bio"
	if err := svc.Update(p); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := svc.GetByID(p.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.DisplayName != "Updated Name" {
		t.Errorf("after update DisplayName = %q, want 'Updated Name'", got.DisplayName)
	}
}

func TestProfileService_Update_SlugConflict(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p1 := validProfile(userID)
	if err := svc.Create(userID, p1); err != nil {
		t.Fatalf("Create p1: %v", err)
	}

	p2 := &model.Profile{Slug: "other-slug", DisplayName: "Other", IsPublic: true}
	if err := svc.Create(userID, p2); err != nil {
		t.Fatalf("Create p2: %v", err)
	}

	// Try to change p2's slug to p1's slug
	p2.Slug = p1.Slug
	if err := svc.Update(p2); err != ErrSlugAlreadyExists {
		t.Errorf("Update with taken slug = %v, want ErrSlugAlreadyExists", err)
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestProfileService_Delete_Valid(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Delete(p.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := svc.GetByID(p.ID)
	if err != ErrProfileNotFound {
		t.Errorf("GetByID after delete = %v, want ErrProfileNotFound", err)
	}
}

func TestProfileService_Delete_NotFound(t *testing.T) {
	svc, _ := newTestProfileService(t)

	if err := svc.Delete("no-such-profile"); err != ErrProfileNotFound {
		t.Errorf("Delete(nonexistent) = %v, want ErrProfileNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// Duplicate
// ---------------------------------------------------------------------------

func TestProfileService_Duplicate(t *testing.T) {
	svc, userID := newTestProfileService(t)

	orig := validProfile(userID)
	if err := svc.Create(userID, orig); err != nil {
		t.Fatalf("Create original: %v", err)
	}

	dup, err := svc.Duplicate(orig.ID, userID)
	if err != nil {
		t.Fatalf("Duplicate: %v", err)
	}

	if dup.ID == orig.ID {
		t.Error("Duplicate should have a different ID")
	}
	if dup.Slug == orig.Slug {
		t.Error("Duplicate should have a different slug")
	}
}

func TestProfileService_Duplicate_WrongUser(t *testing.T) {
	svc, userID := newTestProfileService(t)

	orig := validProfile(userID)
	if err := svc.Create(userID, orig); err != nil {
		t.Fatalf("Create original: %v", err)
	}

	_, err := svc.Duplicate(orig.ID, "wrong-user-id")
	if err == nil {
		t.Error("Duplicate by non-owner should return error")
	}
}

// ---------------------------------------------------------------------------
// IncrementViewCount
// ---------------------------------------------------------------------------

func TestProfileService_IncrementViewCount(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	for i := 0; i < 3; i++ {
		if err := svc.IncrementViewCount(p.ID); err != nil {
			t.Fatalf("IncrementViewCount[%d]: %v", i, err)
		}
	}

	got, err := svc.GetByID(p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ViewCount != 3 {
		t.Errorf("ViewCount = %d, want 3", got.ViewCount)
	}
}

// ---------------------------------------------------------------------------
// SlugExists
// ---------------------------------------------------------------------------

func TestProfileService_SlugExists(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	exists, err := svc.SlugExists(p.Slug)
	if err != nil {
		t.Fatalf("SlugExists: %v", err)
	}
	if !exists {
		t.Error("SlugExists returned false for existing slug")
	}

	exists, err = svc.SlugExists("totally-new-slug-xyz")
	if err != nil {
		t.Fatalf("SlugExists(new): %v", err)
	}
	if exists {
		t.Error("SlugExists returned true for non-existent slug")
	}
}

// ---------------------------------------------------------------------------
// CountByUserID
// ---------------------------------------------------------------------------

func TestProfileService_CountByUserID(t *testing.T) {
	svc, userID := newTestProfileService(t)

	count, err := svc.CountByUserID(userID)
	if err != nil {
		t.Fatalf("CountByUserID (empty): %v", err)
	}
	if count != 0 {
		t.Errorf("CountByUserID empty = %d, want 0", count)
	}

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	count, err = svc.CountByUserID(userID)
	if err != nil {
		t.Fatalf("CountByUserID (1 profile): %v", err)
	}
	if count != 1 {
		t.Errorf("CountByUserID = %d, want 1", count)
	}
}

// ---------------------------------------------------------------------------
// VerifyDomain
// ---------------------------------------------------------------------------

func TestProfileService_VerifyDomain(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	p.CustomDomain = "example.test"
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.VerifyDomain(p.ID, "example.test"); err != nil {
		t.Fatalf("VerifyDomain: %v", err)
	}

	got, err := svc.GetByID(p.ID)
	if err != nil {
		t.Fatalf("GetByID after verify: %v", err)
	}
	if !got.DomainVerified {
		t.Error("DomainVerified should be true after VerifyDomain")
	}
}

func TestProfileService_VerifyDomain_WrongProfile(t *testing.T) {
	svc, _ := newTestProfileService(t)

	if err := svc.VerifyDomain("no-such-profile", "example.test"); err != ErrProfileNotFound {
		t.Errorf("VerifyDomain(nonexistent) = %v, want ErrProfileNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// GetQRCodeSettings & UpdateQRCodeSettings
// ---------------------------------------------------------------------------

func TestProfileService_GetQRCodeSettings_Default(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	settings, err := svc.GetQRCodeSettings(p.ID)
	if err != nil {
		t.Fatalf("GetQRCodeSettings: %v", err)
	}
	if settings == nil {
		t.Fatal("GetQRCodeSettings returned nil")
	}
	// Default should be returned (not from DB)
	if settings.Size != 256 {
		t.Errorf("default Size = %d, want 256", settings.Size)
	}
}

func TestProfileService_UpdateQRCodeSettings(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	settings := &model.QRCodeSettings{
		ProfileID:       p.ID,
		Size:            512,
		ErrorCorrection: "H",
		Style:           "square",
		DarkColor:       "#000000",
		LightColor:      "#ffffff",
		LogoEnabled:     false,
		LogoSize:        30,
		Format:          "png",
	}
	if err := svc.UpdateQRCodeSettings(settings); err != nil {
		t.Fatalf("UpdateQRCodeSettings: %v", err)
	}

	got, err := svc.GetQRCodeSettings(p.ID)
	if err != nil {
		t.Fatalf("GetQRCodeSettings after update: %v", err)
	}
	if got.Size != 512 {
		t.Errorf("Size = %d, want 512", got.Size)
	}
}

// ---------------------------------------------------------------------------
// isValidSlug
// ---------------------------------------------------------------------------

func TestProfileService_IsValidSlug(t *testing.T) {
	db := newTestProfileDB(t)
	svc := NewProfileService(db)

	tests := []struct {
		slug  string
		valid bool
	}{
		{"abc", true},
		{"my-profile", true},
		{"my_profile", true},
		{"ABC123", true},
		{"ab", false},              // too short
		{"bad slug", false},        // space
		{"has!special", false},     // special char
		{string(make([]byte, 101)), false}, // too long
	}

	for _, tt := range tests {
		got := svc.isValidSlug(tt.slug)
		if got != tt.valid {
			t.Errorf("isValidSlug(%q) = %v, want %v", tt.slug, got, tt.valid)
		}
	}
}

// ---------------------------------------------------------------------------
// generateID
// ---------------------------------------------------------------------------

func TestProfileService_GenerateID_Unique(t *testing.T) {
	db := newTestProfileDB(t)
	svc := NewProfileService(db)

	ids := make(map[string]bool)
	for i := 0; i < 20; i++ {
		id := svc.generateID()
		if id == "" {
			t.Fatal("generateID returned empty string")
		}
		if ids[id] {
			t.Errorf("generateID returned duplicate: %s", id)
		}
		ids[id] = true
	}
}

// ---------------------------------------------------------------------------
// getMaxProfilesPerUser
// ---------------------------------------------------------------------------

func TestProfileService_GetMaxProfilesPerUser_Default(t *testing.T) {
	db := newTestProfileDB(t)
	svc := NewProfileService(db)

	max, err := svc.getMaxProfilesPerUser()
	if err != nil {
		t.Fatalf("getMaxProfilesPerUser: %v", err)
	}
	if max <= 0 {
		t.Errorf("getMaxProfilesPerUser = %d, want > 0", max)
	}
}

// ---------------------------------------------------------------------------
// generateUniqueSlug
// ---------------------------------------------------------------------------

func TestProfileService_GenerateUniqueSlug(t *testing.T) {
	db := newTestProfileDB(t)
	svc := NewProfileService(db)

	// No existing slugs — should just append "-copy"
	slug := svc.generateUniqueSlug("my-profile")
	if slug == "my-profile" {
		t.Error("generateUniqueSlug should not return the same slug")
	}
}

// generateUniqueSlug when "-copy" slug already exists should try "-copy-1", etc.
func TestProfileService_GenerateUniqueSlug_Collision(t *testing.T) {
	svc, userID := newTestProfileService(t)

	// Create "base-copy" first so the loop increments.
	copy1 := &model.Profile{Slug: "base-copy", DisplayName: "Copy", IsPublic: true}
	if err := svc.Create(userID, copy1); err != nil {
		t.Fatalf("Create base-copy: %v", err)
	}

	slug := svc.generateUniqueSlug("base")
	if slug == "base-copy" {
		t.Error("generateUniqueSlug should not return already-taken slug 'base-copy'")
	}
}

// ---------------------------------------------------------------------------
// Create — error paths
// ---------------------------------------------------------------------------

func TestProfileService_Create_MaxProfilesReached(t *testing.T) {
	svc, userID := newTestProfileService(t)

	// Set max_profiles_per_user to 1 via raw SQL.
	if _, err := svc.db.Exec(`UPDATE settings SET value = '1' WHERE key = 'max_profiles_per_user'`); err != nil {
		if _, err2 := svc.db.Exec(`INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES ('max_profiles_per_user', '1', CURRENT_TIMESTAMP)`); err2 != nil {
			t.Fatalf("seed max_profiles_per_user: %v / %v", err, err2)
		}
	}

	p1 := validProfile(userID)
	if err := svc.Create(userID, p1); err != nil {
		t.Fatalf("Create first profile (should succeed): %v", err)
	}

	p2 := &model.Profile{Slug: "second-profile", DisplayName: "Second", IsPublic: true}
	if err := svc.Create(userID, p2); err != ErrMaxProfilesReached {
		t.Errorf("Create beyond limit = %v, want ErrMaxProfilesReached", err)
	}
}

func TestProfileService_Create_DBError(t *testing.T) {
	// Closing DB forces CountByUserID (and subsequent DB calls) to fail.
	svc, userID := newTestProfileService(t)
	svc.db.DB.Close()

	p := &model.Profile{Slug: "some-valid-slug", DisplayName: "Test", IsPublic: true}
	err := svc.Create(userID, p)
	if err == nil {
		t.Error("Create with closed DB should return error")
	}
}

func TestProfileService_Create_ValidationError(t *testing.T) {
	svc, userID := newTestProfileService(t)

	// Slug is valid but display_name too long (>100 chars per model.Validate).
	p := &model.Profile{
		Slug:        "valid-slug",
		DisplayName: string(make([]byte, 101)),
		IsPublic:    true,
	}
	err := svc.Create(userID, p)
	if err == nil {
		t.Error("Create with display_name too long should return validation error")
	}
}

// ---------------------------------------------------------------------------
// GetByID — DB error path
// ---------------------------------------------------------------------------

func TestProfileService_GetByID_DBError(t *testing.T) {
	svc, _ := newTestProfileService(t)
	svc.db.DB.Close()

	_, err := svc.GetByID("any-id")
	if err == nil {
		t.Error("GetByID with closed DB should return error")
	}
	if err == ErrProfileNotFound {
		t.Error("GetByID DB error should not equal ErrProfileNotFound")
	}
}

// ---------------------------------------------------------------------------
// GetBySlug — DB error path
// ---------------------------------------------------------------------------

func TestProfileService_GetBySlug_DBError(t *testing.T) {
	svc, _ := newTestProfileService(t)
	svc.db.DB.Close()

	_, err := svc.GetBySlug("any-slug")
	if err == nil {
		t.Error("GetBySlug with closed DB should return error")
	}
	if err == ErrProfileNotFound {
		t.Error("GetBySlug DB error should not equal ErrProfileNotFound")
	}
}

// ---------------------------------------------------------------------------
// GetByUserID — DB error path
// ---------------------------------------------------------------------------

func TestProfileService_GetByUserID_DBError(t *testing.T) {
	svc, userID := newTestProfileService(t)
	svc.db.DB.Close()

	_, err := svc.GetByUserID(userID)
	if err == nil {
		t.Error("GetByUserID with closed DB should return error")
	}
}

// ---------------------------------------------------------------------------
// Update — error paths
// ---------------------------------------------------------------------------

func TestProfileService_Update_ValidationError(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// display_name too long triggers validation failure before the DB is hit.
	p.DisplayName = string(make([]byte, 101))
	if err := svc.Update(p); err == nil {
		t.Error("Update with display_name too long should return validation error")
	}
}

func TestProfileService_Update_NotFound(t *testing.T) {
	svc, _ := newTestProfileService(t)

	p := &model.Profile{
		ID:          "no-such-id",
		Slug:        "ghost-slug",
		DisplayName: "Ghost",
		IsPublic:    true,
	}
	if err := svc.Update(p); err != ErrProfileNotFound {
		t.Errorf("Update(nonexistent) = %v, want ErrProfileNotFound", err)
	}
}

func TestProfileService_Update_DBError(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	svc.db.DB.Close()

	p.DisplayName = "Changed"
	err := svc.Update(p)
	if err == nil {
		t.Error("Update with closed DB should return error")
	}
}

// ---------------------------------------------------------------------------
// Delete — DB error path
// ---------------------------------------------------------------------------

func TestProfileService_Delete_DBError(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	svc.db.DB.Close()

	err := svc.Delete(p.ID)
	if err == nil {
		t.Error("Delete with closed DB should return error")
	}
}

// ---------------------------------------------------------------------------
// Duplicate — error paths
// ---------------------------------------------------------------------------

func TestProfileService_Duplicate_NotFound(t *testing.T) {
	svc, userID := newTestProfileService(t)

	_, err := svc.Duplicate("no-such-profile", userID)
	if err != ErrProfileNotFound {
		t.Errorf("Duplicate(nonexistent) = %v, want ErrProfileNotFound", err)
	}
}

func TestProfileService_Duplicate_CreateFails(t *testing.T) {
	svc, userID := newTestProfileService(t)

	orig := validProfile(userID)
	if err := svc.Create(userID, orig); err != nil {
		t.Fatalf("Create original: %v", err)
	}

	// Cap max profiles at 1 so the duplicate Create fails.
	if _, err := svc.db.Exec(`UPDATE settings SET value = '1' WHERE key = 'max_profiles_per_user'`); err != nil {
		if _, err2 := svc.db.Exec(`INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES ('max_profiles_per_user', '1', CURRENT_TIMESTAMP)`); err2 != nil {
			t.Fatalf("seed max_profiles_per_user: %v / %v", err, err2)
		}
	}

	_, err := svc.Duplicate(orig.ID, userID)
	if err == nil {
		t.Error("Duplicate when at profile limit should return error")
	}
}

// ---------------------------------------------------------------------------
// IncrementViewCount — DB error path
// ---------------------------------------------------------------------------

func TestProfileService_IncrementViewCount_DBError(t *testing.T) {
	svc, _ := newTestProfileService(t)
	svc.db.DB.Close()

	err := svc.IncrementViewCount("any-id")
	if err == nil {
		t.Error("IncrementViewCount with closed DB should return error")
	}
}

// ---------------------------------------------------------------------------
// SlugExists — DB error path
// ---------------------------------------------------------------------------

func TestProfileService_SlugExists_DBError(t *testing.T) {
	svc, _ := newTestProfileService(t)
	svc.db.DB.Close()

	_, err := svc.SlugExists("any-slug")
	if err == nil {
		t.Error("SlugExists with closed DB should return error")
	}
}

// ---------------------------------------------------------------------------
// CountByUserID — DB error path
// ---------------------------------------------------------------------------

func TestProfileService_CountByUserID_DBError(t *testing.T) {
	svc, userID := newTestProfileService(t)
	svc.db.DB.Close()

	_, err := svc.CountByUserID(userID)
	if err == nil {
		t.Error("CountByUserID with closed DB should return error")
	}
}

// ---------------------------------------------------------------------------
// VerifyDomain — DB error path
// ---------------------------------------------------------------------------

func TestProfileService_VerifyDomain_DBError(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	p.CustomDomain = "example.test"
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	svc.db.DB.Close()

	err := svc.VerifyDomain(p.ID, "example.test")
	if err == nil {
		t.Error("VerifyDomain with closed DB should return error")
	}
}

// ---------------------------------------------------------------------------
// GetQRCodeSettings — DB error path
// ---------------------------------------------------------------------------

func TestProfileService_GetQRCodeSettings_DBError(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	svc.db.DB.Close()

	_, err := svc.GetQRCodeSettings(p.ID)
	if err == nil {
		t.Error("GetQRCodeSettings with closed DB should return error")
	}
}

// ---------------------------------------------------------------------------
// UpdateQRCodeSettings — error paths
// ---------------------------------------------------------------------------

func TestProfileService_UpdateQRCodeSettings_ValidationError(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Size = 0 should fail validation.
	settings := &model.QRCodeSettings{
		ProfileID:       p.ID,
		Size:            0,
		ErrorCorrection: "M",
		Style:           "square",
		DarkColor:       "#000000",
		LightColor:      "#ffffff",
		Format:          "png",
	}
	if err := svc.UpdateQRCodeSettings(settings); err == nil {
		t.Error("UpdateQRCodeSettings with Size=0 should return validation error")
	}
}

func TestProfileService_UpdateQRCodeSettings_DBError(t *testing.T) {
	svc, userID := newTestProfileService(t)

	p := validProfile(userID)
	if err := svc.Create(userID, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	svc.db.DB.Close()

	settings := &model.QRCodeSettings{
		ProfileID:       p.ID,
		Size:            256,
		ErrorCorrection: "M",
		Style:           "square",
		DarkColor:       "#000000",
		LightColor:      "#ffffff",
		LogoSize:        30,
		Format:          "png",
	}
	if err := svc.UpdateQRCodeSettings(settings); err == nil {
		t.Error("UpdateQRCodeSettings with closed DB should return error")
	}
}

// ---------------------------------------------------------------------------
// generateSlug
// ---------------------------------------------------------------------------

// TestProfileService_GenerateSlug_Counter exercises the counter-increment branch
// inside generateSlug when the first candidate slug is already taken.
func TestProfileService_GenerateSlug_Counter(t *testing.T) {
	svc, userID := newTestProfileService(t)

	// Create a profile whose slug exactly matches what generateSlug would produce
	// for the display name "Counter Test".
	taken := &model.Profile{Slug: "counter-test", DisplayName: "Counter Test", IsPublic: true}
	if err := svc.Create(userID, taken); err != nil {
		t.Fatalf("Create taken slug: %v", err)
	}

	// Now generate a slug for the same display name — the base "counter-test"
	// is taken so the counter branch (lines 456-457) must fire.
	slug, err := svc.generateSlug("Counter Test")
	if err != nil {
		t.Fatalf("generateSlug: %v", err)
	}
	if slug == "counter-test" {
		t.Error("generateSlug should not return already-taken slug 'counter-test'")
	}
}

// ---------------------------------------------------------------------------
// generateSlug — error path when SlugExists fails
// ---------------------------------------------------------------------------

func TestProfileService_GenerateSlug_DBError(t *testing.T) {
	svc, _ := newTestProfileService(t)
	svc.db.DB.Close()

	_, err := svc.generateSlug("Test Display Name")
	if err == nil {
		t.Error("generateSlug with closed DB should return error")
	}
}

// ---------------------------------------------------------------------------
// getMaxProfilesPerUser — invalid value path (returns default)
// ---------------------------------------------------------------------------

func TestProfileService_GetMaxProfilesPerUser_InvalidValue(t *testing.T) {
	db := newTestProfileDB(t)
	svc := NewProfileService(db)

	_, err := db.Exec(`INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES ('max_profiles_per_user', 'not-a-number', CURRENT_TIMESTAMP)`)
	if err != nil {
		t.Fatalf("seed invalid setting: %v", err)
	}

	max, err := svc.getMaxProfilesPerUser()
	if err != nil {
		t.Fatalf("getMaxProfilesPerUser returned unexpected error: %v", err)
	}
	if max != 5 {
		t.Errorf("getMaxProfilesPerUser with invalid value = %d, want 5 (default)", max)
	}
}
