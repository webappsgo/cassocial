# Cassocial v1.0.0 - Complete Specification

## Project Overview

**Project Name**: cassocial  
**Organization**: casapps  
**Domain**: casjay.link  
**License**: MIT (LICENSE.md)  
**README**: README.md  
**Version**: 1.0.0  
**Description**: Self-hosted link aggregator and social profile landing page solution (like Linktree/Linkstack but self-hostable and enterprise-ready)  
**Embedded Licenses**: Added to LICENSE.md for any embedded code

## Core Principles

### Non-Negotiable Rules
- Target audience: Self-hosted, SMB, and enterprise (assume limited tech knowledge for first two)
- Validate everything, sanitize where appropriate
- Save only valid data, clear invalid
- Tokens and passwords shown once, must be copied
- Test everything applicable
- Show tooltips/documentation where needed
- Security and mobile first
- Set sane defaults for everything
- Security never impairs usability
- Questions asked, never assumed
- **Variables**: Anything wrapped in {} is a variable, everything else is literal
- **/etc/letsencrypt/live/domain**: This is a literal directory path, not a template

### Architecture
- **Platforms**: AMD64, ARM64, all OSes
- **Binary naming**: cassocial-{os}-{arch}
- **Installation**: System-wide preferred (/var/log, /etc/, /var/lib), fallback to user directories
- **Database**: SQLite (default), PostgreSQL, MariaDB/MySQL support
- **No configuration files**: Everything in database
- **Single binary**: All resources embedded

## Database Schema

### Core Tables

```sql
-- Users Table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL CHECK (length(username) BETWEEN 3 AND 30),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT CHECK (role IN ('admin', 'user', 'viewer')) DEFAULT 'user',
    status TEXT CHECK (status IN ('active', 'suspended', 'pending')) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    email_verified BOOLEAN DEFAULT false,
    two_factor_enabled BOOLEAN DEFAULT false,
    two_factor_secret TEXT,
    password_reset_token TEXT,
    password_reset_expires TIMESTAMP
);

-- Profiles Table
CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    slug TEXT UNIQUE NOT NULL,
    display_name TEXT CHECK (length(display_name) <= 100),
    bio TEXT CHECK (length(bio) <= 500),
    avatar_url TEXT,
    header_image_url TEXT,
    theme_id UUID DEFAULT '00000000-0000-0000-0000-000000000001',
    custom_css TEXT,
    show_usernames BOOLEAN DEFAULT true,
    is_public BOOLEAN DEFAULT true,
    password_protected BOOLEAN DEFAULT false,
    protection_password TEXT,
    custom_domain TEXT,
    domain_verified BOOLEAN DEFAULT false,
    analytics_enabled BOOLEAN DEFAULT true,
    meta_title TEXT CHECK (length(meta_title) <= 60),
    meta_description TEXT CHECK (length(meta_description) <= 160),
    og_image_url TEXT,
    view_count INTEGER DEFAULT 0,
    qr_code_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Links Table
CREATE TABLE links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    service_id UUID REFERENCES services(id),
    title TEXT CHECK (length(title) <= 100),
    username TEXT,
    url TEXT NOT NULL,
    icon_url TEXT,
    background_color TEXT,
    text_color TEXT,
    position INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true,
    click_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Services Table (5000+ predefined services)
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    category TEXT CHECK (category IN ('social', 'professional', 'development', 'content', 'payment', 'gaming', 'communication', 'portfolio', 'other')),
    icon_url TEXT,
    icon_svg TEXT,
    url_pattern TEXT,
    background_color TEXT,
    text_color TEXT,
    popularity INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    requires_username BOOLEAN DEFAULT true,
    placeholder_text TEXT,
    validation_pattern TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Settings Table (No config files - everything here)
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Themes Table
CREATE TABLE themes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    background_color TEXT DEFAULT '#1a1a1a',
    text_color TEXT DEFAULT '#ffffff',
    link_background TEXT DEFAULT '#2a2a2a',
    link_hover TEXT DEFAULT '#3a3a3a',
    link_text TEXT DEFAULT '#ffffff',
    border_radius TEXT DEFAULT '12px',
    font_family TEXT DEFAULT 'Inter, system-ui, sans-serif',
    is_premium BOOLEAN DEFAULT false,
    preview_image TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Analytics Tables
CREATE TABLE analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    link_id UUID REFERENCES links(id),
    event_type TEXT CHECK (event_type IN ('view', 'click')),
    ip_hash TEXT,
    user_agent TEXT,
    referrer TEXT,
    country TEXT,
    device_type TEXT CHECK (device_type IN ('mobile', 'tablet', 'desktop')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE analytics_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    session_id TEXT NOT NULL,
    ip_hash TEXT NOT NULL,
    country TEXT,
    region TEXT,
    city TEXT,
    device_type TEXT CHECK (device_type IN ('mobile', 'tablet', 'desktop')),
    browser TEXT,
    os TEXT,
    referrer_domain TEXT,
    referrer_path TEXT,
    utm_source TEXT,
    utm_medium TEXT,
    utm_campaign TEXT,
    landing_page TEXT,
    duration_seconds INTEGER,
    link_clicks INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE analytics_hourly (
    profile_id UUID,
    hour TIMESTAMP,
    views INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    total_clicks INTEGER DEFAULT 0,
    avg_duration_seconds INTEGER,
    top_referrer TEXT,
    top_country TEXT,
    PRIMARY KEY (profile_id, hour)
);

-- Footer Items
CREATE TABLE footer_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    item_type TEXT CHECK (item_type IN ('text', 'link', 'social_row', 'badge', 'html')),
    content JSONB NOT NULL,
    position INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(profile_id, position)
);

-- Profile Customization
CREATE TABLE profile_themes (
    profile_id UUID PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    background_type TEXT CHECK (background_type IN ('color', 'gradient', 'image')) DEFAULT 'color',
    background_value TEXT DEFAULT '#282a36',
    button_style TEXT CHECK (button_style IN ('rounded', 'square', 'pill')) DEFAULT 'rounded',
    button_animation TEXT CHECK (button_animation IN ('none', 'hover-lift', 'hover-glow')) DEFAULT 'hover-lift',
    button_shadow TEXT CHECK (button_shadow IN ('none', 'small', 'medium', 'large')) DEFAULT 'small',
    font_override TEXT,
    custom_css TEXT,
    link_thumbnail_position TEXT CHECK (link_thumbnail_position IN ('left', 'right', 'none')) DEFAULT 'left',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- QR Code Settings
CREATE TABLE qr_code_settings (
    profile_id UUID PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    size INTEGER DEFAULT 256 CHECK (size IN (128, 256, 512, 1024)),
    error_correction TEXT DEFAULT 'M' CHECK (error_correction IN ('L', 'M', 'Q', 'H')),
    style TEXT DEFAULT 'square' CHECK (style IN ('square', 'rounded', 'dots')),
    dark_color TEXT DEFAULT '#000000',
    light_color TEXT DEFAULT '#ffffff',
    logo_enabled BOOLEAN DEFAULT false,
    logo_size INTEGER DEFAULT 30,
    format TEXT DEFAULT 'png' CHECK (format IN ('png', 'svg', 'pdf')),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Import/Export Jobs
CREATE TABLE import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    source TEXT CHECK (source IN ('linktree', 'linkstack', 'carrd', 'aboutme', 'csv', 'json')),
    status TEXT CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    file_path TEXT,
    result JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- API Management
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    scopes TEXT[],
    rate_limit INTEGER DEFAULT 1000,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE api_webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    secret TEXT NOT NULL,
    events TEXT[] NOT NULL,
    active BOOLEAN DEFAULT true,
    failure_count INTEGER DEFAULT 0,
    last_triggered_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Social Features
CREATE TABLE profile_tags (
    profile_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (profile_id, tag)
);

CREATE TABLE featured_profiles (
    profile_id UUID PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    featured_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    featured_until TIMESTAMP,
    reason TEXT
);

CREATE TABLE profile_verification (
    profile_id UUID PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    verified BOOLEAN DEFAULT false,
    verification_type TEXT CHECK (verification_type IN ('email', 'domain', 'social', 'manual')),
    verification_data JSONB,
    verified_at TIMESTAMP,
    verified_by UUID REFERENCES users(id)
);

-- Organizations/Teams
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    owner_id UUID REFERENCES users(id),
    logo_url TEXT,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE organization_members (
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role TEXT CHECK (role IN ('owner', 'admin', 'editor', 'viewer')),
    invited_by UUID REFERENCES users(id),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (org_id, user_id)
);

CREATE TABLE organization_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    role TEXT CHECK (role IN ('admin', 'editor', 'viewer')),
    token TEXT UNIQUE NOT NULL,
    invited_by UUID REFERENCES users(id),
    expires_at TIMESTAMP,
    accepted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Content Moderation
CREATE TABLE blocked_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern TEXT NOT NULL,
    pattern_type TEXT CHECK (pattern_type IN ('domain', 'url', 'word')),
    reason TEXT,
    severity TEXT CHECK (severity IN ('warning', 'block')),
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE reported_content (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_type TEXT CHECK (content_type IN ('profile', 'link')),
    content_id UUID NOT NULL,
    reporter_ip_hash TEXT,
    reporter_email TEXT,
    reason TEXT CHECK (reason IN ('spam', 'inappropriate', 'phishing', 'copyright', 'other')),
    details TEXT,
    status TEXT CHECK (status IN ('pending', 'reviewing', 'resolved', 'dismissed')),
    moderator_id UUID REFERENCES users(id),
    moderator_notes TEXT,
    action_taken TEXT CHECK (action_taken IN ('none', 'warning', 'edited', 'suspended', 'deleted')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP
);

-- Shortlinks
CREATE TABLE shortlinks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_code TEXT UNIQUE NOT NULL,
    target_url TEXT NOT NULL,
    profile_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    title TEXT,
    click_count INTEGER DEFAULT 0,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Compliance & Privacy
CREATE TABLE user_consent (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    terms_version TEXT NOT NULL,
    terms_accepted_at TIMESTAMP NOT NULL,
    privacy_version TEXT NOT NULL,
    privacy_accepted_at TIMESTAMP NOT NULL,
    cookies_accepted BOOLEAN DEFAULT false,
    cookies_accepted_at TIMESTAMP,
    marketing_consent BOOLEAN DEFAULT false,
    marketing_consent_at TIMESTAMP,
    data_export_requested_at TIMESTAMP,
    deletion_requested_at TIMESTAMP,
    deletion_scheduled_for TIMESTAMP
);

CREATE TABLE data_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    status TEXT CHECK (status IN ('pending', 'processing', 'completed', 'expired')),
    file_path TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action TEXT NOT NULL,
    resource_type TEXT,
    resource_id UUID,
    ip_address TEXT,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Profile Maintenance
CREATE TABLE profile_maintenance (
    profile_id UUID PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    status TEXT CHECK (status IN ('active', 'maintenance', 'suspended')),
    message TEXT,
    bypass_token TEXT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    estimated_end TIMESTAMP
);
```

## Default Settings

```sql
-- System Configuration (stored in settings table)
INSERT INTO settings (key, value) VALUES
('site_name', 'Cassocial'),
('site_url', 'https://casjay.link'),
('initialized', 'false'),
('setup_completed', 'false'),
('maintenance_mode', 'false'),
('maintenance_message', 'We are upgrading our systems. Be back soon!'),
('maintenance_bypass_ips', '["127.0.0.1", "::1"]'),
('registration_enabled', 'true'),
('registration_requires_approval', 'false'),
('email_verification_required', 'true'),
('max_links_per_profile', '100'),
('max_profiles_per_user', '5'),
('upload_max_size_mb', '5'),
('allowed_image_types', '["jpg", "jpeg", "png", "gif", "webp"]'),
('cache_ttl_seconds', '3600'),
('rate_limit_requests', '100'),
('rate_limit_window_seconds', '60'),
('backup_enabled', 'true'),
('backup_retention_days', '30'),
('backup_time', '03:00'),
('analytics_retention_days', '90'),
('analytics_anonymous_mode', 'false'),
('analytics_sampling_rate', '100'),
('session_timeout_minutes', '1440'),
('password_min_length', '8'),
('password_require_uppercase', 'true'),
('password_require_number', 'true'),
('password_require_special', 'false'),
('two_factor_enabled', 'false'),
('qr_code_size', '256'),
('default_theme_id', '00000000-0000-0000-0000-000000000001'),
('enable_custom_css', 'true'),
('enable_custom_domains', 'true'),
('ssl_auto_renew', 'true'),
('ssl_renew_days_before', '30'),

-- SMTP Configuration
('smtp_provider', 'CUSTOM'),
('smtp_host', ''),
('smtp_port', '587'),
('smtp_security', 'STARTTLS'),
('smtp_user', ''),
('smtp_password', ''),
('smtp_from_name', 'Cassocial'),
('smtp_from_address', ''),
('admin_email', ''),
('smtp_enabled', 'false'),
('smtp_retry_count', '3'),
('smtp_retry_delay', '60'),

-- Notification Preferences
('notify_emergency', 'true'),
('notify_certificate', 'true'),
('notify_bug_report', 'true'),
('notify_user_registration', 'true'),
('notify_domain_verification', 'true'),
('notify_backup_status', 'true'),
('notify_high_traffic', 'true'),
('notification_batch_delay', '300'),

-- Footer Templates
('default_footer', '[
  {"type": "text", "content": "© 2025 {username}"},
  {"type": "links", "content": [
    {"text": "Privacy", "url": "/privacy"},
    {"text": "Terms", "url": "/terms"}
  ]}
]'),

-- Gradient Presets
('gradient_presets', '[
  {"name": "Sunset", "value": "linear-gradient(135deg, #667eea 0%, #764ba2 100%)"},
  {"name": "Ocean", "value": "linear-gradient(135deg, #667eea 0%, #764ba2 100%)"},
  {"name": "Forest", "value": "linear-gradient(135deg, #11998e 0%, #38ef7d 100%)"},
  {"name": "Fire", "value": "linear-gradient(135deg, #fc466b 0%, #3f5efb 100%)"}
]');
```

## First Run Flow

### Phase 1: Initial Detection
- Check if system initialized (`SELECT value FROM settings WHERE key = 'initialized'`)
- If not initialized, redirect all requests to `/setup`

### Phase 2: First User Registration/Login
- User creates account or logs in
- System detects this is the first user

### Phase 3: Administrator Account Creation
- First user forced to create separate admin account
- Default username: `administrator`
- Cannot be used for regular profiles
- Only resettable via CLI: `cassocial --reset-admin`

### Phase 4: Setup Wizard (As Administrator)

#### Step 1: Welcome
- Introduction to setup process
- Estimated time: 3-5 minutes

#### Step 2: Basic Configuration
- Instance name
- Instance URL (auto-detected)
- Support email
- Timezone
- Default language

#### Step 3: Domain & Access
- Profile URL structure (subdomain vs path)
- Custom domain support
- SSL configuration (Let's Encrypt, custom, or none)

#### Step 4: Email Configuration (Optional)
- Provider selection (CUSTOM, Gmail, Yahoo, Outlook)
- SMTP settings with security dropdown
- Port auto-fill with editable field
- Test connection button

#### Step 5: Features & Limits
- Registration settings
- Profile/link limits
- Feature toggles (analytics, QR codes, custom CSS, API)
- Privacy defaults

#### Step 6: Database Selection
- SQLite (default, zero-config)
- PostgreSQL (optional)
- MariaDB/MySQL (optional)
- Backup schedule configuration

#### Step 7: Review & Complete
- Configuration summary
- System initialization
- Redirect to admin dashboard

## CLI Commands

```bash
# Start server (default)
cassocial

# Version info
cassocial --version
cassocial -v

# Help
cassocial --help
cassocial -h

# Reset admin password (emergency only)
cassocial --reset-admin

# Run with custom settings
cassocial --port 3000
cassocial --host 0.0.0.0
cassocial -p 3000 -h 0.0.0.0
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `POST /api/auth/refresh` - Refresh token
- `POST /api/auth/forgot-password` - Request reset
- `POST /api/auth/reset-password` - Reset password
- `GET /api/auth/verify-email/{token}` - Verify email
- `POST /api/auth/2fa/enable` - Enable 2FA
- `POST /api/auth/2fa/verify` - Verify 2FA code

### Profiles
- `GET /api/profiles` - List user's profiles
- `POST /api/profiles` - Create profile
- `GET /api/profiles/{id}` - Get profile
- `PUT /api/profiles/{id}` - Update profile
- `DELETE /api/profiles/{id}` - Delete profile
- `POST /api/profiles/{id}/duplicate` - Clone profile
- `GET /api/profiles/{id}/qr` - Generate QR code
- `POST /api/profiles/{id}/verify-domain` - Verify custom domain

### Links
- `GET /api/profiles/{id}/links` - List links
- `POST /api/profiles/{id}/links` - Create link
- `PUT /api/links/{id}` - Update link
- `DELETE /api/links/{id}` - Delete link
- `POST /api/links/reorder` - Reorder links
- `POST /api/links/{id}/toggle` - Enable/disable

### Services
- `GET /api/services` - List all services
- `GET /api/services/search` - Search services
- `GET /api/services/categories` - List categories
- `GET /api/services/popular` - Popular services

### Analytics
- `GET /api/analytics/profile/{id}` - Profile stats
- `GET /api/analytics/links/{profile_id}` - Link stats
- `GET /api/analytics/export/{profile_id}` - Export data

### Admin
- `GET /api/admin/users` - List users
- `PUT /api/admin/users/{id}` - Update user
- `DELETE /api/admin/users/{id}` - Delete user
- `GET /api/admin/stats` - System statistics
- `POST /api/admin/backup` - Trigger backup
- `PUT /api/admin/settings` - Update settings
- `POST /api/admin/services/import` - Import services
- `POST /api/admin/cache/clear` - Clear cache
- `GET /api/admin/smtp/config` - Get SMTP config
- `PUT /api/admin/smtp/config` - Update SMTP config
- `POST /api/admin/smtp/test` - Test SMTP connection
- `GET /api/admin/notifications/preferences` - Get notification preferences
- `PUT /api/admin/notifications/preferences` - Update notification preferences

### Public API
- `GET /api/v1/profiles/{username}` - Public profile data
- `GET /api/v1/profiles/{username}/links` - Profile links
- `GET /api/v1/profiles/{username}/qr` - QR code

### Authenticated API
- `POST /api/v1/auth/token` - Get API token
- `GET /api/v1/me` - Current user info
- `CRUD /api/v1/profiles` - Manage profiles
- `CRUD /api/v1/links` - Manage links
- `GET /api/v1/analytics` - Analytics data

## Web UI Standards

### Layout & Responsive Design

#### Container Widths
- Desktop/Tablet (≥720px): 90% width (5% margins)
- Mobile (<720px): 98% width (1% margins)
- Footer: Always centered, always at bottom (scroll to see)

### Theme Specifications

#### Dark Theme (Default - Dracula-inspired)
```css
:root[data-theme="dark"] {
  --bg-primary: #282a36;
  --bg-secondary: #44475a;
  --bg-tertiary: #21222c;
  --text-primary: #f8f8f2;
  --text-secondary: #6272a4;
  --text-link: #8be9fd;
  --accent-primary: #bd93f9;
  --accent-success: #50fa7b;
  --accent-warning: #f1fa8c;
  --accent-danger: #ff5555;
  --accent-info: #8be9fd;
  --border: #6272a4;
  --shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
  --radius: 12px;
  --link-bg: #44475a;
  --link-hover: #6272a4;
  --link-text: #f8f8f2;
}
```

#### Light Theme
```css
:root[data-theme="light"] {
  --bg-primary: #ffffff;
  --bg-secondary: #f8f9fa;
  --bg-tertiary: #e9ecef;
  --text-primary: #212529;
  --text-secondary: #6c757d;
  --text-link: #0066cc;
  --accent-primary: #5e72e4;
  --accent-success: #2dce89;
  --accent-warning: #fb6340;
  --accent-danger: #f5365c;
  --accent-info: #11cdef;
  --border: #dee2e6;
  --shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
  --radius: 12px;
  --link-bg: #ffffff;
  --link-hover: #f8f9fa;
  --link-text: #212529;
  --link-border: 1px solid #dee2e6;
}
```

### Accessibility Requirements
- WCAG 2.1 Level AA compliance
- Semantic HTML structure
- ARIA labels on all interactive elements
- Keyboard navigation support
- Screen reader optimized
- Color contrast ratio: 4.5:1 minimum
- Focus indicators visible
- Skip navigation links
- Reduced motion support
- Minimum touch target: 44x44px
- 8px spacing between links

### Typography
- Font family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Inter", "Roboto", sans-serif
- Base size: 16px
- Line height: 1.6
- Maximum line length: 65ch for readability

### Logo & Favicon Handling
- **Remote source support**: Logos and favicons can be loaded from external URLs
- **Automatic scaling**: Images scaled to fit designated areas
- **Fallback support**: Default logo/favicon if remote fails
- **Supported formats**: PNG, JPG, SVG, ICO
- **Logo constraints**: max-width: 200px, max-height: 60px
- **Favicon sizes**: 16x16, 32x32, 180x180 (Apple Touch Icon)

## Directory Structure

### System Install
```
/etc/cassocial/
├── (no config files - all in database)

/var/lib/cassocial/
├── cassocial.db          # Main database
├── uploads/              # User uploads
│   ├── avatars/
│   └── headers/
└── backups/              # Auto backups

/var/log/cassocial/
└── cassocial.log         # Single log file

/usr/share/cassocial/
├── (resources embedded in binary)
```

### User Install
```
~/.local/share/cassocial/
├── cassocial.db
├── uploads/
└── backups/

~/.local/state/cassocial/
└── cassocial.log
```

### Portable Mode
```
./data/
├── cassocial.db
├── uploads/
├── backups/
└── cassocial.log
```

## Features

### Link Display Format
- **Default format**: `{username}@{Service}` (e.g., `casjay@GitHub`)
- **Multiple accounts per service**: Supports personal + organization accounts
- **Examples**:
  - `casjay@GitHub` (personal)
  - `casapps@GitHub` (organization)  
  - `personal@LinkedIn` / `business@LinkedIn`
- **Username toggle**: Hidden by default in settings
  - When ON: Shows `casjay@GitHub`
  - When OFF: Shows `GitHub` only
- **Blank usernames**: Default to service name with notice
- **Custom links**: Optional username support

### Profile Customization
- Background types: color, gradient, image
- Button styles: rounded, square, pill
- Button animations: none, hover-lift, hover-glow
- Custom CSS support (10KB limit)
- Google Fonts integration
- Theme presets: Sunset, Ocean, Forest, Fire

### Link Display
- Format: `{username}@{Service}` (toggleable)
- Multiple accounts per service supported
- Service database: 5000+ predefined services
- Categories: social, professional, development, content, payment, gaming, communication, portfolio
- Custom link support
- Drag & drop reordering

### Analytics
- Real-time visitor counter
- Geographic heat map
- Device/browser breakdown
- Referrer sources
- Peak traffic times
- Link click funnel
- Hourly aggregation
- Data retention: 90 days default
- Export formats: CSV, JSON, PDF

### QR Code Generation
- Sizes: 128, 256, 512, 1024
- Styles: square, rounded, dots
- Error correction levels: L, M, Q, H
- Custom colors
- Logo embedding option
- Formats: PNG, SVG, PDF

### Import/Export
- Import from: Linktree, Linkstack, Carrd, About.me, CSV, JSON
- Export to: JSON, CSV, HTML, PDF, vCard
- Bulk operations supported
- Merge or replace options

### Custom Domains
- CNAME support
- A/AAAA record support
- Automatic SSL via Let's Encrypt
- Domain verification via TXT record
- Multiple domains per instance
- Subdomain support: `{username}.casjay.link`

### Team/Organization Management
- Organization creation
- Role-based access: owner, admin, editor, viewer
- Invite system with email notifications
- Permissions matrix
- Activity audit log

### Social Features
- Public profile directory
- Profile verification system
- Featured profiles
- Trending profiles
- Tag-based discovery
- Search functionality

### Content Moderation
- Automated pattern blocking
- Report system
- Moderation queue
- Domain blacklist
- Word filters
- Action logging

### API System
- RESTful API
- API key authentication
- Rate limiting: 1000 req/hour
- Webhook support
- Event-driven notifications
- HMAC signature verification

### Shortlinks
- 6-character codes
- Custom codes supported
- Click tracking
- QR code generation
- Expiration support

### Footer Management
- Drag & drop editor
- Element types: text, links, social icons, badges, HTML
- Templates: minimal, standard, complete
- Position management

### Security
- Password requirements configurable
- Two-factor authentication
- Session management
- Rate limiting
- CORS handling
- CSP headers
- XSS protection
- SQL injection prevention
- Encrypted credential storage

### SMTP & Notifications

#### SMTP Configuration and Operational Rules

**Configuration Parameters**:
- `SMTP_HOST`: Hostname or IP address of SMTP server
- `SMTP_PORT`: Port number (commonly 25, 465, or 587)
- `SMTP_USER`: Username for authentication (optional if anonymous allowed)
- `SMTP_PASSWORD`: Password for SMTP_USER (required if user set)
- `SMTP_FROM_NAME`: Name used for outgoing emails
- `SMTP_FROM_ADDRESS`: Email address used as sender
- `ADMIN_EMAIL`: Recipient for system alerts and notifications

**Validation Rules**:
- SMTP_HOST, SMTP_PORT, and SMTP_FROM_ADDRESS are mandatory
- If SMTP_USER specified, SMTP_PASSWORD must be provided
- Email addresses must conform to standard formatting
- Configuration changes trigger validation and connection tests

**Connection and Security**:
- Support plain and TLS-secured SMTP connections
- SMTP authentication via PLAIN mechanism by default
- Support STARTTLS if advertised by server
- Credentials stored encrypted at rest

**Provider Configuration**:
- **Provider dropdown**: CUSTOM (default), Gmail, Yahoo, Outlook
- **Gmail**: Auto-fills `smtp.gmail.com`
- **Yahoo**: Auto-fills `smtp.mail.yahoo.com`  
- **Outlook**: Auto-fills `smtp-mail.outlook.com`

**Security/Port Dropdown** (Format: `PORT (SECURITY)`):
- `25 (NONE)` - Plain SMTP
- `587 (STARTTLS)` - STARTTLS (recommended)
- `465 (SSL/TLS)` - SSL/TLS
- `2525 (STARTTLS)` - Alternative STARTTLS
- Port remains editable after selection
- Security method stays locked to selection

#### Notification System Definitions

**Notification Types**:
- Emergency Alerts: Critical system failures, master key issues
- Certificate Renewal Alerts: SSL expiry warnings (7 days before)
- User Registration: New signups (if approval required)
- Profile Events: Suspension, deletion, reports
- System Events: Backups, updates, high traffic
- Domain Verification: Custom domain verification status

**Delivery Mechanism**:
- Primary delivery via email using SMTP settings
- Descriptive subject lines with priority indicators
- HTML email templates with responsive design
- Future extensibility for SMS, webhooks

**Operational Behavior**:
- Disable email alerting if SMTP invalid
- Retry policies based on priority (emergency: 5, high: 3, normal: 1)
- Log failures with severity levels
- Batch notifications within 5-minute windows

**Security Considerations**:
- Encrypt SMTP credentials using master key
- No plaintext password logging
- Secure credential update mechanisms
- Avoid sensitive information in notifications unless necessary

### Compliance & Privacy
- GDPR compliance
- Data export on request
- Account deletion
- Cookie consent management
- Audit logging
- Terms & Privacy versioning

### Progressive Web App
- Service worker for offline support
- App manifest
- Install prompts
- Push notifications ready
- Cache strategies

### Maintenance Mode
- Global maintenance toggle
- Per-profile maintenance
- Bypass IPs/tokens
- Custom messages
- Estimated completion time

## Database Support

### SQLite (Default)
- Zero configuration
- Embedded database
- Perfect for < 1000 users
- Automatic setup
- File-based backups

### PostgreSQL (Optional)
- Connection string: `postgres://user:pass@host:5432/cassocial`
- Connection pooling
- Advanced features
- Replication support

### MariaDB/MySQL (Optional)
- Connection string: `mysql://user:pass@host:3306/cassocial`
- UTF8MB4 charset
- InnoDB engine
- Master-slave ready

## Docker Configuration

### Docker Compose (Default - SQLite)
```yaml
version: '3.8'
services:
  cassocial:
    image: ghcr.io/casapps/cassocial
    container_name: cassocial
    volumes:
      - cassocial_data:/data
    ports:
      - "8080:8080"
    environment:
      - TZ=UTC
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  cassocial_data:
    driver: local
```

## Build Configuration

### Makefile Targets (Simplified)
- `build`: Build for all platforms (AMD64, ARM64) and create host binary named `cassocial`
- `release`: Release to GitHub
- `test`: Run all tests  
- `docker`: Build and push container to ghcr.io

### Binary Naming Scheme
- Pattern: `cassocial-{os}-{arch}` (e.g., `cassocial-linux-amd64`)
- Host binary: `cassocial` (set to current host architecture)

### Jenkins
- Agents: arm64, amd64
- Server: jenkins.casjay.cc

## Environment Variables

```bash
CASSOCIAL_PORT=8080
CASSOCIAL_HOST=0.0.0.0
CASSOCIAL_DATA=/path/to/data
CASSOCIAL_LOG_LEVEL=info
CASSOCIAL_MASTER_KEY=<generated>
CASSOCIAL_DATABASE_URL=sqlite:///path/to/cassocial.db
CASSOCIAL_SMTP_HOST=smtp.gmail.com
CASSOCIAL_SMTP_PORT=587
CASSOCIAL_SMTP_USER=user@gmail.com
CASSOCIAL_SMTP_PASSWORD=<encrypted>
CASSOCIAL_ADMIN_EMAIL=admin@example.com
```

## Service Categories

### Social Media (500+ services)
- Facebook, Instagram, Twitter/X, TikTok, Snapchat
- Mastodon instances, BlueSky, Threads
- Regional: VK, WeChat, Line, KakaoTalk

### Professional (200+ services)
- LinkedIn, Indeed, Glassdoor, AngelList
- Behance, Dribbble, DeviantArt
- ResearchGate, Academia.edu

### Development (300+ services)
- GitHub, GitLab, Bitbucket, Codeberg
- Stack Overflow, CodePen, JSFiddle
- npm, PyPI, Docker Hub

### Content (400+ services)
- YouTube, Vimeo, Twitch
- Spotify, SoundCloud, Bandcamp
- Medium, Substack, WordPress

### Payment (200+ services)
- PayPal, Venmo, Cash App
- Stripe, Square, Ko-fi
- Cryptocurrency wallets

### Gaming (150+ services)
- Steam, Discord, PlayStation, Xbox
- Epic Games, Battle.net
- Roblox, Minecraft

### Communication (100+ services)
- WhatsApp, Signal, Matrix
- Zoom, Skype, Jitsi

### Portfolio (100+ services)
- Personal website, Blog
- Resume/CV, Calendar
- RSS feed, Podcast

## Health & Monitoring

### Health Endpoints
- `/health` - Basic health check
- `/health/ready` - Readiness check
- `/health/live` - Liveness check

### Metrics
- Request count/duration
- Database query time
- Cache hit/miss ratio
- Active sessions
- Error rates

### Logging
- Levels: DEBUG, INFO, WARNING, ERROR, CRITICAL
- Single log file per instance
- Log rotation support
- Structured logging

## Performance Optimizations

### Caching
- Redis optional for sessions
- Page cache: 1 hour for public profiles
- Static assets: 1 year with versioning
- Service list: 24 hours
- API responses: 5 minutes

### Database
- Connection pooling (PostgreSQL/MariaDB)
- Prepared statements
- Strategic indexes
- Soft deletes
- Query optimization

### Assets
- Image resizing on upload
- WebP conversion
- Lazy loading
- Minified CSS/JS
- Gzip compression

## Security Defaults

### Headers
- X-Frame-Options: DENY
- X-Content-Type-Options: nosniff
- X-XSS-Protection: 1; mode=block
- Strict-Transport-Security: max-age=31536000
- Content-Security-Policy configured
- Referrer-Policy: strict-origin-when-cross-origin

### Rate Limiting
- Login: 5 attempts per 15 minutes
- API: 100 requests per minute
- Registration: 3 per hour per IP
- Password reset: 3 per hour

### Validation
- URL validation with protocol check
- Username: alphanumeric + underscore only
- Email: RFC 5322 compliant
- HTML sanitization
- SQL injection prevention
- XSS protection

## Testing Requirements

### Coverage Targets
- Unit tests: 80% minimum
- Integration tests: Critical paths
- E2E tests: User journeys

### Test Categories
- Authentication flows
- Profile CRUD operations
- Link management
- Analytics tracking
- Security validations
- API endpoints
- Rate limiting

## License

MIT License - See LICENSE.md

## Version

1.0.0 - Production Ready

---

All features specified and ready for implementation. No future features, no TODOs - everything included in v1.0.0.
