# API Rules (PART 13, 14, 15)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Keep obsolete project endpoints or add compatibility redirects/shims.
- Skip server-side validation, expose secrets, or return inconsistent success/error envelopes.
- Guess URL origins or trust forwarded headers from untrusted proxies.
- Weaken TLS verification or accept unsafe redirects/hostnames.

## CRITICAL - ALWAYS DO
- Use the configured API version and canonical route patterns.
- Keep health/version responses, API envelopes, status codes, generated OpenAPI, and route docs synchronized.
- Validate all parameters and request bodies on the server and enforce the route's authentication/authorization.
- Resolve public URLs with the specified per-request URL rules and trusted-proxy configuration.
- Use TLS/Let's Encrypt behavior exactly as specified and preserve safe certificate validation.

## KEY RULES
- API errors use the canonical `ok: false`, `error`, and `message` structure.
- MCP is an API-class POST endpoint on all three aliases; it uses existing Bearer tokens and does not accept cookies.
- Generated specs and llms.txt come from the same registered route definitions.

For complete details, see AI.md PART 13, 14, 15.
