package store

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrations embed.FS

type DB struct {
	*sql.DB
	Driver string
}

// Connect establishes a database connection
func Connect(driver, dbPath string) (*DB, error) {
	var dsn string

	// Build DSN based on driver
	switch driver {
	case "sqlite":
		// Ensure directory exists
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
		dsn = dbPath

	case "pgx", "postgres":
		driver = "pgx"
		dsn = dbPath // Expected to be connection string

	case "mysql":
		dsn = dbPath // Expected to be connection string

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db, Driver: driver}, nil
}

// RunMigrations executes all SQL migration files
func (db *DB) RunMigrations() error {
	// Try to read from embedded FS first
	if migrations.ReadDir == nil {
		log.Println("No embedded migrations found, skipping migration")
		return nil
	}

	files, err := migrations.ReadDir("migrations")
	if err != nil {
		log.Printf("Could not read migrations directory: %v", err)
		return nil
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		content, err := migrations.ReadFile("migrations/" + file.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file.Name(), err)
		}

		// Adapt SQL for different databases
		sqlContent := string(content)
		sqlContent = db.adaptSQL(sqlContent)

		// Execute migration
		if _, err := db.Exec(sqlContent); err != nil {
			log.Printf("Migration %s error (may be already applied): %v", file.Name(), err)
		} else {
			log.Printf("Applied migration: %s", file.Name())
		}
	}

	return nil
}

// adaptSQL adapts SQLite SQL to PostgreSQL or MySQL
func (db *DB) adaptSQL(sql string) string {
	switch db.Driver {
	case "pgx":
		// Replace TEXT PRIMARY KEY with UUID
		sql = strings.ReplaceAll(sql, "TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16))))", "UUID PRIMARY KEY DEFAULT gen_random_uuid()")
		sql = strings.ReplaceAll(sql, "TEXT PRIMARY KEY", "UUID PRIMARY KEY")
		sql = strings.ReplaceAll(sql, "TEXT REFERENCES", "UUID REFERENCES")
		sql = strings.ReplaceAll(sql, "TEXT NOT NULL REFERENCES", "UUID NOT NULL REFERENCES")
		sql = strings.ReplaceAll(sql, "BOOLEAN DEFAULT 0", "BOOLEAN DEFAULT false")
		sql = strings.ReplaceAll(sql, "BOOLEAN DEFAULT 1", "BOOLEAN DEFAULT true")
		sql = strings.ReplaceAll(sql, "INSERT OR IGNORE", "INSERT ON CONFLICT DO NOTHING")

	case "mysql":
		// MySQL adaptations
		sql = strings.ReplaceAll(sql, "TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16))))", "CHAR(36) PRIMARY KEY DEFAULT (UUID())")
		sql = strings.ReplaceAll(sql, "TEXT PRIMARY KEY", "CHAR(36) PRIMARY KEY")
		sql = strings.ReplaceAll(sql, "TEXT REFERENCES", "CHAR(36)")
		sql = strings.ReplaceAll(sql, "TEXT NOT NULL REFERENCES", "CHAR(36) NOT NULL")
		sql = strings.ReplaceAll(sql, "BOOLEAN DEFAULT 0", "TINYINT(1) DEFAULT 0")
		sql = strings.ReplaceAll(sql, "BOOLEAN DEFAULT 1", "TINYINT(1) DEFAULT 1")
		sql = strings.ReplaceAll(sql, "BOOLEAN DEFAULT", "TINYINT(1) DEFAULT")
		sql = strings.ReplaceAll(sql, "INSERT OR IGNORE", "INSERT IGNORE")
		sql = strings.ReplaceAll(sql, "IF NOT EXISTS", "IF NOT EXISTS")
	}

	return sql
}

// getDataDirectory returns the appropriate data directory
func getDataDirectory() string {
	// Check for portable mode
	if _, err := os.Stat("./data"); err == nil {
		return "./data"
	}

	// Check if running as root/system
	if os.Geteuid() == 0 {
		dir := "/var/lib/cassocial"
		os.MkdirAll(dir, 0755)
		return dir
	}

	// User installation
	homeDir, _ := os.UserHomeDir()
	dir := filepath.Join(homeDir, ".local", "share", "cassocial")
	os.MkdirAll(dir, 0755)
	return dir
}

// GetSetting retrieves a setting value
func (db *DB) GetSetting(key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetSetting updates or inserts a setting
func (db *DB) SetSetting(key, value string) error {
	placeholder := "?"
	if db.Driver == "pgx" {
		placeholder = "$1, $2"
	}

	query := fmt.Sprintf(`
		INSERT INTO settings (key, value, updated_at)
		VALUES (%s, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
	`, placeholder)

	if db.Driver == "mysql" {
		query = `
			INSERT INTO settings (key, value, updated_at)
			VALUES (?, ?, CURRENT_TIMESTAMP)
			ON DUPLICATE KEY UPDATE value = VALUES(value), updated_at = CURRENT_TIMESTAMP
		`
	}

	_, err := db.Exec(query, key, value)
	return err
}

// GetAllSettings retrieves all settings
func (db *DB) GetAllSettings() (map[string]string, error) {
rows, err := db.Query("SELECT key, value FROM settings")
if err != nil {
return nil, err
}
defer rows.Close()

settings := make(map[string]string)
for rows.Next() {
var key, value string
if err := rows.Scan(&key, &value); err != nil {
return nil, err
}
settings[key] = value
}

return settings, rows.Err()
}

// User CRUD operations
// Per PART 23: Regular User operations

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(id string) (*User, error) {
user := &User{}
err := db.QueryRow(`
SELECT id, username, email, password_hash, role, status,
       email_verified, two_factor_enabled, two_factor_secret,
       created_at, updated_at, last_login
FROM users WHERE id = ?
`, id).Scan(
&user.ID, &user.Username, &user.Email, &user.PasswordHash,
&user.Role, &user.Status, &user.EmailVerified,
&user.TwoFactorEnabled, &user.TwoFactorSecret,
&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
)
if err != nil {
return nil, err
}
return user, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(email string) (*User, error) {
user := &User{}
err := db.QueryRow(`
SELECT id, username, email, password_hash, role, status,
       email_verified, two_factor_enabled, two_factor_secret,
       created_at, updated_at, last_login
FROM users WHERE email = ?
`, email).Scan(
&user.ID, &user.Username, &user.Email, &user.PasswordHash,
&user.Role, &user.Status, &user.EmailVerified,
&user.TwoFactorEnabled, &user.TwoFactorSecret,
&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
)
if err != nil {
return nil, err
}
return user, nil
}

// GetUserByUsername retrieves a user by username
func (db *DB) GetUserByUsername(username string) (*User, error) {
user := &User{}
err := db.QueryRow(`
SELECT id, username, email, password_hash, role, status,
       email_verified, two_factor_enabled, two_factor_secret,
       created_at, updated_at, last_login
FROM users WHERE username = ?
`, username).Scan(
&user.ID, &user.Username, &user.Email, &user.PasswordHash,
&user.Role, &user.Status, &user.EmailVerified,
&user.TwoFactorEnabled, &user.TwoFactorSecret,
&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
)
if err != nil {
return nil, err
}
return user, nil
}

// CreateUser creates a new user
func (db *DB) CreateUser(user *User) error {
_, err := db.Exec(`
INSERT INTO users (id, username, email, password_hash, role, status,
                   email_verified, two_factor_enabled, two_factor_secret,
                   created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
`, user.ID, user.Username, user.Email, user.PasswordHash, user.Role,
user.Status, user.EmailVerified, user.TwoFactorEnabled, user.TwoFactorSecret)
return err
}

// UpdateUser updates an existing user
func (db *DB) UpdateUser(user *User) error {
_, err := db.Exec(`
UPDATE users SET
username = ?, email = ?, password_hash = ?, role = ?, status = ?,
email_verified = ?, two_factor_enabled = ?, two_factor_secret = ?,
updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, user.Username, user.Email, user.PasswordHash, user.Role, user.Status,
user.EmailVerified, user.TwoFactorEnabled, user.TwoFactorSecret, user.ID)
return err
}

// DeleteUser deletes a user
func (db *DB) DeleteUser(id string) error {
_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
return err
}

// ListUsers retrieves a paginated list of users
func (db *DB) ListUsers(limit, offset int) ([]*User, error) {
rows, err := db.Query(`
SELECT id, username, email, password_hash, role, status,
       email_verified, two_factor_enabled, two_factor_secret,
       created_at, updated_at, last_login
FROM users
ORDER BY created_at DESC
LIMIT ? OFFSET ?
`, limit, offset)
if err != nil {
return nil, err
}
defer rows.Close()

var users []*User
for rows.Next() {
user := &User{}
if err := rows.Scan(
&user.ID, &user.Username, &user.Email, &user.PasswordHash,
&user.Role, &user.Status, &user.EmailVerified,
&user.TwoFactorEnabled, &user.TwoFactorSecret,
&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
); err != nil {
return nil, err
}
users = append(users, user)
}

return users, rows.Err()
}

// CountUsers returns the total number of users
func (db *DB) CountUsers() (int, error) {
var count int
err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
return count, err
}

// Server Admin operations
// Per PART 23: Server Admins are separate from Regular Users

// GetServerAdminByID retrieves a server admin by ID
func (db *DB) GetServerAdminByID(id string) (*ServerAdmin, error) {
admin := &ServerAdmin{}
err := db.QueryRow(`
SELECT id, username, email, password_hash, is_primary,
       two_factor_enabled, two_factor_secret,
       created_at, updated_at, last_login
FROM server_admins WHERE id = ?
`, id).Scan(
&admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash,
&admin.IsPrimary, &admin.TwoFactorEnabled, &admin.TwoFactorSecret,
&admin.CreatedAt, &admin.UpdatedAt, &admin.LastLogin,
)
if err != nil {
return nil, err
}
return admin, nil
}

// GetServerAdminByEmail retrieves a server admin by email
func (db *DB) GetServerAdminByEmail(email string) (*ServerAdmin, error) {
admin := &ServerAdmin{}
err := db.QueryRow(`
SELECT id, username, email, password_hash, is_primary,
       two_factor_enabled, two_factor_secret,
       created_at, updated_at, last_login
FROM server_admins WHERE email = ?
`, email).Scan(
&admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash,
&admin.IsPrimary, &admin.TwoFactorEnabled, &admin.TwoFactorSecret,
&admin.CreatedAt, &admin.UpdatedAt, &admin.LastLogin,
)
if err != nil {
return nil, err
}
return admin, nil
}

// GetServerAdminByUsername retrieves a server admin by username
func (db *DB) GetServerAdminByUsername(username string) (*ServerAdmin, error) {
admin := &ServerAdmin{}
err := db.QueryRow(`
SELECT id, username, email, password_hash, is_primary,
       two_factor_enabled, two_factor_secret,
       created_at, updated_at, last_login
FROM server_admins WHERE username = ?
`, username).Scan(
&admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash,
&admin.IsPrimary, &admin.TwoFactorEnabled, &admin.TwoFactorSecret,
&admin.CreatedAt, &admin.UpdatedAt, &admin.LastLogin,
)
if err != nil {
return nil, err
}
return admin, nil
}

// CreateServerAdmin creates a new server admin
func (db *DB) CreateServerAdmin(admin *ServerAdmin) error {
_, err := db.Exec(`
INSERT INTO server_admins (id, username, email, password_hash, is_primary,
                            two_factor_enabled, two_factor_secret,
                            created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
`, admin.ID, admin.Username, admin.Email, admin.PasswordHash, admin.IsPrimary,
admin.TwoFactorEnabled, admin.TwoFactorSecret)
return err
}

// UpdateServerAdmin updates an existing server admin
func (db *DB) UpdateServerAdmin(admin *ServerAdmin) error {
_, err := db.Exec(`
UPDATE server_admins SET
username = ?, email = ?, password_hash = ?, is_primary = ?,
two_factor_enabled = ?, two_factor_secret = ?,
updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, admin.Username, admin.Email, admin.PasswordHash, admin.IsPrimary,
admin.TwoFactorEnabled, admin.TwoFactorSecret, admin.ID)
return err
}

// DeleteServerAdmin deletes a server admin
// Per PART 23: Primary Admin cannot be deleted
func (db *DB) DeleteServerAdmin(id string) error {
var isPrimary bool
err := db.QueryRow("SELECT is_primary FROM server_admins WHERE id = ?", id).Scan(&isPrimary)
if err != nil {
return err
}

if isPrimary {
return fmt.Errorf("cannot delete primary admin")
}

_, err = db.Exec("DELETE FROM server_admins WHERE id = ?", id)
return err
}

// GetPrimaryAdmin retrieves the primary admin
func (db *DB) GetPrimaryAdmin() (*ServerAdmin, error) {
admin := &ServerAdmin{}
err := db.QueryRow(`
SELECT id, username, email, password_hash, is_primary,
       two_factor_enabled, two_factor_secret,
       created_at, updated_at, last_login
FROM server_admins WHERE is_primary = 1
`).Scan(
&admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash,
&admin.IsPrimary, &admin.TwoFactorEnabled, &admin.TwoFactorSecret,
&admin.CreatedAt, &admin.UpdatedAt, &admin.LastLogin,
)
if err != nil {
return nil, err
}
return admin, nil
}

// ListServerAdmins retrieves all server admins
func (db *DB) ListServerAdmins() ([]*ServerAdmin, error) {
rows, err := db.Query(`
SELECT id, username, email, password_hash, is_primary,
       two_factor_enabled, two_factor_secret,
       created_at, updated_at, last_login
FROM server_admins
ORDER BY is_primary DESC, created_at ASC
`)
if err != nil {
return nil, err
}
defer rows.Close()

var admins []*ServerAdmin
for rows.Next() {
admin := &ServerAdmin{}
if err := rows.Scan(
&admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash,
&admin.IsPrimary, &admin.TwoFactorEnabled, &admin.TwoFactorSecret,
&admin.CreatedAt, &admin.UpdatedAt, &admin.LastLogin,
); err != nil {
return nil, err
}
admins = append(admins, admin)
}

return admins, rows.Err()
}

// Session operations
// Per PART 23: Session management for authentication

// CreateSession creates a new session
func (db *DB) CreateSession(session *Session) error {
_, err := db.Exec(`
INSERT INTO sessions (id, user_id, user_type, username, role, expires_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
`, session.ID, session.UserID, session.UserType, session.Username, session.Role, session.ExpiresAt)
return err
}

// GetSession retrieves a session by ID
func (db *DB) GetSession(sessionID string) (*Session, error) {
session := &Session{}
err := db.QueryRow(`
SELECT id, user_id, user_type, username, role, expires_at, created_at
FROM sessions WHERE id = ?
`, sessionID).Scan(
&session.ID, &session.UserID, &session.UserType,
&session.Username, &session.Role, &session.ExpiresAt, &session.CreatedAt,
)
if err != nil {
return nil, err
}
return session, nil
}

// DeleteSession deletes a session
func (db *DB) DeleteSession(sessionID string) error {
_, err := db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
return err
}

// DeleteSessionsByUserID deletes all sessions for a user
func (db *DB) DeleteSessionsByUserID(userID string) error {
_, err := db.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
return err
}

// CleanupExpiredSessions removes expired sessions
func (db *DB) CleanupExpiredSessions() error {
_, err := db.Exec("DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP")
return err
}

// Email verification operations
// Per PART 23: Email verification flow

// CreateEmailVerificationToken creates a verification token
func (db *DB) CreateEmailVerificationToken(token *EmailVerificationToken) error {
_, err := db.Exec(`
INSERT INTO email_verification_tokens (token, user_id, expires_at, created_at)
VALUES (?, ?, ?, CURRENT_TIMESTAMP)
`, token.Token, token.UserID, token.ExpiresAt)
return err
}

// GetEmailVerificationToken retrieves a verification token
func (db *DB) GetEmailVerificationToken(token string) (*EmailVerificationToken, error) {
evt := &EmailVerificationToken{}
err := db.QueryRow(`
SELECT token, user_id, expires_at, created_at
FROM email_verification_tokens WHERE token = ?
`, token).Scan(&evt.Token, &evt.UserID, &evt.ExpiresAt, &evt.CreatedAt)
if err != nil {
return nil, err
}
return evt, nil
}

// DeleteEmailVerificationToken deletes a verification token
func (db *DB) DeleteEmailVerificationToken(token string) error {
_, err := db.Exec("DELETE FROM email_verification_tokens WHERE token = ?", token)
return err
}

// DeleteExpiredEmailVerificationTokens removes expired tokens
func (db *DB) DeleteExpiredEmailVerificationTokens() error {
_, err := db.Exec("DELETE FROM email_verification_tokens WHERE expires_at < CURRENT_TIMESTAMP")
return err
}

// Password reset operations
// Per PART 23: Password reset flow

// CreatePasswordResetToken creates a password reset token
func (db *DB) CreatePasswordResetToken(token *PasswordResetToken) error {
_, err := db.Exec(`
INSERT INTO password_reset_tokens (token, user_id, expires_at, created_at)
VALUES (?, ?, ?, CURRENT_TIMESTAMP)
`, token.Token, token.UserID, token.ExpiresAt)
return err
}

// GetPasswordResetToken retrieves a password reset token
func (db *DB) GetPasswordResetToken(token string) (*PasswordResetToken, error) {
prt := &PasswordResetToken{}
err := db.QueryRow(`
SELECT token, user_id, expires_at, created_at
FROM password_reset_tokens WHERE token = ?
`, token).Scan(&prt.Token, &prt.UserID, &prt.ExpiresAt, &prt.CreatedAt)
if err != nil {
return nil, err
}
return prt, nil
}

// DeletePasswordResetToken deletes a password reset token
func (db *DB) DeletePasswordResetToken(token string) error {
_, err := db.Exec("DELETE FROM password_reset_tokens WHERE token = ?", token)
return err
}

// DeleteExpiredPasswordResetTokens removes expired tokens
func (db *DB) DeleteExpiredPasswordResetTokens() error {
_, err := db.Exec("DELETE FROM password_reset_tokens WHERE expires_at < CURRENT_TIMESTAMP")
return err
}
package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Profile operations
// Per PART 36: Profile is a user's public landing page

// GetProfileByID retrieves a profile by ID
func (db *DB) GetProfileByID(id string) (*Profile, error) {
	profile := &Profile{}
	err := db.QueryRow(`
		SELECT id, user_id, slug, display_name, bio, avatar_url, header_image_url,
		       theme_id, custom_css, show_usernames, is_public, password_protected,
		       protection_password, custom_domain, domain_verified, analytics_enabled,
		       meta_title, meta_description, og_image_url, view_count, qr_code_enabled,
		       created_at, updated_at
		FROM profiles WHERE id = ?
	`, id).Scan(
		&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
		&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
		&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
		&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
		&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
		&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
		&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// GetProfileBySlug retrieves a profile by slug
func (db *DB) GetProfileBySlug(slug string) (*Profile, error) {
	profile := &Profile{}
	err := db.QueryRow(`
		SELECT id, user_id, slug, display_name, bio, avatar_url, header_image_url,
		       theme_id, custom_css, show_usernames, is_public, password_protected,
		       protection_password, custom_domain, domain_verified, analytics_enabled,
		       meta_title, meta_description, og_image_url, view_count, qr_code_enabled,
		       created_at, updated_at
		FROM profiles WHERE slug = ?
	`, slug).Scan(
		&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
		&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
		&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
		&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
		&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
		&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
		&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// GetProfileByCustomDomain retrieves a profile by custom domain
func (db *DB) GetProfileByCustomDomain(domain string) (*Profile, error) {
	profile := &Profile{}
	err := db.QueryRow(`
		SELECT id, user_id, slug, display_name, bio, avatar_url, header_image_url,
		       theme_id, custom_css, show_usernames, is_public, password_protected,
		       protection_password, custom_domain, domain_verified, analytics_enabled,
		       meta_title, meta_description, og_image_url, view_count, qr_code_enabled,
		       created_at, updated_at
		FROM profiles WHERE custom_domain = ? AND domain_verified = 1
	`, domain).Scan(
		&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
		&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
		&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
		&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
		&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
		&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
		&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// GetProfilesByUserID retrieves all profiles for a user
func (db *DB) GetProfilesByUserID(userID string) ([]*Profile, error) {
	rows, err := db.Query(`
		SELECT id, user_id, slug, display_name, bio, avatar_url, header_image_url,
		       theme_id, custom_css, show_usernames, is_public, password_protected,
		       protection_password, custom_domain, domain_verified, analytics_enabled,
		       meta_title, meta_description, og_image_url, view_count, qr_code_enabled,
		       created_at, updated_at
		FROM profiles WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*Profile
	for rows.Next() {
		profile := &Profile{}
		if err := rows.Scan(
			&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
			&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
			&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
			&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
			&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
			&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
			&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, rows.Err()
}

// CreateProfile creates a new profile
func (db *DB) CreateProfile(profile *Profile) error {
	_, err := db.Exec(`
		INSERT INTO profiles (
			id, user_id, slug, display_name, bio, avatar_url, header_image_url,
			theme_id, custom_css, show_usernames, is_public, password_protected,
			protection_password, custom_domain, domain_verified, analytics_enabled,
			meta_title, meta_description, og_image_url, view_count, qr_code_enabled,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, profile.ID, profile.UserID, profile.Slug, profile.DisplayName, profile.Bio,
		profile.AvatarURL, profile.HeaderImageURL, profile.ThemeID, profile.CustomCSS,
		profile.ShowUsernames, profile.IsPublic, profile.PasswordProtected,
		profile.ProtectionPassword, profile.CustomDomain, profile.DomainVerified,
		profile.AnalyticsEnabled, profile.MetaTitle, profile.MetaDescription,
		profile.OgImageURL, profile.ViewCount, profile.QRCodeEnabled)
	return err
}

// UpdateProfile updates an existing profile
func (db *DB) UpdateProfile(profile *Profile) error {
	_, err := db.Exec(`
		UPDATE profiles SET
			slug = ?, display_name = ?, bio = ?, avatar_url = ?, header_image_url = ?,
			theme_id = ?, custom_css = ?, show_usernames = ?, is_public = ?,
			password_protected = ?, protection_password = ?, custom_domain = ?,
			domain_verified = ?, analytics_enabled = ?, meta_title = ?,
			meta_description = ?, og_image_url = ?, qr_code_enabled = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, profile.Slug, profile.DisplayName, profile.Bio, profile.AvatarURL,
		profile.HeaderImageURL, profile.ThemeID, profile.CustomCSS, profile.ShowUsernames,
		profile.IsPublic, profile.PasswordProtected, profile.ProtectionPassword,
		profile.CustomDomain, profile.DomainVerified, profile.AnalyticsEnabled,
		profile.MetaTitle, profile.MetaDescription, profile.OgImageURL,
		profile.QRCodeEnabled, profile.ID)
	return err
}

// DeleteProfile deletes a profile
func (db *DB) DeleteProfile(id string) error {
	_, err := db.Exec("DELETE FROM profiles WHERE id = ?", id)
	return err
}

// CountProfilesByUserID counts profiles for a user
func (db *DB) CountProfilesByUserID(userID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM profiles WHERE user_id = ?", userID).Scan(&count)
	return count, err
}

// IncrementProfileViewCount increments the view counter
func (db *DB) IncrementProfileViewCount(profileID string) error {
	_, err := db.Exec("UPDATE profiles SET view_count = view_count + 1 WHERE id = ?", profileID)
	return err
}

// ProfileTheme operations

// GetProfileTheme retrieves theme settings for a profile
func (db *DB) GetProfileTheme(profileID string) (*ProfileTheme, error) {
	theme := &ProfileTheme{}
	err := db.QueryRow(`
		SELECT profile_id, background_type, background_value, button_style,
		       button_animation, button_shadow, font_override, custom_css,
		       link_thumbnail_position, updated_at
		FROM profile_themes WHERE profile_id = ?
	`, profileID).Scan(
		&theme.ProfileID, &theme.BackgroundType, &theme.BackgroundValue,
		&theme.ButtonStyle, &theme.ButtonAnimation, &theme.ButtonShadow,
		&theme.FontOverride, &theme.CustomCSS, &theme.LinkThumbnailPosition,
		&theme.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return theme, nil
}

// UpdateProfileTheme updates or creates theme settings
func (db *DB) UpdateProfileTheme(theme *ProfileTheme) error {
	_, err := db.Exec(`
		INSERT INTO profile_themes (
			profile_id, background_type, background_value, button_style,
			button_animation, button_shadow, font_override, custom_css,
			link_thumbnail_position, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(profile_id) DO UPDATE SET
			background_type = EXCLUDED.background_type,
			background_value = EXCLUDED.background_value,
			button_style = EXCLUDED.button_style,
			button_animation = EXCLUDED.button_animation,
			button_shadow = EXCLUDED.button_shadow,
			font_override = EXCLUDED.font_override,
			custom_css = EXCLUDED.custom_css,
			link_thumbnail_position = EXCLUDED.link_thumbnail_position,
			updated_at = CURRENT_TIMESTAMP
	`, theme.ProfileID, theme.BackgroundType, theme.BackgroundValue,
		theme.ButtonStyle, theme.ButtonAnimation, theme.ButtonShadow,
		theme.FontOverride, theme.CustomCSS, theme.LinkThumbnailPosition)
	return err
}

// DeleteProfileTheme deletes theme settings
func (db *DB) DeleteProfileTheme(profileID string) error {
	_, err := db.Exec("DELETE FROM profile_themes WHERE profile_id = ?", profileID)
	return err
}

// QR Code Settings operations

// GetQRCodeSettings retrieves QR code settings
func (db *DB) GetQRCodeSettings(profileID string) (*QRCodeSettings, error) {
	settings := &QRCodeSettings{}
	err := db.QueryRow(`
		SELECT profile_id, size, error_correction, style, dark_color, light_color,
		       logo_enabled, logo_size, format, updated_at
		FROM qr_code_settings WHERE profile_id = ?
	`, profileID).Scan(
		&settings.ProfileID, &settings.Size, &settings.ErrorCorrection,
		&settings.Style, &settings.DarkColor, &settings.LightColor,
		&settings.LogoEnabled, &settings.LogoSize, &settings.Format,
		&settings.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

// UpdateQRCodeSettings updates or creates QR code settings
func (db *DB) UpdateQRCodeSettings(settings *QRCodeSettings) error {
	_, err := db.Exec(`
		INSERT INTO qr_code_settings (
			profile_id, size, error_correction, style, dark_color, light_color,
			logo_enabled, logo_size, format, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(profile_id) DO UPDATE SET
			size = EXCLUDED.size,
			error_correction = EXCLUDED.error_correction,
			style = EXCLUDED.style,
			dark_color = EXCLUDED.dark_color,
			light_color = EXCLUDED.light_color,
			logo_enabled = EXCLUDED.logo_enabled,
			logo_size = EXCLUDED.logo_size,
			format = EXCLUDED.format,
			updated_at = CURRENT_TIMESTAMP
	`, settings.ProfileID, settings.Size, settings.ErrorCorrection,
		settings.Style, settings.DarkColor, settings.LightColor,
		settings.LogoEnabled, settings.LogoSize, settings.Format)
	return err
}

// DeleteQRCodeSettings deletes QR code settings
func (db *DB) DeleteQRCodeSettings(profileID string) error {
	_, err := db.Exec("DELETE FROM qr_code_settings WHERE profile_id = ?", profileID)
	return err
}

// Service operations
// Per PART 36: 5000+ predefined services

// GetServiceByID retrieves a service by ID
func (db *DB) GetServiceByID(id string) (*Service, error) {
	service := &Service{}
	err := db.QueryRow(`
		SELECT id, name, category, icon_url, icon_svg, url_pattern,
		       background_color, text_color, popularity, is_active,
		       requires_username, placeholder_text, validation_pattern,
		       created_at, updated_at
		FROM services WHERE id = ?
	`, id).Scan(
		&service.ID, &service.Name, &service.Category, &service.IconURL,
		&service.IconSVG, &service.URLPattern, &service.BackgroundColor,
		&service.TextColor, &service.Popularity, &service.IsActive,
		&service.RequiresUsername, &service.PlaceholderText,
		&service.ValidationPattern, &service.CreatedAt, &service.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// GetServiceByName retrieves a service by name
func (db *DB) GetServiceByName(name string) (*Service, error) {
	service := &Service{}
	err := db.QueryRow(`
		SELECT id, name, category, icon_url, icon_svg, url_pattern,
		       background_color, text_color, popularity, is_active,
		       requires_username, placeholder_text, validation_pattern,
		       created_at, updated_at
		FROM services WHERE name = ? AND is_active = 1
	`, name).Scan(
		&service.ID, &service.Name, &service.Category, &service.IconURL,
		&service.IconSVG, &service.URLPattern, &service.BackgroundColor,
		&service.TextColor, &service.Popularity, &service.IsActive,
		&service.RequiresUsername, &service.PlaceholderText,
		&service.ValidationPattern, &service.CreatedAt, &service.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// ListServices retrieves services with optional category filter
func (db *DB) ListServices(category string, limit, offset int) ([]*Service, error) {
	var rows *sql.Rows
	var err error

	if category != "" {
		rows, err = db.Query(`
			SELECT id, name, category, icon_url, icon_svg, url_pattern,
			       background_color, text_color, popularity, is_active,
			       requires_username, placeholder_text, validation_pattern,
			       created_at, updated_at
			FROM services
			WHERE category = ? AND is_active = 1
			ORDER BY popularity DESC, name ASC
			LIMIT ? OFFSET ?
		`, category, limit, offset)
	} else {
		rows, err = db.Query(`
			SELECT id, name, category, icon_url, icon_svg, url_pattern,
			       background_color, text_color, popularity, is_active,
			       requires_username, placeholder_text, validation_pattern,
			       created_at, updated_at
			FROM services
			WHERE is_active = 1
			ORDER BY popularity DESC, name ASC
			LIMIT ? OFFSET ?
		`, limit, offset)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*Service
	for rows.Next() {
		service := &Service{}
		if err := rows.Scan(
			&service.ID, &service.Name, &service.Category, &service.IconURL,
			&service.IconSVG, &service.URLPattern, &service.BackgroundColor,
			&service.TextColor, &service.Popularity, &service.IsActive,
			&service.RequiresUsername, &service.PlaceholderText,
			&service.ValidationPattern, &service.CreatedAt, &service.UpdatedAt,
		); err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, rows.Err()
}

// SearchServices searches services by name
func (db *DB) SearchServices(query string, limit int) ([]*Service, error) {
	rows, err := db.Query(`
		SELECT id, name, category, icon_url, icon_svg, url_pattern,
		       background_color, text_color, popularity, is_active,
		       requires_username, placeholder_text, validation_pattern,
		       created_at, updated_at
		FROM services
		WHERE is_active = 1 AND (name LIKE ? OR category LIKE ?)
		ORDER BY popularity DESC, name ASC
		LIMIT ?
	`, "%"+query+"%", "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*Service
	for rows.Next() {
		service := &Service{}
		if err := rows.Scan(
			&service.ID, &service.Name, &service.Category, &service.IconURL,
			&service.IconSVG, &service.URLPattern, &service.BackgroundColor,
			&service.TextColor, &service.Popularity, &service.IsActive,
			&service.RequiresUsername, &service.PlaceholderText,
			&service.ValidationPattern, &service.CreatedAt, &service.UpdatedAt,
		); err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, rows.Err()
}

// CreateService creates a new service
func (db *DB) CreateService(service *Service) error {
	_, err := db.Exec(`
		INSERT INTO services (
			id, name, category, icon_url, icon_svg, url_pattern,
			background_color, text_color, popularity, is_active,
			requires_username, placeholder_text, validation_pattern,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, service.ID, service.Name, service.Category, service.IconURL,
		service.IconSVG, service.URLPattern, service.BackgroundColor,
		service.TextColor, service.Popularity, service.IsActive,
		service.RequiresUsername, service.PlaceholderText,
		service.ValidationPattern)
	return err
}

// UpdateService updates an existing service
func (db *DB) UpdateService(service *Service) error {
	_, err := db.Exec(`
		UPDATE services SET
			name = ?, category = ?, icon_url = ?, icon_svg = ?, url_pattern = ?,
			background_color = ?, text_color = ?, popularity = ?, is_active = ?,
			requires_username = ?, placeholder_text = ?, validation_pattern = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, service.Name, service.Category, service.IconURL, service.IconSVG,
		service.URLPattern, service.BackgroundColor, service.TextColor,
		service.Popularity, service.IsActive, service.RequiresUsername,
		service.PlaceholderText, service.ValidationPattern, service.ID)
	return err
}

// DeleteService deletes a service
func (db *DB) DeleteService(id string) error {
	_, err := db.Exec("DELETE FROM services WHERE id = ?", id)
	return err
}

// CountServices returns the total number of services
func (db *DB) CountServices() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM services WHERE is_active = 1").Scan(&count)
	return count, err
}

// Link operations
// Per PART 36: Links on user profiles

// GetLinkByID retrieves a link by ID
func (db *DB) GetLinkByID(id string) (*Link, error) {
	link := &Link{}
	err := db.QueryRow(`
		SELECT id, profile_id, service_id, title, username, url, icon_url,
		       background_color, text_color, position, is_active, click_count,
		       created_at, updated_at
		FROM links WHERE id = ?
	`, id).Scan(
		&link.ID, &link.ProfileID, &link.ServiceID, &link.Title,
		&link.Username, &link.URL, &link.IconURL, &link.BackgroundColor,
		&link.TextColor, &link.Position, &link.IsActive, &link.ClickCount,
		&link.CreatedAt, &link.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return link, nil
}

// GetLinksByProfileID retrieves all links for a profile
func (db *DB) GetLinksByProfileID(profileID string) ([]*Link, error) {
	rows, err := db.Query(`
		SELECT id, profile_id, service_id, title, username, url, icon_url,
		       background_color, text_color, position, is_active, click_count,
		       created_at, updated_at
		FROM links
		WHERE profile_id = ?
		ORDER BY position ASC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*Link
	for rows.Next() {
		link := &Link{}
		if err := rows.Scan(
			&link.ID, &link.ProfileID, &link.ServiceID, &link.Title,
			&link.Username, &link.URL, &link.IconURL, &link.BackgroundColor,
			&link.TextColor, &link.Position, &link.IsActive, &link.ClickCount,
			&link.CreatedAt, &link.UpdatedAt,
		); err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, rows.Err()
}

// CreateLink creates a new link
func (db *DB) CreateLink(link *Link) error {
	_, err := db.Exec(`
		INSERT INTO links (
			id, profile_id, service_id, title, username, url, icon_url,
			background_color, text_color, position, is_active, click_count,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, link.ID, link.ProfileID, link.ServiceID, link.Title, link.Username,
		link.URL, link.IconURL, link.BackgroundColor, link.TextColor,
		link.Position, link.IsActive, link.ClickCount)
	return err
}

// UpdateLink updates an existing link
func (db *DB) UpdateLink(link *Link) error {
	_, err := db.Exec(`
		UPDATE links SET
			service_id = ?, title = ?, username = ?, url = ?, icon_url = ?,
			background_color = ?, text_color = ?, position = ?, is_active = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, link.ServiceID, link.Title, link.Username, link.URL, link.IconURL,
		link.BackgroundColor, link.TextColor, link.Position, link.IsActive, link.ID)
	return err
}

// DeleteLink deletes a link
func (db *DB) DeleteLink(id string) error {
	_, err := db.Exec("DELETE FROM links WHERE id = ?", id)
	return err
}

// ReorderLinks updates link positions
func (db *DB) ReorderLinks(profileID string, linkIDs []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for position, linkID := range linkIDs {
		_, err := tx.Exec("UPDATE links SET position = ? WHERE id = ? AND profile_id = ?",
			position, linkID, profileID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// CountLinksByProfileID counts links for a profile
func (db *DB) CountLinksByProfileID(profileID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM links WHERE profile_id = ?", profileID).Scan(&count)
	return count, err
}

// IncrementLinkClickCount increments the click counter
func (db *DB) IncrementLinkClickCount(linkID string) error {
	_, err := db.Exec("UPDATE links SET click_count = click_count + 1 WHERE id = ?", linkID)
	return err
}

// FooterItem operations

// GetFooterItemsByProfileID retrieves footer items for a profile
func (db *DB) GetFooterItemsByProfileID(profileID string) ([]*FooterItem, error) {
	rows, err := db.Query(`
		SELECT id, profile_id, item_type, content, position, is_active, created_at
		FROM footer_items
		WHERE profile_id = ?
		ORDER BY position ASC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*FooterItem
	for rows.Next() {
		item := &FooterItem{}
		if err := rows.Scan(
			&item.ID, &item.ProfileID, &item.ItemType, &item.Content,
			&item.Position, &item.IsActive, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// CreateFooterItem creates a new footer item
func (db *DB) CreateFooterItem(item *FooterItem) error {
	_, err := db.Exec(`
		INSERT INTO footer_items (id, profile_id, item_type, content, position, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, item.ID, item.ProfileID, item.ItemType, item.Content, item.Position, item.IsActive)
	return err
}

// UpdateFooterItem updates an existing footer item
func (db *DB) UpdateFooterItem(item *FooterItem) error {
	_, err := db.Exec(`
		UPDATE footer_items SET
			item_type = ?, content = ?, position = ?, is_active = ?
		WHERE id = ?
	`, item.ItemType, item.Content, item.Position, item.IsActive, item.ID)
	return err
}

// DeleteFooterItem deletes a footer item
func (db *DB) DeleteFooterItem(id string) error {
	_, err := db.Exec("DELETE FROM footer_items WHERE id = ?", id)
	return err
}

// Shortlink operations
// Per PART 36: URL shortener functionality

// GetShortlinkByID retrieves a shortlink by ID
func (db *DB) GetShortlinkByID(id string) (*Shortlink, error) {
	shortlink := &Shortlink{}
	err := db.QueryRow(`
		SELECT id, short_code, target_url, profile_id, title, click_count, expires_at, created_at
		FROM shortlinks WHERE id = ?
	`, id).Scan(
		&shortlink.ID, &shortlink.ShortCode, &shortlink.TargetURL,
		&shortlink.ProfileID, &shortlink.Title, &shortlink.ClickCount,
		&shortlink.ExpiresAt, &shortlink.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return shortlink, nil
}

// GetShortlinkByCode retrieves a shortlink by short code
func (db *DB) GetShortlinkByCode(code string) (*Shortlink, error) {
	shortlink := &Shortlink{}
	err := db.QueryRow(`
		SELECT id, short_code, target_url, profile_id, title, click_count, expires_at, created_at
		FROM shortlinks WHERE short_code = ?
	`, code).Scan(
		&shortlink.ID, &shortlink.ShortCode, &shortlink.TargetURL,
		&shortlink.ProfileID, &shortlink.Title, &shortlink.ClickCount,
		&shortlink.ExpiresAt, &shortlink.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return shortlink, nil
}

// GetShortlinksByProfileID retrieves all shortlinks for a profile
func (db *DB) GetShortlinksByProfileID(profileID string) ([]*Shortlink, error) {
	rows, err := db.Query(`
		SELECT id, short_code, target_url, profile_id, title, click_count, expires_at, created_at
		FROM shortlinks
		WHERE profile_id = ?
		ORDER BY created_at DESC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shortlinks []*Shortlink
	for rows.Next() {
		shortlink := &Shortlink{}
		if err := rows.Scan(
			&shortlink.ID, &shortlink.ShortCode, &shortlink.TargetURL,
			&shortlink.ProfileID, &shortlink.Title, &shortlink.ClickCount,
			&shortlink.ExpiresAt, &shortlink.CreatedAt,
		); err != nil {
			return nil, err
		}
		shortlinks = append(shortlinks, shortlink)
	}

	return shortlinks, rows.Err()
}

// CreateShortlink creates a new shortlink
func (db *DB) CreateShortlink(shortlink *Shortlink) error {
	_, err := db.Exec(`
		INSERT INTO shortlinks (id, short_code, target_url, profile_id, title, click_count, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, shortlink.ID, shortlink.ShortCode, shortlink.TargetURL, shortlink.ProfileID,
		shortlink.Title, shortlink.ClickCount, shortlink.ExpiresAt)
	return err
}

// UpdateShortlink updates an existing shortlink
func (db *DB) UpdateShortlink(shortlink *Shortlink) error {
	_, err := db.Exec(`
		UPDATE shortlinks SET
			target_url = ?, title = ?, expires_at = ?
		WHERE id = ?
	`, shortlink.TargetURL, shortlink.Title, shortlink.ExpiresAt, shortlink.ID)
	return err
}

// DeleteShortlink deletes a shortlink
func (db *DB) DeleteShortlink(id string) error {
	_, err := db.Exec("DELETE FROM shortlinks WHERE id = ?", id)
	return err
}

// IncrementShortlinkClickCount increments the click counter
func (db *DB) IncrementShortlinkClickCount(id string) error {
	_, err := db.Exec("UPDATE shortlinks SET click_count = click_count + 1 WHERE id = ?", id)
	return err
}

// DeleteExpiredShortlinks removes expired shortlinks
func (db *DB) DeleteExpiredShortlinks() error {
	_, err := db.Exec("DELETE FROM shortlinks WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP")
	return err
}

// Analytics operations
// Per PART 36: Analytics tracking with hashed IPs for GDPR

// RecordProfileView records a profile view
func (db *DB) RecordProfileView(view *ProfileView) error {
	_, err := db.Exec(`
		INSERT INTO profile_views (profile_id, viewer_ip, referrer, user_agent, country, timestamp)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, view.ProfileID, view.ViewerIP, view.Referrer, view.UserAgent, view.Country)
	return err
}

// RecordLinkClick records a link click
func (db *DB) RecordLinkClick(click *LinkClick) error {
	_, err := db.Exec(`
		INSERT INTO link_clicks (link_id, clicker_ip, referrer, user_agent, country, timestamp)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, click.LinkID, click.ClickerIP, click.Referrer, click.UserAgent, click.Country)
	return err
}

// GetProfileAnalytics retrieves analytics for a profile
func (db *DB) GetProfileAnalytics(profileID string, days int) (*ProfileAnalytics, error) {
	analytics := &ProfileAnalytics{ProfileID: profileID}

	cutoffDate := time.Now().AddDate(0, 0, -days)

	err := db.QueryRow(`
		SELECT COUNT(*), COUNT(DISTINCT viewer_ip)
		FROM profile_views
		WHERE profile_id = ? AND timestamp >= ?
	`, profileID, cutoffDate).Scan(&analytics.Views, &analytics.UniqueIPs)
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM link_clicks
		WHERE link_id IN (SELECT id FROM links WHERE profile_id = ?)
		AND timestamp >= ?
	`, profileID, cutoffDate).Scan(&analytics.Clicks)
	if err != nil {
		return nil, err
	}

	topLinks, err := db.GetTopLinks(profileID, 10)
	if err == nil {
		analytics.TopLinks = topLinks
	}

	topReferrers, err := db.GetTopReferrers(profileID, 10)
	if err == nil {
		analytics.TopReferrers = topReferrers
	}

	return analytics, nil
}

// GetLinkAnalytics retrieves analytics for a link
func (db *DB) GetLinkAnalytics(linkID string, days int) (*LinkAnalytics, error) {
	analytics := &LinkAnalytics{LinkID: linkID}

	cutoffDate := time.Now().AddDate(0, 0, -days)

	err := db.QueryRow(`
		SELECT COUNT(*), COUNT(DISTINCT clicker_ip)
		FROM link_clicks
		WHERE link_id = ? AND timestamp >= ?
	`, linkID, cutoffDate).Scan(&analytics.Clicks, &analytics.UniqueIPs)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`
		SELECT referrer, COUNT(*) as count
		FROM link_clicks
		WHERE link_id = ? AND timestamp >= ? AND referrer != ''
		GROUP BY referrer
		ORDER BY count DESC
		LIMIT 10
	`, linkID, cutoffDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat ReferrerStat
			if err := rows.Scan(&stat.Referrer, &stat.Count); err == nil {
				analytics.TopReferrers = append(analytics.TopReferrers, &stat)
			}
		}
	}

	return analytics, nil
}

// GetTopLinks retrieves top links by clicks
func (db *DB) GetTopLinks(profileID string, limit int) ([]*LinkStat, error) {
	rows, err := db.Query(`
		SELECT l.id, l.title, COUNT(lc.id) as clicks
		FROM links l
		LEFT JOIN link_clicks lc ON l.id = lc.link_id
		WHERE l.profile_id = ?
		GROUP BY l.id, l.title
		ORDER BY clicks DESC
		LIMIT ?
	`, profileID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*LinkStat
	for rows.Next() {
		stat := &LinkStat{}
		if err := rows.Scan(&stat.LinkID, &stat.Title, &stat.Clicks); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

// GetTopReferrers retrieves top referrers
func (db *DB) GetTopReferrers(profileID string, limit int) ([]*ReferrerStat, error) {
	rows, err := db.Query(`
		SELECT pv.referrer, COUNT(*) as count
		FROM profile_views pv
		WHERE pv.profile_id = ? AND pv.referrer != ''
		GROUP BY pv.referrer
		ORDER BY count DESC
		LIMIT ?
	`, profileID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*ReferrerStat
	for rows.Next() {
		stat := &ReferrerStat{}
		if err := rows.Scan(&stat.Referrer, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

// Cluster operations
// Per PART 24: Cluster support with heartbeat monitoring

// CreateClusterNode creates a new cluster node
func (db *DB) CreateClusterNode(node *ClusterNode) error {
	_, err := db.Exec(`
		INSERT INTO cluster_nodes (id, hostname, address, port, status, is_primary, last_heartbeat, created_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, node.ID, node.Hostname, node.Address, node.Port, node.Status, node.IsPrimary)
	return err
}

// UpdateClusterNode updates a cluster node
func (db *DB) UpdateClusterNode(node *ClusterNode) error {
	_, err := db.Exec(`
		UPDATE cluster_nodes SET
			hostname = ?, address = ?, port = ?, status = ?, is_primary = ?, last_heartbeat = ?
		WHERE id = ?
	`, node.Hostname, node.Address, node.Port, node.Status, node.IsPrimary, node.LastHeartbeat, node.ID)
	return err
}

// GetClusterNode retrieves a cluster node by ID
func (db *DB) GetClusterNode(id string) (*ClusterNode, error) {
	node := &ClusterNode{}
	err := db.QueryRow(`
		SELECT id, hostname, address, port, status, is_primary, last_heartbeat, created_at
		FROM cluster_nodes WHERE id = ?
	`, id).Scan(
		&node.ID, &node.Hostname, &node.Address, &node.Port,
		&node.Status, &node.IsPrimary, &node.LastHeartbeat, &node.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// ListClusterNodes retrieves all cluster nodes
func (db *DB) ListClusterNodes() ([]*ClusterNode, error) {
	rows, err := db.Query(`
		SELECT id, hostname, address, port, status, is_primary, last_heartbeat, created_at
		FROM cluster_nodes
		ORDER BY is_primary DESC, created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*ClusterNode
	for rows.Next() {
		node := &ClusterNode{}
		if err := rows.Scan(
			&node.ID, &node.Hostname, &node.Address, &node.Port,
			&node.Status, &node.IsPrimary, &node.LastHeartbeat, &node.CreatedAt,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, rows.Err()
}

// UpdateNodeHeartbeat updates a node's heartbeat timestamp
func (db *DB) UpdateNodeHeartbeat(id string) error {
	_, err := db.Exec("UPDATE cluster_nodes SET last_heartbeat = CURRENT_TIMESTAMP, status = 'healthy' WHERE id = ?", id)
	return err
}

// DeleteClusterNode deletes a cluster node
func (db *DB) DeleteClusterNode(id string) error {
	_, err := db.Exec("DELETE FROM cluster_nodes WHERE id = ?", id)
	return err
}

// GetPrimaryNode retrieves the primary cluster node
func (db *DB) GetPrimaryNode() (*ClusterNode, error) {
	node := &ClusterNode{}
	err := db.QueryRow(`
		SELECT id, hostname, address, port, status, is_primary, last_heartbeat, created_at
		FROM cluster_nodes WHERE is_primary = 1
	`).Scan(
		&node.ID, &node.Hostname, &node.Address, &node.Port,
		&node.Status, &node.IsPrimary, &node.LastHeartbeat, &node.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// MarkNodeOffline marks a node as offline
func (db *DB) MarkNodeOffline(id string) error {
	_, err := db.Exec("UPDATE cluster_nodes SET status = 'offline' WHERE id = ?", id)
	return err
}

// Profile Tags operations

// AddProfileTag adds a tag to a profile
func (db *DB) AddProfileTag(profileID, tag string) error {
	_, err := db.Exec("INSERT OR IGNORE INTO profile_tags (profile_id, tag) VALUES (?, ?)", profileID, tag)
	return err
}

// RemoveProfileTag removes a tag from a profile
func (db *DB) RemoveProfileTag(profileID, tag string) error {
	_, err := db.Exec("DELETE FROM profile_tags WHERE profile_id = ? AND tag = ?", profileID, tag)
	return err
}

// GetProfileTags retrieves all tags for a profile
func (db *DB) GetProfileTags(profileID string) ([]string, error) {
	rows, err := db.Query("SELECT tag FROM profile_tags WHERE profile_id = ? ORDER BY tag", profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// SearchProfilesByTag searches profiles by tag
func (db *DB) SearchProfilesByTag(tag string, limit, offset int) ([]*Profile, error) {
	rows, err := db.Query(`
		SELECT p.id, p.user_id, p.slug, p.display_name, p.bio, p.avatar_url, p.header_image_url,
		       p.theme_id, p.custom_css, p.show_usernames, p.is_public, p.password_protected,
		       p.protection_password, p.custom_domain, p.domain_verified, p.analytics_enabled,
		       p.meta_title, p.meta_description, p.og_image_url, p.view_count, p.qr_code_enabled,
		       p.created_at, p.updated_at
		FROM profiles p
		INNER JOIN profile_tags pt ON p.id = pt.profile_id
		WHERE pt.tag = ? AND p.is_public = 1
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*Profile
	for rows.Next() {
		profile := &Profile{}
		if err := rows.Scan(
			&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
			&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
			&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
			&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
			&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
			&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
			&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, rows.Err()
}
