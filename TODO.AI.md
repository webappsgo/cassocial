# Cassocial Implementation TODO

**Current Status: Phase 1 Complete (Database Layer), Handlers Exist But Not Wired**

## ✅ Phase 1: Core Infrastructure - 100% COMPLETE

- [x] Configuration system (config.ParseBool, all options)
- [x] Server structure (proper package organization)
- [x] Database schema (388 lines, all PART 36 models)
- [x] Store interface (~150 methods)
- [x] SQLite implementation (1,671 lines, complete)
- [x] PostgreSQL/MySQL support (via adaptSQL)
- [x] Migration system (embedded)

## Phase 2-4: Critical Gap - Routes Not Wired

**Problem:** Handlers exist (10K+ lines) but routes.go only has 2 routes connected.

### PRIORITY 1: Wire Up All Routes (routes.go)

Must implement complete SetupRoutes() per PART 20 (API Structure):

Authentication Routes:
- [ ] POST /auth/register (handler exists: auth_handlers.go)
- [ ] POST /auth/login (handler exists)
- [ ] POST /auth/logout (handler exists)
- [ ] POST /auth/forgot-password (handler exists)
- [ ] POST /auth/reset-password (handler exists)
- [ ] GET /auth/verify-email (handler exists)

Public Routes (PART 17: Web Frontend):
- [ ] GET / (homepage - handler: public.go)
- [ ] GET /{slug} (profile view - handler: profile.go)
- [ ] GET /qr/{slug} (QR code - handler: qr.go)
- [ ] GET /s/{code} (shortlink redirect - handler: shortlink.go)

API Routes - Profiles (handler: profile_handlers.go):
- [ ] GET /api/v1/profiles
- [ ] POST /api/v1/profiles
- [ ] GET /api/v1/profiles/{id}
- [ ] PUT /api/v1/profiles/{id}
- [ ] DELETE /api/v1/profiles/{id}

API Routes - Links (handler: link_handlers.go):
- [ ] GET /api/v1/profiles/{id}/links
- [ ] POST /api/v1/profiles/{id}/links
- [ ] PUT /api/v1/links/{id}
- [ ] DELETE /api/v1/links/{id}
- [ ] POST /api/v1/links/reorder

API Routes - Services (handler: service_handlers.go):
- [ ] GET /api/v1/services
- [ ] GET /api/v1/services/{id}
- [ ] GET /api/v1/services/search

API Routes - Analytics (handler: analytics_handlers.go):
- [ ] GET /api/v1/profiles/{id}/analytics
- [ ] GET /api/v1/links/{id}/analytics

API Routes - Shortlinks (handler: shortlink.go):
- [ ] POST /api/v1/shortlinks
- [ ] GET /api/v1/shortlinks
- [ ] DELETE /api/v1/shortlinks/{id}

Admin Routes (handler: admin_handlers.go):
- [ ] GET /admin
- [ ] GET /admin/users
- [ ] POST /admin/users
- [ ] GET /admin/settings
- [ ] POST /admin/settings
- [ ] POST /admin/backup
- [ ] POST /admin/restore

API Documentation:
- [ ] GET /openapi (Swagger UI)
- [ ] GET /openapi.json (OpenAPI spec)
- [ ] GET /graphql (GraphQL playground)

### PRIORITY 2: Verify Handler Implementations

Check each handler file against PART 20 & PART 36:
- [ ] auth_handlers.go - registration, login, password reset
- [ ] profile_handlers.go - profile CRUD
- [ ] link_handlers.go - link CRUD and reordering
- [ ] service_handlers.go - service listing
- [ ] analytics_handlers.go - analytics aggregation
- [ ] admin_handlers.go - admin panel operations
- [ ] public_handlers.go - public pages
- [ ] setup.go - first-run setup wizard

### PRIORITY 3: Frontend Templates (PART 17)

Templates needed in src/server/template/:
- [ ] layout/base.html
- [ ] layout/admin.html
- [ ] page/login.html
- [ ] page/register.html
- [ ] page/profile.html
- [ ] page/dashboard.html
- [ ] page/admin.html
- [ ] partial/header.html
- [ ] partial/footer.html
- [ ] partial/link_card.html

Static assets in src/server/static/:
- [ ] css/main.css
- [ ] css/admin.css
- [ ] css/themes.css
- [ ] js/main.js
- [ ] js/admin.js
- [ ] images/ (icons, logos)

### PRIORITY 4: Testing (PART 13)

Unit Tests:
- [ ] config package tests
- [ ] store package tests (all CRUD operations)
- [ ] auth tests (password hashing, sessions)
- [ ] model tests (validation)

Integration Tests:
- [ ] API endpoint tests
- [ ] Database tests
- [ ] Auth flow tests

End-to-End Tests:
- [ ] User registration → profile creation → link management
- [ ] Admin user management
- [ ] Analytics tracking

### PRIORITY 5: Additional Features

Setup Wizard:
- [ ] First-run detection
- [ ] Primary Admin creation
- [ ] Setup token generation
- [ ] Initial configuration

Session Management:
- [ ] Session cleanup cron job
- [ ] Token expiration handling
- [ ] "Remember me" functionality

Email System (PART 26):
- [ ] SMTP configuration
- [ ] Email templates
- [ ] Verification emails
- [ ] Password reset emails

### PRIORITY 6: API Documentation

OpenAPI/Swagger:
- [ ] Generate OpenAPI spec from code
- [ ] Swagger UI setup
- [ ] Theme support (light/dark)
- [ ] Authentication in Swagger

GraphQL:
- [ ] Schema definition
- [ ] Resolvers implementation
- [ ] GraphiQL playground
- [ ] Theme support

### PRIORITY 7: Advanced Features

Custom Domains (PART 35):
- [ ] DNS verification
- [ ] Let's Encrypt integration
- [ ] Domain routing

Analytics (PART 29):
- [ ] GeoIP integration (ip-location-db)
- [ ] IP hashing for GDPR
- [ ] Analytics dashboard
- [ ] Export functionality

Clustering (PART 24):
- [ ] Heartbeat system
- [ ] Primary election
- [ ] Config sync
- [ ] Node management

QR Codes:
- [ ] QR generation library integration
- [ ] Customization settings
- [ ] Multiple format support

Import/Export:
- [ ] Linktree import
- [ ] Linkstack import
- [ ] Profile export
- [ ] Bulk link import

### PRIORITY 8: Production Readiness

Docker (PART 14):
- [ ] Verify Dockerfile completeness
- [ ] Test docker-compose setup
- [ ] Test multi-stage build
- [ ] Verify entrypoint.sh

CI/CD (PART 15):
- [ ] GitHub Actions workflows
- [ ] All 8 platform builds
- [ ] Docker image publishing
- [ ] Release automation

Documentation (PART 33):
- [ ] docs/installation.md
- [ ] docs/configuration.md
- [ ] docs/api.md
- [ ] docs/admin.md
- [ ] docs/development.md
- [ ] mkdocs.yml configuration

## Summary

**Completed:** Phase 1 (Database Layer) - 1,671 lines, all Store methods
**Existing:** Handlers (10K+ lines), Auth (Argon2id), Middleware
**Critical Gap:** routes.go not wired (only 2/50+ routes connected)
**Estimate:** 60-70% infrastructure done, needs wiring + frontend + tests

**Next Critical Steps:**
1. Complete routes.go (wire all existing handlers)
2. Implement frontend templates
3. Add comprehensive tests
4. Verify handler completeness

**Following AI.md:**
- PART 0: AI Assistant Rules (re-read before implementing)
- PART 20: API Structure (route patterns)
- PART 17: Web Frontend (templates)
- PART 13: Testing
- No drift, 100% spec compliance
