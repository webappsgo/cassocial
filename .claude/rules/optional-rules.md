# Optional Rules (PART 34-36)

⚠️ **These PARTs are OPTIONAL but NON-NEGOTIABLE WHEN IMPLEMENTED.** ⚠️

## STATUS FOR THIS PROJECT (cassocial)

Based on IDEA.md:
- **PART 34 (Multi-User)**: IMPLEMENTED - cassocial has user accounts, registration, profiles
- **PART 35 (Organizations)**: NOT YET IMPLEMENTED (viewer role planned but not surfaced)
- **PART 36 (Custom Domains)**: IMPLEMENTED - custom domain support per IDEA.md

## PART 34: MULTI-USER (ACTIVE)
- Regular Users separate from Server Admins (different DB tables)
- Registration with optional email verification
- User-owned profiles (multiple per user)
- Profile slug URL routing: `/{slug}`
- User dashboard at `/dashboard`
- 2FA (TOTP) support
- Password protection on individual profiles

## PART 35: ORGANIZATIONS (NOT YET ACTIVE)
- Not yet implemented
- Do NOT add org features without declaring in SPEC.md first

## PART 36: CUSTOM DOMAINS (ACTIVE)
- Profile owners can set custom_domain
- DNS TXT record verification required before domain_verified = true
- SSRF mitigation: resolve domain, reject RFC 1918 / loopback addresses
- Custom domain routing does NOT perform outbound HTTP to claimed domain

## ACTIVATION RULE
To activate an optional PART:
1. Change PART title from `OPTIONAL` → `NON-NEGOTIABLE` in AI.md
2. Update IDEA.md to document the feature as implemented
3. Follow ALL rules in that PART exactly

---
For complete details, see AI.md PART 34, 35, 36
