# AI implementation backlog

Every item is project implementation or verification work. Defects in the
AI.md template itself belong only in the parent `TODO.SERVER.md`.

## Foundation and backend

- [ ] Complete PARTS 0–8: project/spec structure, approved dependencies, all platform paths, configuration validation and secrets, metadata, logging, server/client flags, and output conventions.
- [ ] Complete PARTS 9–15: server lifecycle, REST/Swagger/GraphQL API parity, API envelopes and validation, security headers, authentication/session/token storage, database schema initialization, networking, proxy/base-path behavior, TLS, and URL safety.
- [x] Add the PART 12 `server.trusted_proxies` configuration and trust-gated forwarded host/proto/port/base-path/client-IP resolution; include its additional CIDR allow-list in `IsSecureRequest` and other proxy consumers.
- [x] Implement PART 16 frontend CSRF protection for every state-changing HTML form, including public forms; add tokens to server-rendered forms and validate mutations while excluding backend and Bearer requests.
- [x] Replace the migration-file runner with the PART 10 idempotent schema bootstrap; verify a stored user row survives a repeated schema pass.
- [x] Use the canonical SQLite `server.db` filename by default and preserve the previous `cassocial.db` contents through an online snapshot during first startup after upgrade.
- [ ] Split SQLite persistence into `server.db` and `users.db` per PART 10, and apply the required `srv_`/`usr_` table prefixes in cluster databases.
- [x] Add separate `admin_sessions` and `user_sessions`, distinct cookies, persistent session expiry/revocation, and `adm_`/`usr_` bearer-token validation.
- [ ] Finish the canonical route migration using configurable `server.api_version` and `server.admin_path`; remove non-spec legacy aliases and keep generated API specifications in sync.
- [ ] Complete PART 6 debug diagnostics: protect every `/debug/*` path with the generated `server.token`, redact command-line arguments, render themed error output, and verify diagnostics in debug and non-debug modes.
- [x] Add the IDEA.md API-route contract test that compares project routes with the router and generated OpenAPI output; generate OpenAPI with the configured API version.
- [ ] Complete GraphQL route/API synchronization, typed selectable read-result fields with reduced over-fetching, the themed same-origin GraphiQL assets and startup code, documentation, and end-to-end checks.
- [x] Match IDEA.md delete contracts: use path parameters for shortlinks, return `204` without bodies for profile/link/shortlink deletion, and let the CLI accept an empty `204` response.

## Pages and product features

- [ ] Complete PARTS 16–18: all server-rendered pages and forms work with JavaScript disabled; accessibility, themes, and account/admin navigation follow the templates.
- [ ] Implement the debug-only, operator-token-protected `/debug/components` gallery with every shared and project component state, both themes, both text directions, and browser layout coverage.
- [x] Implement the PART 16 guest `/server/preferences` and `/server/preferences/sync` pages and API mirrors, including consent-gated cookie writes, export/import, safe redirects, and reset behavior.
- [ ] Complete the MCP contract: scope-correct OpenAPI tool listing and dispatch, audit source, optional progress SSE, binary responses, help-page conformance, and operator editing/reload of `server.pages.help.mcp` from `yaml.d/pages.yml`.
- [x] Complete PART 12 analytics tracking providers, consent and Tor/I2P suppression, same-origin tracking script generation, CSP allow-listing, validation, and the admin tracking page.
- [x] Add optional GeoIP country attribution to public profile view and link-click analytics while retaining only the hashed visitor address.
- [ ] Complete PARTS 19–25: scheduler, backup/restore, email, GeoIP, metrics, updates, audit/security, cluster, and administration workflows.
- [x] Remove the unsupported GeoIP checksum gate; validate MMDB format, database type, build epoch, tree metadata, and the required public-address lookup, while reporting unavailable GeoIP databases as degraded.
- [ ] Complete PART 31: registry-driven i18n for every user-facing string, CLDR plural rules, locale selection and persistence, accessibility checks, and translated CLI/API/email/page content.
- [ ] Complete active PART 34: registration modes, user account lifecycle, profile privacy, user/session/token controls, OIDC/LDAP/SAML provider management and sign-in, and the first-login username-confirmation flow.
- [x] Store Regular User notification channel preferences as the PART 17 per-event JSON document, generate the no-JavaScript settings form from the event catalog, lock required security events, and honor server email-event defaults.
- [ ] Complete active PART 36: custom-domain verification, DNS safety, routing, certificates, ownership controls, and admin/user management.

## Client

- [ ] Complete required PART 33 CLI behavior: protected cross-platform `cli.yml`, token login and revocation, command coverage, output formats and exit codes, TUI, API-version discovery, cluster failover, and verified self-update.
- [x] Enforce the PART 33 canonical `--output` flag and the only supported `table`, `json`, and `plain` response formats in the CLI, tests, help, and CLI guide.
- [x] Add platform-appropriate Windows ACL checks and Windows-only tests for CLI config/token files; Windows test binary cross-compiles successfully.
- [ ] Keep client routes and command coverage aligned with the IDEA.md API inventory and generated API contract.

## Build, operations, and CI

- [ ] Complete PARTS 26–28: exact Makefile targets, all 8 OS/architecture builds for server and CLI, static CGO-free binaries, Docker images/compose files, runtime configuration, and container checks.
- [x] Cross-compile server and CLI for all eight OS/architecture combinations inside the Go container; FreeBSD/amd64 and FreeBSD/arm64 both build successfully.
- [ ] Complete PARTS 29–30: required tests, coverage gates, linters, security checks, all provider workflows, documentation build, and generated OpenAPI/GraphQL artifacts.
- [x] Add the required `go vet`, `staticcheck`, `govulncheck`, `shellcheck`, and `shfmt` gates to the Docker-based `make test` target, and format the test shell scripts; the latest full target passes with 64.7% coverage and all-platform/runtime checks.
- [ ] Add the PART 16 CSS architecture lint test over embedded project stylesheets using the specified pure-Go parser and all selector/token/layer rules.
- [x] Add the pure-Go parser dependency and verify every embedded project stylesheet parses as CSS; the full selector, token, and cascade-layer lint contract remains open above.
- [ ] Expand the standalone E2E suite to cover every IDEA.md user-facing feature in named SSR, no-JavaScript, and full-browser scenarios; add the required full crawl, admin/user flows, theme/language checks, form validation/CSRF, error pages, and static asset checks.
- [x] Add a Tier-1 SSR crawl from `/` that checks same-origin links and resources for HTTP failures and requires title, charset, viewport, and HTML document metadata on discovered pages; the crawl exposed and now covers the Swagger viewport metadata fix.
- [x] Check computed dark, light, and auto theme backgrounds in Chromium under both preferred color schemes.
- [x] Exercise the registered user's dashboard, profile editor, and profile-link management in all three E2E tiers; verify JavaScript-free login and full-browser console/request errors.
- [x] Exercise admin and user login/logout in SSR, no-JavaScript, and full-browser E2E tiers; verify embedded static asset routing/MIME, stale asset cache policy, and mobile header overflow.
- [x] Stub server outbound HTTP through a local `httptest` proxy and run the E2E containers on an internal-only Docker network; verify the browser suite passes without external network access.
- [x] Exercise user and admin notification-center empty states in SSR, verify the admin preferences page has its distinct canonical route, and cover owner-scoped notification history/read operations in handler and store tests.
- [ ] Verify GitHub, GitLab, Gitea, Forgejo, and Jenkins workflows against AI.md, including their real CI runs after push and resolution of failures.

## Documentation and final verification

- [ ] Bring README, MkDocs pages, API/CLI/config/admin/security/integrations/development guides, translation guidance, and examples into agreement with the implemented routes and behavior.
- [ ] Ensure all runtime HTML pages, templates, error/empty states, and emails use the required translations and server-side validation.
- [ ] Complete every applicable FINAL CHECKPOINT item in AI.md and confirm IDEA.md, SPEC.md, rule files, and project memory are current.
- [ ] Run the prescribed Docker-based formatting, tests, coverage gate, lint/security checks, MkDocs build, Docker build, and all-platform build after integrating the final changes; fix every failure before committing.
