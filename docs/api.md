# API Reference

Cassocial provides three APIs:
- **REST API** at `/api/`
- **GraphQL** at `/graphql`
- **OpenAPI/Swagger** at `/openapi`

## Authentication

Authenticated endpoints require a Bearer token obtained from the login endpoint:

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  https://api.example.com/api/profiles
```

## REST API Endpoints

### Health

```
GET /health        # Basic health check
GET /health/ready  # Readiness check
GET /health/live   # Liveness check
```

Returns server health status as JSON.

### Auth

```
POST /api/auth/register                 # Register new account
POST /api/auth/login                    # Login (returns JWT)
POST /api/auth/login/2fa               # Login with 2FA code
POST /api/auth/forgot-password          # Request password reset
POST /api/auth/reset-password           # Reset password with token
GET  /api/auth/verify-email/{token}     # Verify email address

POST /api/auth/logout        # Logout (auth required)
POST /api/auth/refresh       # Refresh JWT token (auth required)
POST /api/auth/2fa/enable    # Enable 2FA (auth required)
POST /api/auth/2fa/verify    # Verify 2FA setup (auth required)
POST /api/auth/2fa/disable   # Disable 2FA (auth required)
```

### Profiles

```
GET    /api/profiles              # List profiles (auth required)
POST   /api/profiles              # Create profile (auth required)
GET    /api/profiles/{id}         # Get profile (auth required)
PUT    /api/profiles/{id}         # Update profile (auth required)
DELETE /api/profiles/{id}         # Delete profile (auth required)
POST   /api/profiles/{id}/duplicate      # Duplicate profile (auth required)
GET    /api/profiles/{id}/qr             # Generate QR code (auth required)
POST   /api/profiles/{id}/verify-domain  # Verify custom domain (auth required)
```

### Links

```
GET    /api/profiles/{id}/links   # List links for profile (auth required)
POST   /api/profiles/{id}/links   # Create link on profile (auth required)
PUT    /api/links/{id}            # Update link (auth required)
DELETE /api/links/{id}            # Delete link (auth required)
POST   /api/links/reorder         # Reorder links (auth required)
POST   /api/links/{id}/toggle     # Toggle link enabled/disabled (auth required)
```

### Services

```
GET /api/services             # List all supported services
GET /api/services/search      # Search services
GET /api/services/categories  # List service categories
GET /api/services/popular     # List popular services
GET /api/services/{id}        # Get specific service
```

### Analytics

```
GET /api/analytics/profile/{id}          # Get profile analytics (auth required)
GET /api/analytics/links/{profile_id}    # Get link analytics (auth required)
GET /api/analytics/export/{profile_id}   # Export analytics (auth required)
```

### Admin

```
GET    /api/admin/users/{id}                  # Get user (admin required)
GET    /api/admin/users                       # List users (admin required)
PUT    /api/admin/users/{id}                  # Update user (admin required)
DELETE /api/admin/users/{id}                  # Delete user (admin required)
GET    /api/admin/stats                       # System statistics (admin required)
POST   /api/admin/backup                      # Trigger backup (admin required)
GET    /api/admin/settings                    # Get settings (admin required)
PUT    /api/admin/settings                    # Update settings (admin required)
POST   /api/admin/services/import             # Import services (admin required)
POST   /api/admin/cache/clear                 # Clear cache (admin required)
GET    /api/admin/smtp/config                 # Get SMTP config (admin required)
PUT    /api/admin/smtp/config                 # Update SMTP config (admin required)
POST   /api/admin/smtp/test                   # Test SMTP connection (admin required)
GET    /api/admin/notifications/preferences   # Get notification prefs (admin required)
PUT    /api/admin/notifications/preferences   # Update notification prefs (admin required)
```

### Public API

No authentication required:

```
GET /api/v1/profiles/{username}        # Get public profile by username
GET /api/v1/profiles/{username}/links  # Get public profile links
GET /api/v1/profiles/{username}/qr     # Get public profile QR code
GET /api/v1/link/{id}/click            # Track link click
```

## GraphQL API

Access GraphiQL playground at `/graphql`

### Example Query

```graphql
query {
  profile(username: "myprofile") {
    id
    title
    description
    links {
      service
      url
      title
    }
  }
}
```

### Example Mutation

```graphql
mutation {
  createProfile(input: {
    username: "myprofile"
    title: "My Profile"
    description: "My awesome profile"
  }) {
    id
    username
  }
}
```

## OpenAPI/Swagger

Interactive API documentation available at `/openapi`

The OpenAPI specification is available at `/openapi.json`

## Rate Limiting

| Endpoint Type | Limit | Window |
|---------------|-------|--------|
| Authenticated API | 100 requests | 1 minute |
| Unauthenticated API | 20 requests | 1 minute |
| Login attempts | 5 attempts | 15 minutes |
| Registration | 5 attempts | 1 hour |

Rate limit exceeded returns `429 Too Many Requests` with `Retry-After` header.

## Error Responses

All errors return JSON:

```json
{
  "error": "Error message here",
  "code": "ERROR_CODE",
  "details": {}
}
```

### HTTP Status Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `429` - Too Many Requests
- `500` - Internal Server Error
- `503` - Service Unavailable
