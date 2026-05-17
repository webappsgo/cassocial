package service

import (
	"testing"

	"github.com/casapps/cassocial/src/server/model"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestLinkDB creates an in-memory DB with all migrations run.
func newTestLinkDB(t *testing.T) *store.DB {
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

// createTestUser inserts a user row directly so profiles/links can be created.
func createTestUser(t *testing.T, db *store.DB, id, username string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, status, email_verified, two_factor_enabled, created_at, updated_at)
		 VALUES (?, ?, ?, '$argon2id$v=19$m=65536,t=3,p=4$s$h', 'user', 'active', 1, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		id, username, username+"@example.com",
	)
	if err != nil {
		t.Fatalf("createTestUser(%s): %v", id, err)
	}
}

// createTestProfile inserts a profile via the store, returning the profile ID used.
func createTestProfile(t *testing.T, db *store.DB, userID, slug string) string {
	t.Helper()
	p := &store.Profile{
		ID:       "profile-" + slug,
		UserID:   userID,
		Slug:     slug,
		IsPublic: true,
	}
	if err := db.CreateProfile(p); err != nil {
		t.Fatalf("createTestProfile(%s): %v", slug, err)
	}
	return p.ID
}

// newTestLinkService creates a LinkService backed by an in-memory DB seeded with
// one user and one profile.
func newTestLinkService(t *testing.T) (*LinkService, string) {
	t.Helper()

	db := newTestLinkDB(t)
	createTestUser(t, db, "user-link-001", "linkowner")
	profileID := createTestProfile(t, db, "user-link-001", "link-test-profile")

	return NewLinkService(db), profileID
}

// validLink returns a model.Link ready to be inserted.
func validLink(profileID string) *model.Link {
	return &model.Link{
		ProfileID: profileID,
		Title:     "Test Link",
		URL:       "https://example.com",
	}
}

// ---------------------------------------------------------------------------
// NewLinkService
// ---------------------------------------------------------------------------

func TestNewLinkService(t *testing.T) {
	db := newTestLinkDB(t)
	svc := NewLinkService(db)
	if svc == nil {
		t.Fatal("NewLinkService returned nil")
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestLinkService_Create_Valid(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	link := validLink(profileID)
	if err := svc.Create(link); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if link.ID == "" {
		t.Error("Create did not populate ID")
	}
}

func TestLinkService_Create_SetsPosition(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	l1 := validLink(profileID)
	if err := svc.Create(l1); err != nil {
		t.Fatalf("Create l1: %v", err)
	}

	l2 := validLink(profileID)
	l2.Title = "Second Link"
	if err := svc.Create(l2); err != nil {
		t.Fatalf("Create l2: %v", err)
	}

	if l1.Position >= l2.Position {
		t.Errorf("positions not sequential: l1=%d l2=%d", l1.Position, l2.Position)
	}
}

func TestLinkService_Create_InvalidURL(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	link := &model.Link{
		ProfileID: profileID,
		Title:     "Bad URL",
		URL:       "not-a-url",
	}
	if err := svc.Create(link); err == nil {
		t.Error("Create with invalid URL should return error")
	}
}

func TestLinkService_Create_TitleTooLong(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	link := &model.Link{
		ProfileID: profileID,
		Title:     string(make([]byte, 101)), // 101 chars
		URL:       "https://example.com",
	}
	if err := svc.Create(link); err == nil {
		t.Error("Create with title too long should return error")
	}
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestLinkService_GetByID_Found(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	link := validLink(profileID)
	if err := svc.Create(link); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := svc.GetByID(link.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != link.ID {
		t.Errorf("GetByID ID = %q, want %q", got.ID, link.ID)
	}
	if got.Title != link.Title {
		t.Errorf("GetByID Title = %q, want %q", got.Title, link.Title)
	}
}

func TestLinkService_GetByID_NotFound(t *testing.T) {
	svc, _ := newTestLinkService(t)

	_, err := svc.GetByID("no-such-id")
	if err != ErrLinkNotFound {
		t.Errorf("GetByID(nonexistent) = %v, want ErrLinkNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// GetByProfileID
// ---------------------------------------------------------------------------

func TestLinkService_GetByProfileID_Empty(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	links, err := svc.GetByProfileID(profileID)
	if err != nil {
		t.Fatalf("GetByProfileID: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("GetByProfileID empty profile = %d links, want 0", len(links))
	}
}

func TestLinkService_GetByProfileID_Multiple(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	for i := 0; i < 3; i++ {
		l := validLink(profileID)
		l.Title = "Link"
		if err := svc.Create(l); err != nil {
			t.Fatalf("Create[%d]: %v", i, err)
		}
	}

	links, err := svc.GetByProfileID(profileID)
	if err != nil {
		t.Fatalf("GetByProfileID: %v", err)
	}
	if len(links) != 3 {
		t.Errorf("GetByProfileID = %d links, want 3", len(links))
	}
}

// ---------------------------------------------------------------------------
// GetActiveByProfileID
// ---------------------------------------------------------------------------

func TestLinkService_GetActiveByProfileID(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	active := validLink(profileID)
	active.Title = "Active"
	if err := svc.Create(active); err != nil {
		t.Fatalf("Create active: %v", err)
	}

	inactive := validLink(profileID)
	inactive.Title = "Inactive"
	if err := svc.Create(inactive); err != nil {
		t.Fatalf("Create inactive: %v", err)
	}
	// Toggle to disable it
	if err := svc.Toggle(inactive.ID); err != nil {
		t.Fatalf("Toggle: %v", err)
	}

	links, err := svc.GetActiveByProfileID(profileID)
	if err != nil {
		t.Fatalf("GetActiveByProfileID: %v", err)
	}
	if len(links) != 1 {
		t.Errorf("GetActiveByProfileID = %d links, want 1", len(links))
	}
	if links[0].ID != active.ID {
		t.Errorf("GetActiveByProfileID returned wrong link")
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestLinkService_Update_Valid(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	link := validLink(profileID)
	if err := svc.Create(link); err != nil {
		t.Fatalf("Create: %v", err)
	}

	link.Title = "Updated Title"
	link.URL = "https://updated.example.com"
	if err := svc.Update(link); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := svc.GetByID(link.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Title != "Updated Title" {
		t.Errorf("after update Title = %q, want 'Updated Title'", got.Title)
	}
}

func TestLinkService_Update_NotFound(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	link := &model.Link{
		ID:        "no-such-link",
		ProfileID: profileID,
		Title:     "Ghost",
		URL:       "https://example.com",
	}
	if err := svc.Update(link); err != ErrLinkNotFound {
		t.Errorf("Update(nonexistent) = %v, want ErrLinkNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestLinkService_Delete_Valid(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	link := validLink(profileID)
	if err := svc.Create(link); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Delete(link.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := svc.GetByID(link.ID)
	if err != ErrLinkNotFound {
		t.Errorf("GetByID after delete = %v, want ErrLinkNotFound", err)
	}
}

func TestLinkService_Delete_NotFound(t *testing.T) {
	svc, _ := newTestLinkService(t)

	if err := svc.Delete("no-such-link"); err != ErrLinkNotFound {
		t.Errorf("Delete(nonexistent) = %v, want ErrLinkNotFound", err)
	}
}

func TestLinkService_Delete_ReordersPositions(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	// Create 3 links
	var links []*model.Link
	for i := 0; i < 3; i++ {
		l := validLink(profileID)
		l.Title = "Link"
		if err := svc.Create(l); err != nil {
			t.Fatalf("Create[%d]: %v", i, err)
		}
		links = append(links, l)
	}

	// Delete the middle one
	if err := svc.Delete(links[1].ID); err != nil {
		t.Fatalf("Delete middle: %v", err)
	}

	remaining, err := svc.GetByProfileID(profileID)
	if err != nil {
		t.Fatalf("GetByProfileID: %v", err)
	}
	if len(remaining) != 2 {
		t.Fatalf("expected 2 remaining links, got %d", len(remaining))
	}
	// Positions should be 1 and 2
	if remaining[0].Position != 1 || remaining[1].Position != 2 {
		t.Errorf("positions after delete = %d, %d; want 1, 2",
			remaining[0].Position, remaining[1].Position)
	}
}

// ---------------------------------------------------------------------------
// Toggle
// ---------------------------------------------------------------------------

func TestLinkService_Toggle(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	link := validLink(profileID)
	if err := svc.Create(link); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Initially active — toggle should deactivate
	if err := svc.Toggle(link.ID); err != nil {
		t.Fatalf("Toggle (deactivate): %v", err)
	}

	got, _ := svc.GetByID(link.ID)
	if got.IsActive {
		t.Error("after first Toggle: IsActive should be false")
	}

	// Toggle again — should reactivate
	if err := svc.Toggle(link.ID); err != nil {
		t.Fatalf("Toggle (reactivate): %v", err)
	}

	got, _ = svc.GetByID(link.ID)
	if !got.IsActive {
		t.Error("after second Toggle: IsActive should be true")
	}
}

func TestLinkService_Toggle_NotFound(t *testing.T) {
	svc, _ := newTestLinkService(t)

	if err := svc.Toggle("no-such-link"); err != ErrLinkNotFound {
		t.Errorf("Toggle(nonexistent) = %v, want ErrLinkNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// Reorder
// ---------------------------------------------------------------------------

func TestLinkService_Reorder(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	var links []*model.Link
	for i := 0; i < 3; i++ {
		l := validLink(profileID)
		l.Title = "Link"
		if err := svc.Create(l); err != nil {
			t.Fatalf("Create[%d]: %v", i, err)
		}
		links = append(links, l)
	}

	// Reverse the order
	newOrder := []string{links[2].ID, links[1].ID, links[0].ID}
	if err := svc.Reorder(profileID, newOrder); err != nil {
		t.Fatalf("Reorder: %v", err)
	}

	reordered, err := svc.GetByProfileID(profileID)
	if err != nil {
		t.Fatalf("GetByProfileID after reorder: %v", err)
	}

	if reordered[0].ID != links[2].ID {
		t.Errorf("after reorder: first link = %q, want %q", reordered[0].ID, links[2].ID)
	}
}

// ---------------------------------------------------------------------------
// IncrementClickCount
// ---------------------------------------------------------------------------

func TestLinkService_IncrementClickCount(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	link := validLink(profileID)
	if err := svc.Create(link); err != nil {
		t.Fatalf("Create: %v", err)
	}

	for i := 0; i < 3; i++ {
		if err := svc.IncrementClickCount(link.ID); err != nil {
			t.Fatalf("IncrementClickCount[%d]: %v", i, err)
		}
	}

	got, err := svc.GetByID(link.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ClickCount != 3 {
		t.Errorf("ClickCount = %d, want 3", got.ClickCount)
	}
}

// ---------------------------------------------------------------------------
// CountByProfileID
// ---------------------------------------------------------------------------

func TestLinkService_CountByProfileID(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	count, err := svc.CountByProfileID(profileID)
	if err != nil {
		t.Fatalf("CountByProfileID (empty): %v", err)
	}
	if count != 0 {
		t.Errorf("CountByProfileID empty = %d, want 0", count)
	}

	for i := 0; i < 3; i++ {
		l := validLink(profileID)
		l.Title = "Link"
		if err := svc.Create(l); err != nil {
			t.Fatalf("Create[%d]: %v", i, err)
		}
	}

	count, err = svc.CountByProfileID(profileID)
	if err != nil {
		t.Fatalf("CountByProfileID (3 links): %v", err)
	}
	if count != 3 {
		t.Errorf("CountByProfileID = %d, want 3", count)
	}
}

// ---------------------------------------------------------------------------
// GetTopClickedLinks
// ---------------------------------------------------------------------------

func TestLinkService_GetTopClickedLinks(t *testing.T) {
	svc, profileID := newTestLinkService(t)

	// Create two links and give the second more clicks
	l1 := validLink(profileID)
	l1.Title = "Low Clicks"
	if err := svc.Create(l1); err != nil {
		t.Fatalf("Create l1: %v", err)
	}

	l2 := validLink(profileID)
	l2.Title = "High Clicks"
	if err := svc.Create(l2); err != nil {
		t.Fatalf("Create l2: %v", err)
	}
	for i := 0; i < 5; i++ {
		svc.IncrementClickCount(l2.ID) //nolint:errcheck
	}

	top, err := svc.GetTopClickedLinks(profileID, 2)
	if err != nil {
		t.Fatalf("GetTopClickedLinks: %v", err)
	}
	if len(top) == 0 {
		t.Fatal("GetTopClickedLinks returned no links")
	}
	if top[0].ID != l2.ID {
		t.Errorf("top link = %q, want %q (most clicks)", top[0].ID, l2.ID)
	}
}

// ---------------------------------------------------------------------------
// generateID
// ---------------------------------------------------------------------------

func TestLinkService_GenerateID_Unique(t *testing.T) {
	db := newTestLinkDB(t)
	svc := NewLinkService(db)

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
// getMaxLinksPerProfile
// ---------------------------------------------------------------------------

func TestLinkService_GetMaxLinksPerProfile_Default(t *testing.T) {
	db := newTestLinkDB(t)
	svc := NewLinkService(db)

	max, err := svc.getMaxLinksPerProfile()
	if err != nil {
		t.Fatalf("getMaxLinksPerProfile: %v", err)
	}
	if max <= 0 {
		t.Errorf("getMaxLinksPerProfile = %d, want > 0", max)
	}
}
