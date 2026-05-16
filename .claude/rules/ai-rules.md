# AI Rules (PART 0, 1)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Never guess or assume - READ THE SPEC or ASK
- Never implement without reading relevant PART first
- Never modify AI.md PART content (read-only spec)
- Never add features not in spec without asking
- Never use "I think" or "probably" - KNOW from spec or ASK
- Never use generic placeholder content
- Never create /server/about or /server/help with placeholder text
- Never leave TODO comments in code - implement fully or don't implement
- Never create stub functions or "future" placeholders
- Never do partial implementations - every feature must be 100% complete

## CRITICAL - ALWAYS DO
- Read relevant PART before implementing ANY feature
- Search AI.md before asking questions (answer is likely there)
- Follow spec EXACTLY - no "improvements" without approval
- Update IDEA.md when features change
- Keep all docs in sync with code
- When unsure, ASK - never guess or assume
- Source /server/about and /server/help content from IDEA.md
- Implement features 100% complete - no stubs, no TODOs, no "future"
- ONE thing at a time - finish current task completely before starting another

## KEY DECISIONS (pre-answered)
| Question | Answer | Reference |
|----------|--------|-----------|
| What password hash? | Argon2id (NEVER bcrypt) | PART 11 |
| Where is Dockerfile? | `docker/Dockerfile` (NEVER root) | PART 27 |
| CGO enabled? | NEVER (CGO_ENABLED=0 always) | PART 7 |
| Premium features? | NEVER (all features free) | PART 1 |
| External cron? | NEVER (built-in scheduler) | PART 19 |
| Client-side rendering? | NEVER (server-side Go templates) | PART 16 |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| server | Main binary `cassocial` - runs as service |
| client | CLI binary `cassocial-cli` - REQUIRED |
| agent | Optional binary `cassocial-agent` |
| Server Admin | App administrator (NOT OS root) |
| Regular User | End-user (PART 34, optional feature) |

## COMPLIANCE CHECK
Before completing ANY task:
- [ ] Read relevant PART(s) in AI.md
- [ ] Implementation matches spec EXACTLY
- [ ] No guessing - all decisions from spec
- [ ] Docs updated if code changed

---
For complete details, see AI.md PART 0, 1
