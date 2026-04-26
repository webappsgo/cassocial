# Cassocial Documentation

Welcome to the Cassocial documentation! Cassocial is a self-hosted link aggregator and social profile landing page platform.

## What is Cassocial?

Cassocial allows you to create beautiful, customizable profile pages with links to all your social media, websites, and content in one place. Think Linktree, but fully under your control with advanced features like:

- **Unlimited Profiles** - Create multiple profiles for different purposes
- **5000+ Services** - Auto-detected icons for popular platforms
- **Custom Domains** - Use your own domain for branded pages
- **Analytics** - Track views, clicks, and visitor insights
- **QR Codes** - Generate QR codes for easy sharing
- **Themes** - Beautiful dark/light themes with customization
- **Multi-User** - Support for teams and organizations
- **Privacy-Focused** - Self-hosted, GDPR compliant
- **API Access** - Full REST, GraphQL, and OpenAPI support

## Quick Start

### Docker (Recommended)

```bash
docker run -d \
  --name cassocial \
  -p 64580:80 \
  -v ./rootfs/config:/config:z \
  -v ./rootfs/data:/data:z \
  ghcr.io/casapps/cassocial:latest
```

Access the admin panel at `http://localhost:64580/admin`

### Features

- ✅ Self-hosted and private
- ✅ No tracking scripts or ads
- ✅ Full API access (REST + GraphQL)
- ✅ Import from Linktree, Linkstack, Carrd
- ✅ Export to JSON, CSV, HTML, vCard
- ✅ Mobile-first responsive design
- ✅ GDPR compliant

## Documentation

- [Installation](installation.md) - Setup and deployment
- [Configuration](configuration.md) - Configuration options
- [Admin Panel](admin.md) - Admin panel guide
- [API Reference](api.md) - API documentation
- [Development](development.md) - Development guide

## Support

- **Issues**: [GitHub Issues](https://github.com/casapps/cassocial/issues)
- **Repository**: [GitHub](https://github.com/casapps/cassocial)
- **License**: MIT
