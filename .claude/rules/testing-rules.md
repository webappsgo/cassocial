# Testing Rules (PART 29, 30, 31)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Run Go builds, tests, or binaries on the host; use the approved Docker/Incus environment.
- Treat a successful build or test run as proof that the application is complete.
- Leave docs, translations, accessibility states, or platform workflows unverified.
- Claim checks that were not run or hide their actual failure status.

## CRITICAL - ALWAYS DO
- Run the required unit, integration, E2E, coverage, lint, security, documentation, container, and platform checks.
- Use Docker for the Go toolchain and the specified Incus OS tests for service lifecycle.
- Build strict MkDocs documentation and keep examples/routes current.
- Translate every user-facing string into every supported locale and verify locale rendering, placeholders, and accessibility.
- Use the prescribed test/lint wrapper so gate results reflect child-command exit status.

## KEY RULES
- The coverage gate is at least 60%; test tiers include named server-rendered, no-JavaScript, and browser scenarios.
- AI beta testing aims to find edge cases, then fix and reverify them.

For complete details, see AI.md PART 29, 30, 31.
