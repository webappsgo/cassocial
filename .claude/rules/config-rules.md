# Configuration Rules (PART 5, 6, 12)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Hardcode development, host, domain, credential, or environment-specific values.
- Add undocumented config aliases, obsolete keys, or settings for features that AI.md defines as always on.
- Log secrets or serialize secret values into public/runtime output.

## CRITICAL - ALWAYS DO
- Use the canonical `server.yml` and the exact `server.*` key names from AI.md.
- Validate configuration server-side, including ranges, paths, URLs, and cross-field constraints.
- Apply the specified environment/config/default resolution order and detect runtime modes rather than guessing.
- Keep sensitive values out of YAML output and protect configuration files and directories with required permissions.

## KEY RULES
- Resolve project paths from the configured directories and OS-specific conventions.
- Do not create an MCP enable switch; MCP is always available and uses existing Bearer authentication.
- Make configuration changes observable and persist them using the specified config area files.

For complete details, see AI.md PART 5, 6, 12.
