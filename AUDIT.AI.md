# Project Audit

Started: 2026-06-04

Auditor: spec compliance pass against AI.md (canonical spec).

Severity legend: **CRITICAL** = spec violation that breaks documented behavior or security guarantees; **HIGH** = spec violation, ships broken/stub output to users; **MEDIUM** = missing infrastructure required by spec; **LOW** = cleanup.

## Pass 1: Security
- [x] `src/server/template/layout/base.html` / `admin.html` / `page/setup.html`: PART 17 forbids inline `<script>`. Extracted theme-init JS to `src/server/static/js/theme-init.js` and replaced inline blocks with `<script src=...>` — FIXED.

## Pass 2: Code Quality
- [x] `src/server/handlers.go.old`, `src/server/templates.go.old`: dead `.old` source files committed. Deleted — FIXED.
- [ ] `src/server/template/page/setup.html` and other templates contain `style="..."` inline-attribute CSS (PART 17: "Never use inline CSS"). MEDIUM — full template audit and migration to `setup.css` / utility classes pending; not blocking compile.

## Pass 3: Logic and Correctness
- [x] `src/server/password.go` `GenerateEmailVerificationToken`: was reusing `password_reset_token` column with `EMAIL_` prefix. Fixed with dedicated `email_verification_tokens` table (migration 001 updated, migration 005 added). Token stored as SHA-256 hash. — FIXED.
- [x] `src/server/password.go` `ForcePasswordChange`: arg order bug (userID, time passed in wrong position). Fixed. — FIXED.

## Pass 4: Documentation Completeness
- [x] `tests/` directory: `tests/run_tests.sh`, `tests/docker.sh`, `tests/incus.sh` created — FIXED.
- [ ] `docs/` exists but not verified against PART 30 required pages (index, installation, configuration, api, admin, development). **LOW** — spot check pending.

## Pass 5: Spec and Rules Compliance

### PART 0/1 — No stubs in production code (CRITICAL/HIGH)
- [ ] `src/ssl/ssl.go:83` `getLetsEncryptConfig()` returns `errors.New("Let's Encrypt support coming soon")`. PART 15 lists Auto-TLS via Let's Encrypt (ACME) as required. **CRITICAL** — feature is wired into config but unimplemented; server silently falls back to HTTP. Requires adding `golang.org/x/crypto/acme/autocert` (pure-Go, CGO-free) and persisting certs under `{config_dir}/ssl/letsencrypt/`.
- [ ] `src/server/service/qr_service.go:208-212` `generatePDF()` returns `[]byte("%PDF-1.4\n%placeholder for QR code PDF\n")`. Spec (PART 0: "no stub functions") and CLAUDE.md "QR Code Generation → Formats: PNG, SVG, PDF" require a real PDF. **HIGH** — `/api/profiles/{id}/qr?format=pdf` ships an invalid PDF. Fix: bundle a pure-Go PDF writer (e.g. `github.com/phpdave11/gofpdf`) — pure-Go, CGO-free.
- [ ] `src/server/service/export_service.go:238-256` `exportToPDF()` returns plain text masquerading as a PDF. **HIGH** — same fix as above.
- [x] `src/server/service/import_service.go` `importFromCarrd()`: implemented real Carrd HTML parser using `regexp` (no new deps). — FIXED.
- [x] `src/common/i18n/locales/ar.json`, `ja.json`: created with all 75 keys. All 7 PART 31 locale files complete. — FIXED.
- [x] `src/server/i18n.go`: rewritten to use `//go:embed` via `src/common/i18n` package; per-key fallback to default language. — FIXED.

### PART 28 — CI/CD providers (CRITICAL)
- [x] `.github/workflows/build.yml`, `release.yml`, `security.yml` — created with SHA-pinned actions, truffleHog, govulncheck. — FIXED.
- [x] `.gitea/workflows/build.yml`, `release.yml`, `security.yml` — same. — FIXED.
- [x] `.forgejo/workflows/build.yml`, `release.yml`, `security.yml` — same. — FIXED.
- [ ] `.gitlab-ci.yml` and `Jenkinsfile` — not audited for SHA-pinning and truffleHog usage. **MEDIUM**.

### PART 29 — Testing infrastructure
- [x] `tests/run_tests.sh`, `tests/docker.sh`, `tests/incus.sh` created. — FIXED.

### PART 31 — i18n locale completeness
- [x] All 7 locale files complete (en, de, es, fr, zh, ar, ja) with 75 keys each. — FIXED.

### PART 32 — Tor hidden service (verify)
- [x] `src/server/tor.go` exists with tests. Wiring into startup banner/healthz not re-verified in this pass.

## Pass 6: Code Flow Trace

- [ ] `src/ssl/ssl.go` `GetTLSConfig()` is reachable from server bootstrap and will hard-fail if `LetsEncrypt=true` is ever set in config because of the stub above. Currently masked because no code path sets `LetsEncrypt=true` — spec requires it to work.
- [ ] `src/server/service/qr_service.go` `Generate(format="pdf")` flows directly to the placeholder bytes — surfaced via API endpoint `GET /api/profiles/{id}/qr` and admin UI.
- [ ] `src/server/service/export_service.go` `Export(format="pdf")` flows to placeholder.

## Outstanding (ordered by priority)

1. **CRITICAL**: Replace Let's Encrypt stub with real `autocert` implementation (`src/ssl/ssl.go`).
2. **HIGH**: Replace QR PDF placeholder with real PDF (`src/server/service/qr_service.go`).
3. **HIGH**: Replace export-to-PDF placeholder with real PDF (`src/server/service/export_service.go`).
4. **MEDIUM**: Audit `.gitlab-ci.yml` and `Jenkinsfile` for SHA-pinning and truffleHog.
5. **MEDIUM**: Audit `docs/` pages against PART 30 required set.
6. **LOW**: Migrate template `style="..."` inline attributes to external CSS classes.

## Completed
- Removed inline `<script>` blocks; added `static/js/theme-init.js`.
- Deleted `src/server/handlers.go.old` and `src/server/templates.go.old`.
- Dedicated `email_verification_tokens` table (migration 001 updated, migration 005 added).
- `ForcePasswordChange` arg order bug fixed.
- `importFromCarrd()` real HTML parser implemented.
- All 7 PART 31 locale files (ar, de, en, es, fr, ja, zh) created with 75 keys each.
- `src/server/i18n.go` rewritten with embed + per-key fallback.
- `tests/run_tests.sh`, `tests/docker.sh`, `tests/incus.sh` created.
- `.github/`, `.gitea/`, `.forgejo/` workflow trios created (SHA-pinned, truffleHog).
