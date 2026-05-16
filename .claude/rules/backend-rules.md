# Backend Rules (PART 9, 10, 11, 32)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Never use bcrypt → Argon2id only
- Never store passwords/tokens in plaintext → Argon2id (passwords), SHA-256 (tokens)
- Never log sensitive fields (password_hash, two_factor_secret, tokens, smtp_password)
- Never return sensitive fields in API responses
- Never use mattn/go-sqlite3 (CGO) → use modernc.org/sqlite (pure Go)
- Never use lib/pq → use jackc/pgx/v5

## CRITICAL - ALWAYS DO
- Password hashing: Argon2id (golang.org/x/crypto/argon2)
- Token storage: SHA-256 hash only, never plaintext
- SQLite: modernc.org/sqlite (pure Go, CGO_ENABLED=0 compatible)
- Session tokens: random 32-byte values, server-side storage
- Session cookies: HttpOnly, Secure, SameSite=Lax
- Rate limiting on all auth endpoints (5 attempts/15min per IP)
- Identical error messages for wrong password vs unknown user

## DATABASE PATHS
| DB | Path |
|----|------|
| server.db | `/var/lib/casapps/cassocial/db/server.db` |
| users.db | `/var/lib/casapps/cassocial/db/users.db` |
| Docker | `/data/db/sqlite/` |

## TOR HIDDEN SERVICE (PART 32 - REQUIRED)
- Tor binary installed in container
- Server binary controls Tor startup (NOT container init)
- Auto-enabled when Tor binary found
- Tor address shown in startup banner and /server/healthz

## SECURITY RULES
- SSRF: validate URLs against RFC 1918 ranges before accepting
- XSS: HTML-escape all user-supplied content
- CSRF: token required on all state-mutating requests
- CSP: `default-src 'self'` on public profile pages
- Path traversal: serve static files from embedded FS only

---
For complete details, see AI.md PART 9, 10, 11, 32
