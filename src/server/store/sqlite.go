package store

import (
	"crypto/rand"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// generateUUID returns a random hex string suitable for use as a row ID.
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck
	return fmt.Sprintf("%x", b)
}

// NewUUID is the exported form of generateUUID for use by other packages.
func NewUUID() string {
	return generateUUID()
}

// sqliteTimeFmt is SQLite's native datetime format for string comparison
// with datetime('now') and CURRENT_TIMESTAMP.
const sqliteTimeFmt = "2006-01-02 15:04:05"

// BindTime returns a value suitable for binding a time.Time in a prepared
// statement. For SQLite the driver stores time.Time as RFC3339 (with 'T'
// separator and 'Z' suffix), which SQLite's datetime() function cannot parse
// in all versions. Pre-formatting to SQLite's native space-separated format
// ensures that plain string comparison with datetime('now') works correctly.
// For other drivers (postgres, mysql), time.Time is returned as-is.
func (db *DB) BindTime(t time.Time) interface{} {
	if db.Driver == "sqlite" {
		return t.UTC().Format(sqliteTimeFmt)
	}
	return t
}

// BindNullableTime returns a value suitable for binding a *time.Time that
// may be nil (NULL in the database). Nil is passed through as nil. Non-nil
// values are formatted via BindTime.
func (db *DB) BindNullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return db.BindTime(*t)
}

// Rebind replaces all `?` placeholders with `$N` positional parameters when
// the driver requires it (postgres/pgx). For SQLite and MySQL `?` is kept as-is.
func (db *DB) Rebind(query string) string {
	if db.Driver != "postgres" && db.Driver != "pgx" {
		return query
	}
	var out []byte
	n := 1
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			out = append(out, fmt.Sprintf("$%d", n)...)
			n++
		} else {
			out = append(out, query[i])
		}
	}
	return string(out)
}

// ExecR runs Exec after rebinding `?` placeholders for the current driver.
func (db *DB) ExecR(query string, args ...interface{}) (sql.Result, error) {
	return db.Exec(db.Rebind(query), args...)
}

// QueryR runs Query after rebinding `?` placeholders for the current driver.
func (db *DB) QueryR(query string, args ...interface{}) (*sql.Rows, error) {
	return db.Query(db.Rebind(query), args...)
}

// QueryRowR runs QueryRow after rebinding `?` placeholders for the current driver.
func (db *DB) QueryRowR(query string, args ...interface{}) *sql.Row {
	return db.QueryRow(db.Rebind(query), args...)
}

//go:embed migrations/*.sql
var migrations embed.FS

// migrationsReadDir and migrationsReadFile are used by RunMigrations to read
// migration files from the embedded filesystem. Overridable in tests.
var migrationsReadDir = func(name string) ([]fs.DirEntry, error) {
	return migrations.ReadDir(name)
}
var migrationsReadFile = func(name string) ([]byte, error) {
	return migrations.ReadFile(name)
}

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
	files, err := migrationsReadDir("migrations")
	if err != nil {
		log.Printf("Could not read migrations directory: %v", err)
		return nil
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		content, err := migrationsReadFile("migrations/" + file.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file.Name(), err)
		}

		// Adapt SQL for different databases
		sqlContent := string(content)
		sqlContent = db.adaptSQL(sqlContent)

		// Execute migration
		if _, err := db.ExecR(sqlContent); err != nil {
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

// getEUID returns the effective user ID. Overridable in tests.
var getEUID = os.Geteuid

// getDataDirectory returns the appropriate data directory
func getDataDirectory() string {
	// Check for portable mode
	if _, err := os.Stat("./data"); err == nil {
		return "./data"
	}

	// Check if running as root/system
	if getEUID() == 0 {
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
	err := db.QueryRowR("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetSetting updates or inserts a setting
func (db *DB) SetSetting(key, value string) error {
	var query string
	if db.Driver == "mysql" {
		query = `
			INSERT INTO settings (key, value, updated_at)
			VALUES (?, ?, CURRENT_TIMESTAMP)
			ON DUPLICATE KEY UPDATE value = VALUES(value), updated_at = CURRENT_TIMESTAMP
		`
	} else {
		query = `
			INSERT INTO settings (key, value, updated_at)
			VALUES (?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
		`
	}

	_, err := db.ExecR(query, key, value)
	return err
}

// GetAllSettings retrieves all settings
func (db *DB) GetAllSettings() (map[string]string, error) {
rows, err := db.QueryR("SELECT key, value FROM settings")
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
err := db.QueryRowR(`
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
err := db.QueryRowR(`
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
err := db.QueryRowR(`
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
_, err := db.ExecR(`
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
_, err := db.ExecR(`
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
_, err := db.ExecR("DELETE FROM users WHERE id = ?", id)
return err
}

// ListUsers retrieves a paginated list of users
func (db *DB) ListUsers(limit, offset int) ([]*User, error) {
rows, err := db.QueryR(`
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
err := db.QueryRowR("SELECT COUNT(*) FROM users").Scan(&count)
return count, err
}

// Server Admin operations
// Per PART 23: Server Admins are separate from Regular Users

// GetServerAdminByID retrieves a server admin by ID
func (db *DB) GetServerAdminByID(id string) (*ServerAdmin, error) {
admin := &ServerAdmin{}
err := db.QueryRowR(`
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
err := db.QueryRowR(`
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
err := db.QueryRowR(`
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
_, err := db.ExecR(`
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
_, err := db.ExecR(`
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
err := db.QueryRowR("SELECT is_primary FROM server_admins WHERE id = ?", id).Scan(&isPrimary)
if err != nil {
return err
}

if isPrimary {
return fmt.Errorf("cannot delete primary admin")
}

_, err = db.ExecR("DELETE FROM server_admins WHERE id = ?", id)
return err
}

// GetPrimaryAdmin retrieves the primary admin
func (db *DB) GetPrimaryAdmin() (*ServerAdmin, error) {
admin := &ServerAdmin{}
err := db.QueryRowR(`
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
rows, err := db.QueryR(`
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
_, err := db.ExecR(`
INSERT INTO sessions (id, user_id, user_type, username, role, expires_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
`, session.ID, session.UserID, session.UserType, session.Username, session.Role, db.BindTime(session.ExpiresAt))
return err
}

// GetSession retrieves a session by ID
func (db *DB) GetSession(sessionID string) (*Session, error) {
session := &Session{}
err := db.QueryRowR(`
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
_, err := db.ExecR("DELETE FROM sessions WHERE id = ?", sessionID)
return err
}

// DeleteSessionsByUserID deletes all sessions for a user
func (db *DB) DeleteSessionsByUserID(userID string) error {
_, err := db.ExecR("DELETE FROM sessions WHERE user_id = ?", userID)
return err
}

// CleanupExpiredSessions removes expired sessions
func (db *DB) CleanupExpiredSessions() error {
_, err := db.ExecR("DELETE FROM sessions WHERE expires_at < datetime('now')")
return err
}

// Email verification operations
// Per PART 23: Email verification flow

// CreateEmailVerificationToken creates a verification token.
// token.TokenHash must be the SHA-256 hex digest of the raw token — never the raw token itself.
func (db *DB) CreateEmailVerificationToken(token *EmailVerificationToken) error {
_, err := db.ExecR(`
INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
VALUES (?, ?, ?)
`, token.UserID, token.TokenHash, db.BindTime(token.ExpiresAt))
return err
}

// GetEmailVerificationToken retrieves a verification token by its SHA-256 hash.
func (db *DB) GetEmailVerificationToken(tokenHash string) (*EmailVerificationToken, error) {
evt := &EmailVerificationToken{}
err := db.QueryRowR(`
SELECT id, token_hash, user_id, expires_at, created_at
FROM email_verification_tokens WHERE token_hash = ?
`, tokenHash).Scan(&evt.ID, &evt.TokenHash, &evt.UserID, &evt.ExpiresAt, &evt.CreatedAt)
if err != nil {
return nil, err
}
return evt, nil
}

// DeleteEmailVerificationToken deletes a verification token by its SHA-256 hash.
func (db *DB) DeleteEmailVerificationToken(tokenHash string) error {
_, err := db.ExecR("DELETE FROM email_verification_tokens WHERE token_hash = ?", tokenHash)
return err
}

// DeleteExpiredEmailVerificationTokens removes expired tokens
func (db *DB) DeleteExpiredEmailVerificationTokens() error {
_, err := db.ExecR("DELETE FROM email_verification_tokens WHERE expires_at < datetime('now')")
return err
}

// Password reset operations
// Per PART 23: Password reset flow

// CreatePasswordResetToken creates a password reset token
func (db *DB) CreatePasswordResetToken(token *PasswordResetToken) error {
_, err := db.ExecR(`
INSERT INTO password_reset_tokens (token, user_id, expires_at, created_at)
VALUES (?, ?, ?, CURRENT_TIMESTAMP)
`, token.Token, token.UserID, db.BindTime(token.ExpiresAt))
return err
}

// GetPasswordResetToken retrieves a password reset token
func (db *DB) GetPasswordResetToken(token string) (*PasswordResetToken, error) {
prt := &PasswordResetToken{}
err := db.QueryRowR(`
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
_, err := db.ExecR("DELETE FROM password_reset_tokens WHERE token = ?", token)
return err
}

// DeleteExpiredPasswordResetTokens removes expired tokens
func (db *DB) DeleteExpiredPasswordResetTokens() error {
_, err := db.ExecR("DELETE FROM password_reset_tokens WHERE expires_at < datetime('now')")
return err
}

// StoreBackupCodes deletes all existing backup codes for the user and inserts new ones
func (db *DB) StoreBackupCodes(userID string, codeHashes []string) error {
tx, err := db.Begin()
if err != nil {
return fmt.Errorf("failed to start transaction: %w", err)
}
defer tx.Rollback()

if _, err = tx.Exec(db.Rebind("DELETE FROM two_factor_backup_codes WHERE user_id = ?"), userID); err != nil {
return fmt.Errorf("failed to delete old backup codes: %w", err)
}

for _, hash := range codeHashes {
id := generateUUID()
if _, err = tx.Exec(
db.Rebind("INSERT INTO two_factor_backup_codes (id, user_id, code_hash, created_at) VALUES (?, ?, ?, ?)"),
id, userID, hash, db.BindTime(time.Now()),
); err != nil {
return fmt.Errorf("failed to insert backup code: %w", err)
}
}

return tx.Commit()
}

// GetUnusedBackupCodes returns all unused backup codes for a user
func (db *DB) GetUnusedBackupCodes(userID string) ([]BackupCode, error) {
rows, err := db.QueryR("SELECT id, user_id, code_hash, used_at, created_at FROM two_factor_backup_codes WHERE user_id = ? AND used_at IS NULL", userID)
if err != nil {
return nil, fmt.Errorf("failed to query backup codes: %w", err)
}
defer rows.Close()

var codes []BackupCode
for rows.Next() {
var bc BackupCode
if err := rows.Scan(&bc.ID, &bc.UserID, &bc.CodeHash, &bc.UsedAt, &bc.CreatedAt); err != nil {
return nil, fmt.Errorf("failed to scan backup code: %w", err)
}
codes = append(codes, bc)
}
return codes, rows.Err()
}

// MarkBackupCodeUsed marks a backup code as used
func (db *DB) MarkBackupCodeUsed(id string) error {
_, err := db.ExecR("UPDATE two_factor_backup_codes SET used_at = ? WHERE id = ?", db.BindTime(time.Now()), id)
return err
}

// DeleteBackupCodes removes all backup codes for a user
func (db *DB) DeleteBackupCodes(userID string) error {
_, err := db.ExecR("DELETE FROM two_factor_backup_codes WHERE user_id = ?", userID)
return err
}
