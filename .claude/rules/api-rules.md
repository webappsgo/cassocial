# API Rules (PART 13, 14, 15)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Never return 501 Not Implemented → implement fully or don't add the route
- Never expose /metrics publicly → internal only, optional bearer token
- Never redirect /healthz → /server/healthz (they must be separate handlers)

## CRITICAL - ALWAYS DO
- Health endpoint: `/server/healthz` (public, HTML/JSON/text content negotiation)
- API health: `/api/v1/server/healthz` (JSON by default)
- Metrics: `/metrics` (internal only, Prometheus format)
- API prefix: `/api/v1/`
- GraphQL: `/graphql`
- Swagger/OpenAPI: `/api/docs`

## ENDPOINT SUMMARY
| Endpoint | Access | Format |
|----------|--------|--------|
| `/server/healthz` | Public | HTML/JSON/text |
| `/api/v1/server/healthz` | Public | JSON |
| `/metrics` | Internal | Prometheus |
| `/api/v1/*` | Auth required | JSON |
| `/server/admin/*` | Admin only | HTML |

## SSL/TLS (PART 15)
- Auto-TLS via Let's Encrypt (ACME)
- Self-signed cert for development
- Cert storage: `{config_dir}/certs/`
- HTTP→HTTPS redirect when TLS enabled
- Failure mode: serve HTTP only, log error, notify admin

---
For complete details, see AI.md PART 13, 14, 15
