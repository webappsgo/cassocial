# Docker Rules (PART 27)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Put `Dockerfile` in the repository root.
- Run the application as root or bake secrets, host paths, or development values into an image.
- Use mutable/unverified build inputs or omit health/runtime checks required by AI.md.

## CRITICAL - ALWAYS DO
- Keep the Dockerfile under `docker/` and use the specified multi-stage, CGO-free build.
- Run with an unprivileged user, explicit volumes/config, minimal runtime dependencies, and the required health check.
- Provide the required Compose/runtime configuration and validate the built image in an isolated container.
- Keep runtime data and logs in mounted application paths with correct ownership and permissions.

## KEY RULES
- Tor and i2pd are the only documented overlay daemons; the server must still run without them.
- Keep image defaults aligned with the bare-metal/container mode and I2P activation rules.

For complete details, see AI.md PART 27.
