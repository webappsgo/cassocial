# Project Rules (PART 2, 3, 4)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Put source, docs, or runtime data in locations other than those defined by AI.md.
- Put the Dockerfile in the repository root; it belongs in `docker/`.
- Guess installation paths, names, ownership, or permissions when the spec or runtime can determine them.
- Remove required license notices or embed third-party content without documented provenance.

## CRITICAL - ALWAYS DO
- Preserve the project identity and required source-tree structure.
- Resolve runtime config, data, cache, log, backup, and PID paths using the OS-specific path rules.
- Keep source and documentation in the project; place runtime artifacts in a private project-scoped temp directory and clean them up.
- Record provenance before shipping third-party datasets, media, fonts, or translations.

## KEY RULES
- `IDEA.md` defines product behavior and project variables; AI.md defines implementation requirements.
- The server and required CLI are separate binaries; build them for Linux, macOS, Windows, and FreeBSD on amd64 and arm64.
- Keep generated and ignored artifacts out of the committed project tree.

For complete details, see AI.md PART 2, 3, 4.
