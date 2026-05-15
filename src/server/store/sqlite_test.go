package store

import (
	"testing"
)

// newTestDB opens an in-memory SQLite database and runs migrations.
func newTestDB(t *testing.T) *DB {
	t.Helper()

	db, err := Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Connect(\"sqlite\", \":memory:\") returned error: %v", err)
	}

	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	return db
}

func TestRunMigrations(t *testing.T) {
	db, err := Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	defer db.Close()

	if err := db.RunMigrations(); err != nil {
		t.Errorf("RunMigrations returned error: %v", err)
	}
}

func TestCreateAndGetUser(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		wantErr  bool
	}{
		{
			name: "valid user",
			user: &User{
				ID:               "test-user-id-001",
				Username:         "testuser",
				Email:            "test@example.com",
				PasswordHash:     "$argon2id$v=19$m=65536,t=3,p=4$fakesalt$fakehash",
				Role:             "user",
				Status:           "active",
				EmailVerified:    true,
				TwoFactorEnabled: false,
				TwoFactorSecret:  "",
			},
			wantErr: false,
		},
		{
			name: "admin user",
			user: &User{
				ID:               "test-user-id-002",
				Username:         "adminuser",
				Email:            "admin@example.com",
				PasswordHash:     "$argon2id$v=19$m=65536,t=3,p=4$fakesalt$fakehash",
				Role:             "admin",
				Status:           "active",
				EmailVerified:    true,
				TwoFactorEnabled: false,
				TwoFactorSecret:  "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)

			err := db.CreateUser(tt.user)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateUser returned error %v, wantErr=%v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			got, err := db.GetUserByID(tt.user.ID)
			if err != nil {
				t.Fatalf("GetUserByID(%q) returned error: %v", tt.user.ID, err)
			}

			if got.ID != tt.user.ID {
				t.Errorf("GetUserByID ID = %q, want %q", got.ID, tt.user.ID)
			}
			if got.Username != tt.user.Username {
				t.Errorf("GetUserByID Username = %q, want %q", got.Username, tt.user.Username)
			}
			if got.Email != tt.user.Email {
				t.Errorf("GetUserByID Email = %q, want %q", got.Email, tt.user.Email)
			}
			if got.Role != tt.user.Role {
				t.Errorf("GetUserByID Role = %q, want %q", got.Role, tt.user.Role)
			}
			if got.Status != tt.user.Status {
				t.Errorf("GetUserByID Status = %q, want %q", got.Status, tt.user.Status)
			}
			if got.EmailVerified != tt.user.EmailVerified {
				t.Errorf("GetUserByID EmailVerified = %v, want %v", got.EmailVerified, tt.user.EmailVerified)
			}
		})
	}
}

func TestCreateAndGetProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile *Profile
	}{
		{
			name: "public profile",
			profile: &Profile{
				ID:          "test-profile-id-001",
				UserID:      "test-user-id-001",
				Slug:        "myprofile",
				DisplayName: "My Profile",
				Bio:         "A test bio",
				IsPublic:    true,
			},
		},
		{
			name: "private profile",
			profile: &Profile{
				ID:          "test-profile-id-002",
				UserID:      "test-user-id-002",
				Slug:        "private-profile",
				DisplayName: "Private Profile",
				Bio:         "",
				IsPublic:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)

			// CreateProfile requires the user to exist due to foreign key constraints.
			user := &User{
				ID:           tt.profile.UserID,
				Username:     "owner-" + tt.profile.ID,
				Email:        tt.profile.ID + "@example.com",
				PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$s$h",
				Role:         "user",
				Status:       "active",
			}
			if err := db.CreateUser(user); err != nil {
				t.Fatalf("CreateUser (prerequisite) returned error: %v", err)
			}

			if err := db.CreateProfile(tt.profile); err != nil {
				t.Fatalf("CreateProfile returned error: %v", err)
			}

			got, err := db.GetProfileByID(tt.profile.ID)
			if err != nil {
				t.Fatalf("GetProfileByID(%q) returned error: %v", tt.profile.ID, err)
			}

			if got.ID != tt.profile.ID {
				t.Errorf("GetProfileByID ID = %q, want %q", got.ID, tt.profile.ID)
			}
			if got.UserID != tt.profile.UserID {
				t.Errorf("GetProfileByID UserID = %q, want %q", got.UserID, tt.profile.UserID)
			}
			if got.Slug != tt.profile.Slug {
				t.Errorf("GetProfileByID Slug = %q, want %q", got.Slug, tt.profile.Slug)
			}
			if got.DisplayName != tt.profile.DisplayName {
				t.Errorf("GetProfileByID DisplayName = %q, want %q", got.DisplayName, tt.profile.DisplayName)
			}
			if got.IsPublic != tt.profile.IsPublic {
				t.Errorf("GetProfileByID IsPublic = %v, want %v", got.IsPublic, tt.profile.IsPublic)
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	db := newTestDB(t)

	user := &User{
		ID:           "upd-user-001",
		Username:     "updateowner",
		Email:        "updateowner@example.com",
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$s$h",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser (prerequisite) returned error: %v", err)
	}

	profile := &Profile{
		ID:          "upd-profile-001",
		UserID:      "upd-user-001",
		Slug:        "update-test",
		DisplayName: "Original Name",
		Bio:         "Original bio",
		IsPublic:    false,
	}
	if err := db.CreateProfile(profile); err != nil {
		t.Fatalf("CreateProfile returned error: %v", err)
	}

	profile.DisplayName = "Updated Name"
	profile.Bio = "Updated bio"
	profile.IsPublic = true

	if err := db.UpdateProfile(profile); err != nil {
		t.Fatalf("UpdateProfile returned error: %v", err)
	}

	got, err := db.GetProfileByID(profile.ID)
	if err != nil {
		t.Fatalf("GetProfileByID after update returned error: %v", err)
	}

	if got.DisplayName != "Updated Name" {
		t.Errorf("after update: DisplayName = %q, want %q", got.DisplayName, "Updated Name")
	}
	if got.Bio != "Updated bio" {
		t.Errorf("after update: Bio = %q, want %q", got.Bio, "Updated bio")
	}
	if !got.IsPublic {
		t.Errorf("after update: IsPublic = false, want true")
	}
}

func TestDeleteProfile(t *testing.T) {
	db := newTestDB(t)

	user := &User{
		ID:           "del-user-001",
		Username:     "deleteowner",
		Email:        "deleteowner@example.com",
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$s$h",
		Role:         "user",
		Status:       "active",
	}
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("CreateUser (prerequisite) returned error: %v", err)
	}

	profile := &Profile{
		ID:     "del-profile-001",
		UserID: "del-user-001",
		Slug:   "delete-test",
	}
	if err := db.CreateProfile(profile); err != nil {
		t.Fatalf("CreateProfile returned error: %v", err)
	}

	if err := db.DeleteProfile(profile.ID); err != nil {
		t.Fatalf("DeleteProfile returned error: %v", err)
	}

	_, err := db.GetProfileByID(profile.ID)
	if err == nil {
		t.Errorf("GetProfileByID after delete returned nil error, want an error (record not found)")
	}
}

