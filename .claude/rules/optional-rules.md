# Optional Feature Rules (PART 34, 35, 36)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Activate an optional PART unless the project's `SPEC.md` declares it.
- Treat organizations or managed external nodes as cluster nodes.
- Weaken the baseline security, privacy, ownership, or validation rules for an optional feature.

## CRITICAL - ALWAYS DO
- Implement every declared optional PART fully, following its own data, API, UI, and security contract.
- Keep Server Admins distinct from Regular Users and enforce ownership at the API boundary.
- Validate custom-domain ownership and DNS safety before routing traffic or issuing certificates.

## PROJECT ACTIVATION
- `SPEC.md` activates PART 34 (Multi-User) and PART 36 (Custom Domains) as REQUIRED.
- PART 35 (Organizations) is not declared active; do not implement it unless SPEC.md changes.

For complete details, see AI.md PART 34-36.
