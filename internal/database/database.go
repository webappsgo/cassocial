package database

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// Note: embed paths would need to be set when building
// For now, migrations will be loaded from filesystem
var migrations embed.FS

type DB struct {
	*sql.DB
	Driver string
}

// Connect establishes a database connection based on the database URL
func Connect(databaseURL string) (*DB, error) {
	var driver string
	var dsn string

	if databaseURL == "" || strings.HasPrefix(databaseURL, "sqlite://") {
		driver = "sqlite3"
		if databaseURL == "" {
			// Default SQLite location
			dataDir := getDataDirectory()
			dbPath := filepath.Join(dataDir, "cassocial.db")
			dsn = dbPath
		} else {
			dsn = strings.TrimPrefix(databaseURL, "sqlite://")
		}
	} else if strings.HasPrefix(databaseURL, "postgres://") {
		driver = "postgres"
		dsn = databaseURL
	} else if strings.HasPrefix(databaseURL, "mysql://") {
		driver = "mysql"
		dsn = strings.TrimPrefix(databaseURL, "mysql://")
	} else {
		return nil, fmt.Errorf("unsupported database URL: %s", databaseURL)
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
	files, err := migrations.ReadDir("../../sql/migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		content, err := migrations.ReadFile("../../sql/migrations/" + file.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file.Name(), err)
		}

		// Adapt SQL for different databases
		sqlContent := string(content)
		sqlContent = db.adaptSQL(sqlContent)

		// Execute migration
		if _, err := db.Exec(sqlContent); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
		}
	}

	return nil
}

// adaptSQL adapts SQLite SQL to PostgreSQL or MySQL
func (db *DB) adaptSQL(sql string) string {
	switch db.Driver {
	case "postgres":
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
	if db.Driver == "postgres" {
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
