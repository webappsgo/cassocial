# Makefile Rules (PART 26)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Use Makefile targets in CI/CD workflows.
- Run Go builds or tests on the host.
- Hardcode project identity, version, build timestamp, or platform values.

## CRITICAL - ALWAYS DO
- Keep Makefile commands for local development and run Go commands in `casjaysdev/go:latest` with CGO disabled.
- Resolve project variables from IDEA.md, then use the documented fallback order.
- Provide the required build, test, Docker, cleanup, and development targets with bounded container resources.
- Keep all eight platform builds available and use explicit linker/build metadata.

## KEY RULES
- `make test` enforces the coverage threshold and runs the prescribed Docker/runtime tiers.
- CI reproduces the actual commands explicitly and never calls `make`.

For complete details, see AI.md PART 26.
