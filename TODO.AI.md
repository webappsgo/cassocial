# AI implementation backlog

Every item is project implementation or verification work. Defects in the
AI.md template itself belong only in the parent `TODO.SERVER.md`.

## Foundation and backend

- [ ] Complete PARTS 0–8: project/spec structure, approved dependencies, all platform paths, configuration validation and secrets, metadata, logging, server/client flags, and output conventions.
- [ ] Complete PARTS 9–15: server lifecycle, API envelopes and validation, security headers, authentication/session/token storage, database schema initialization, networking, proxy/base-path behavior, TLS, and URL safety.
- [ ] Add the PART 12 `server.trusted_proxies` configuration and trust-gated forwarded host/proto/port/base-path/client-IP resolution; include its additional CIDR allow-list in `IsSecureRequest` and other proxy consumers.
- [ ] Implement PART 15 frontend CSRF protection for cookie-authenticated HTML mutations, including tokens in server-rendered forms and validation on every state-changing web route.
- [ ] Replace the current migration-file/versioning approach with the PART 10 idempotent schema bootstrap while preserving existing user data.
- [ ] Split SQLite persistence into `server.db` and `users.db` per PART 10, and apply the required `srv_`/`usr_` table prefixes in cluster databases.
- [ ] Add separate `admin_sessions` and `user_sessions`, distinct cookies, persistent session expiry/revocation, and `adm_`/`usr_` bearer-token validation.
- [ ] Finish the canonical route migration using configurable `server.api_version` and `server.admin_path`; remove non-spec legacy aliases and keep generated API specifications in sync.
- [ ] Add the IDEA.md API-route contract test that compares AI.md template routes, IDEA.md project routes, the router, and generated OpenAPI output.

## Pages and product features

- [ ] Complete PARTS 16–18: all server-rendered pages and forms work with JavaScript disabled; accessibility, themes, and account/admin navigation follow the templates.
- [ ] Complete PARTS 19–25: scheduler, backup/restore, email, GeoIP, metrics, updates, audit/security, cluster, and administration workflows.
- [ ] Complete PART 31: registry-driven i18n for every user-facing string, CLDR plural rules, locale selection and persistence, accessibility checks, and translated CLI/API/email/page content.
- [ ] Complete active PART 34: registration modes, user account lifecycle, profile privacy, user/session/token controls, and the full account-management pages and APIs.
- [ ] Complete active PART 36: custom-domain verification, DNS safety, routing, certificates, ownership controls, and admin/user management.

## Client

- [ ] Complete required PART 33 CLI behavior: protected cross-platform `cli.yml`, token login and revocation, command coverage, output formats and exit codes, TUI, API-version discovery, cluster failover, and verified self-update.
- [ ] Add platform-appropriate Windows ACL checks and tests for CLI config/token files.
- [ ] Keep client routes and command coverage aligned with the IDEA.md API inventory and generated API contract.

## Build, operations, and CI

- [ ] Complete PARTS 26–28: exact Makefile targets, all 8 OS/architecture builds for server and CLI, static CGO-free binaries, Docker images/compose files, runtime configuration, and container checks.
- [ ] Fix the current FreeBSD/amd64 build failure and rerun the complete all-platform build using the corrected in-container `go env` recipe.
- [ ] Complete PARTS 29–30: required tests, coverage gates, linters, security checks, all provider workflows, documentation build, and generated OpenAPI/GraphQL artifacts.
- [ ] Verify GitHub, GitLab, Gitea, Forgejo, and Jenkins workflows against AI.md, including their real CI runs after push and resolution of failures.

## Documentation and final verification

- [ ] Bring README, MkDocs pages, API/CLI/config/admin/security/integrations/development guides, translation guidance, and examples into agreement with the implemented routes and behavior.
- [ ] Ensure all runtime HTML pages, templates, error/empty states, and emails use the required translations and server-side validation.
- [ ] Complete every applicable FINAL CHECKPOINT item in AI.md and confirm IDEA.md, SPEC.md, rule files, and project memory are current.
- [ ] Run the prescribed Docker-based formatting, tests, coverage gate, lint/security checks, MkDocs build, Docker build, and all-platform build after integrating the final changes; fix every failure before committing.
