# Cassocial — Claude Code Loader

**Spec:** Read `AI.md` (source of truth — read-only).
**Plan:** Read `IDEA.md` (project variables, description, business logic).
**Tasks:** Read `TODO.AI.md` when present.

## First-Turn Behavior

On every conversation start:
1. Read `IDEA.md` → resolve `{project_name}=cassocial`, `{project_org}=casapps`, `{internal_name}=cassocial`
2. Read active task files (`TODO.AI.md`, `PLAN.AI.md`) if present
3. Apply all rules in `AI.md` for the session

## Critical Rules (non-negotiable)

- `CGO_ENABLED=0` always — pure Go, no C
- Build source: `./src` (server), `./src/client` (CLI) — never `./src/main.go`
- All builds run inside `golang:alpine` Docker container via `GO_DOCKER`
- `OfficialSite` must be set in LDFLAGS for all builds
- Password hashing: Argon2id only
- Tokens stored as SHA-256 hashes — never plaintext
- No TODO/FIXME/HACK in committed code
- `make local` = local-platform build with version info (6 core Makefile targets only)
- `release.txt` = canonical version source (`1.0.0`)
- `src/client/` is REQUIRED — the CLI binary is mandatory
- All third-party GitHub Actions pinned to full 40-char commit SHA
