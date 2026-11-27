# Frontend Rules (PART 16, 17)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Use React/Vue/Angular rendering, SPA routing, or put business logic in JavaScript.
- Require JavaScript for core features or forms.
- Use desktop-first layouts, inline CSS/JavaScript, browser alerts, or placeholder/coming-soon pages.
- Let long strings break mobile layouts or omit required accessibility labels/status feedback.

## CRITICAL - ALWAYS DO
- Render pages and forms with server-side Go templates; use JavaScript only for progressive enhancement.
- Use mobile-first responsive CSS, accessible controls, and minimum 44x44px touch targets.
- Translate every user-facing string and validate all submitted values on the server.
- Source help/about content from IDEA.md and keep every admin setting and route functional.
- Keep all state-changing HTML forms CSRF-protected.

## PAGE CONTENT SOURCING
| Page | Content Source |
|---|---|
| `/server/about` | IDEA.md project description and implemented features |
| `/server/help` | IDEA.md, actual endpoints and examples |
| `/server/help/mcp` project section | IDEA.md Business logic; live MCP tool registry |

For complete details, see AI.md PART 16, 17.
