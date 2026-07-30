module github.com/casapps/cassocial

go 1.25.0

require (
	github.com/go-sql-driver/mysql v1.8.1 // MySQL/MariaDB

	// Authentication
	github.com/golang-jwt/jwt/v5 v5.2.2 // JWT tokens
	github.com/google/uuid v1.6.0 // indirect; UUID generation
	github.com/jackc/pgx/v5 v5.9.2 // PostgreSQL - REQUIRED instead of lib/pq

	// Utilities
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e // QR code generation
	golang.org/x/crypto v0.51.0 // Argon2id password hashing

	// Core
	gopkg.in/yaml.v3 v3.0.1 // YAML config
	// Database drivers (NON-NEGOTIABLE: pure Go only, CGO_ENABLED=0)
	modernc.org/sqlite v1.34.5 // SQLite (pure Go) - REQUIRED instead of mattn/go-sqlite3
)

require golang.org/x/sys v0.45.0

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/text v0.39.0 // indirect
	modernc.org/libc v1.55.3 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.8.0 // indirect
)
