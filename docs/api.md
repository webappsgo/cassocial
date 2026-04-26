# API Reference

Cassocial provides three APIs:
- **REST API** at `/api/v1/`
- **GraphQL** at `/graphql`
- **OpenAPI/Swagger** at `/openapi`

## Authentication

API endpoints require Bearer token authentication:

```bash
curl -H "Authorization: Bearer YOUR_API_TOKEN" \
  https://api.example.com/api/v1/profiles
```

Generate API tokens in the admin panel at `/admin/api-keys`.

## REST API Endpoints

### Health

```
GET /healthz
GET /api/v1/healthz
```

Returns server health status.

### Profiles

```
GET    /api/v1/profiles          # List public profiles
GET    /api/v1/profiles/{slug}   # Get profile by slug
POST   /api/v1/profiles          # Create profile (auth required)
PUT    /api/v1/profiles/{id}     # Update profile (auth required)
DELETE /api/v1/profiles/{id}     # Delete profile (auth required)
```

### Links

```
GET    /api/v1/links?profile_id={id}  # Get links for profile
POST   /api/v1/links                  # Create link (auth required)
PUT    /api/v1/links/{id}             # Update link (auth required)
DELETE /api/v1/links/{id}             # Delete link (auth required)
POST   /api/v1/links/reorder          # Reorder links (auth required)
```

### Analytics

```
GET /api/v1/analytics?profile_id={id}&days=30  # Get profile analytics
GET /api/v1/analytics/link?link_id={id}       # Get link analytics
GET /api/v1/analytics/export?profile_id={id}&format=csv  # Export analytics
```

### Services

```
GET /api/v1/services             # List all supported services
GET /api/v1/services/search?q={query}  # Search services
```

### QR Codes

```
GET /api/v1/qr?url={url}&size=256&format=png  # Generate QR code
GET /api/v1/qr/profile?slug={slug}            # Profile QR code
```

### Themes

```
GET /api/v1/themes               # List available themes
GET /api/v1/themes/{id}          # Get specific theme
POST /api/v1/themes              # Save custom theme (auth required)
```

### Import/Export

```
POST /api/v1/import              # Import data (auth required)
GET  /api/v1/export?profile_id={id}&format=json  # Export profile
```

### Shortlinks

```
POST   /api/v1/shortlinks        # Create shortlink (auth required)
GET    /api/v1/shortlinks/{code} # Get shortlink info
DELETE /api/v1/shortlinks/{code} # Delete shortlink (auth required)
GET    /s/{code}                 # Redirect to target URL
```

## GraphQL API

Access GraphiQL playground at `/graphql`

### Example Query

```graphql
query {
  profile(slug: "myprofile") {
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
    slug: "myprofile"
    title: "My Profile"
    description: "My awesome profile"
  }) {
    id
    slug
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
