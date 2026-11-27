# Binary Rules (PART 7, 8, 33)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Enable CGO or build binaries with a host Go toolchain.
- Omit the required client binary or skip any supported OS/architecture.
- Change CLI flags, output formats, or exit behavior without following the specified contract.
- Print secrets, tokens, or sensitive config in CLI output.

## CRITICAL - ALWAYS DO
- Build with `CGO_ENABLED=0` in `casjaysdev/go:latest` and strip binaries per the build rules.
- Provide the server binary `cassocial` and required companion `cassocial-cli` for all eight targets.
- Support the documented flags, version/help output, `NO_COLOR`, and supported `table`, `json`, and `plain` output formats.
- Protect CLI config and token files with platform-appropriate permissions and ACL checks.

## KEY RULES
- Server CLI handles lifecycle and maintenance; the client handles API operations.
- Use runtime executable names and detected platform values rather than hardcoding development paths.
- Keep CLI routes and commands aligned with IDEA.md's API inventory.

For complete details, see AI.md PART 7, 8, 33.
