# Backend Rules (PART 9, 10, 11, 32)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Swallow errors, expose internal paths/secrets, or panic on request-controlled input.
- Store passwords in plaintext or use bcrypt for new hashes; use Argon2id.
- Store bearer tokens in plaintext; store token hashes using the specified SHA-256 format.
- Use external cron, invent a managed-node model, or disable required Tor/I2P support.

## CRITICAL - ALWAYS DO
- Return canonical, localized errors and validate all untrusted input.
- Initialize schema idempotently and preserve data across repeated startup/migration passes.
- Separate server-admin and regular-user records, sessions, and token ownership as specified.
- Emit structured audit/security events with the required source and fields; never include raw secrets.
- Support cluster nodes as instances of this application and use the built-in scheduler.
- Enable Tor automatically when detected and implement I2P with the specified runtime defaults.

## KEY RULES
- Apply rate limits and security controls at the server boundary and preserve request identity through in-process dispatch.
- Use the exact SQLite/database split, table prefixes, token scopes, and session cookies from AI.md.
- Keep cache failures and database failures within their documented fallback behavior.

For complete details, see AI.md PART 9, 10, 11, 32.
