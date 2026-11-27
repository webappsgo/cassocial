# AI Rules (PART 0, 1)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Guess or assume; read the relevant AI.md section or ask when product behavior is unresolved.
- Implement a feature before reading its applicable PART.
- Modify AI.md; it is read-only and authoritative.
- Add stubs, placeholders, TODO/FIXME/HACK comments, or partial implementations.
- Add unrequested features or change IDEA.md project variables without the user's confirmation.
- Treat PART 37 as implementation requirements; it is reference-only.

## CRITICAL - ALWAYS DO
- Read the relevant AI.md PART before implementation and reread it after compaction.
- Search AI.md for answers before asking the user.
- Keep code, docs, tests, IDEA.md, and generated API documentation aligned with AI.md.
- Follow exact project terminology, paths, defaults, validation, and security rules.
- Continue after cross-references and complete the referenced section before returning.
- Verify work against the FINAL CHECKPOINT and report only checks actually run.

## KEY RULES
- The server binary is `cassocial`; its required client is `cassocial-cli`; the agent is not implemented.
- Server Admins are application identities, distinct from OS users and Regular Users.
- All features are free; never add paid tiers or feature gates.
- All eight OS/architecture combinations are required; use CGO-free builds.

For complete details, see AI.md PART 0, 1.
