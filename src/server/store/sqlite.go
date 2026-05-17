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
	placeholder := "?, ?"
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
