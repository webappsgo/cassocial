# Frontend Rules (PART 16, 17)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Never client-side rendering (React, Vue, Angular, etc.)
- Never require JavaScript for core functionality
- Never use inline CSS or JavaScript
- Never use JavaScript alerts → use toast notifications
- Never use generic placeholder content in /server/about or /server/help
- Never create stub templates or "coming soon" pages
- Never create empty handlers or placeholder routes
- Never use CDN scripts in HTML → bundle all assets at build time

## CRITICAL - ALWAYS DO
- Server-side rendering (Go templates)
- Progressive enhancement (works without JS)
- Mobile-first responsive CSS
- CSS `word-break: break-all` for long strings (IPv6, .onion, tokens)
- Dark mode default, light/auto support
- Full admin panel at `/server/admin/` with ALL settings
- WCAG 2.1 AA accessibility
- Touch targets minimum 44x44px
- /server/about content from IDEA.md
- /server/help with real endpoints and real examples

## BREAKPOINTS (mobile-first)
| Target | CSS |
|--------|-----|
| Mobile (base) | No media query |
| Tablet+ | `@media (min-width: 768px)` |
| Desktop+ | `@media (min-width: 1024px)` |

## LONG STRINGS (REQUIRED CSS)
```css
.long-string, .ip-address, .onion-address, .api-token, .hash {
  word-break: break-all;
  overflow-wrap: break-word;
  font-family: monospace;
}
```

## SERVER VS CLIENT
| Task | Where |
|------|-------|
| Data validation | SERVER |
| HTML rendering | SERVER |
| Business logic | SERVER |
| Theme toggle | Client JS |
| Copy to clipboard | Client JS |

---
For complete details, see AI.md PART 16, 17
