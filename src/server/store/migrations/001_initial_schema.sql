-- Cassocial v1.0.0 - Initial Database Schema
-- This schema supports SQLite, PostgreSQL, and MariaDB/MySQL

-- Users Table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    username TEXT UNIQUE NOT NULL CHECK (length(username) BETWEEN 3 AND 30),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT CHECK (role IN ('admin', 'user', 'viewer')) DEFAULT 'user',
    status TEXT CHECK (status IN ('active', 'suspended', 'pending')) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    email_verified BOOLEAN DEFAULT 0,
    two_factor_enabled BOOLEAN DEFAULT 0,
    two_factor_secret TEXT,
    password_reset_token TEXT,
    password_reset_expires TIMESTAMP
);

-- Profiles Table
CREATE TABLE IF NOT EXISTS profiles (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slug TEXT UNIQUE NOT NULL,
    display_name TEXT CHECK (length(display_name) <= 100),
    bio TEXT CHECK (length(bio) <= 500),
    avatar_url TEXT,
    header_image_url TEXT,
    theme_id TEXT DEFAULT '00000000-0000-0000-0000-000000000001',
    custom_css TEXT,
    show_usernames BOOLEAN DEFAULT 1,
    is_public BOOLEAN DEFAULT 1,
    password_protected BOOLEAN DEFAULT 0,
    protection_password TEXT,
    custom_domain TEXT,
    domain_verified BOOLEAN DEFAULT 0,
    analytics_enabled BOOLEAN DEFAULT 1,
    meta_title TEXT CHECK (length(meta_title) <= 60),
    meta_description TEXT CHECK (length(meta_description) <= 160),
    og_image_url TEXT,
    view_count INTEGER DEFAULT 0,
    qr_code_enabled BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Services Table (5000+ predefined services)
CREATE TABLE IF NOT EXISTS services (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT UNIQUE NOT NULL,
    category TEXT CHECK (category IN ('social', 'professional', 'development', 'content', 'payment', 'gaming', 'communication', 'portfolio', 'other')),
    icon_url TEXT,
    icon_svg TEXT,
    url_pattern TEXT,
    background_color TEXT,
    text_color TEXT,
    popularity INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT 1,
    requires_username BOOLEAN DEFAULT 1,
    placeholder_text TEXT,
    validation_pattern TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Links Table
CREATE TABLE IF NOT EXISTS links (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    service_id TEXT REFERENCES services(id),
    title TEXT CHECK (length(title) <= 100),
    username TEXT,
    url TEXT NOT NULL,
    icon_url TEXT,
    background_color TEXT,
    text_color TEXT,
    position INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT 1,
    click_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Settings Table (No config files - everything here)
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Themes Table
CREATE TABLE IF NOT EXISTS themes (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT NOT NULL,
    background_color TEXT DEFAULT '#1a1a1a',
    text_color TEXT DEFAULT '#ffffff',
    link_background TEXT DEFAULT '#2a2a2a',
    link_hover TEXT DEFAULT '#3a3a3a',
    link_text TEXT DEFAULT '#ffffff',
    border_radius TEXT DEFAULT '12px',
    font_family TEXT DEFAULT 'Inter, system-ui, sans-serif',
    is_premium BOOLEAN DEFAULT 0,
    preview_image TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Analytics Tables
CREATE TABLE IF NOT EXISTS analytics (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    profile_id TEXT REFERENCES profiles(id) ON DELETE CASCADE,
    link_id TEXT REFERENCES links(id),
    event_type TEXT CHECK (event_type IN ('view', 'click')),
    ip_hash TEXT,
    user_agent TEXT,
    referrer TEXT,
    country TEXT,
    device_type TEXT CHECK (device_type IN ('mobile', 'tablet', 'desktop')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS analytics_sessions (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    profile_id TEXT REFERENCES profiles(id) ON DELETE CASCADE,
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

CREATE TABLE IF NOT EXISTS analytics_hourly (
    profile_id TEXT NOT NULL,
    hour TIMESTAMP NOT NULL,
    views INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    total_clicks INTEGER DEFAULT 0,
    avg_duration_seconds INTEGER,
    top_referrer TEXT,
    top_country TEXT,
    PRIMARY KEY (profile_id, hour)
);

-- Footer Items
CREATE TABLE IF NOT EXISTS footer_items (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    profile_id TEXT REFERENCES profiles(id) ON DELETE CASCADE,
    item_type TEXT CHECK (item_type IN ('text', 'link', 'social_row', 'badge', 'html')),
    content TEXT NOT NULL,
    position INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(profile_id, position)
);

-- Profile Customization
CREATE TABLE IF NOT EXISTS profile_themes (
    profile_id TEXT PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
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
CREATE TABLE IF NOT EXISTS qr_code_settings (
    profile_id TEXT PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    size INTEGER DEFAULT 256 CHECK (size IN (128, 256, 512, 1024)),
    error_correction TEXT DEFAULT 'M' CHECK (error_correction IN ('L', 'M', 'Q', 'H')),
    style TEXT DEFAULT 'square' CHECK (style IN ('square', 'rounded', 'dots')),
    dark_color TEXT DEFAULT '#000000',
    light_color TEXT DEFAULT '#ffffff',
    logo_enabled BOOLEAN DEFAULT 0,
    logo_size INTEGER DEFAULT 30,
    format TEXT DEFAULT 'png' CHECK (format IN ('png', 'svg', 'pdf')),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Import/Export Jobs
CREATE TABLE IF NOT EXISTS import_jobs (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT REFERENCES users(id),
    source TEXT CHECK (source IN ('linktree', 'linkstack', 'carrd', 'aboutme', 'csv', 'json')),
    status TEXT CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    file_path TEXT,
    result TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- API Management
CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    scopes TEXT,
    rate_limit INTEGER DEFAULT 1000,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS api_webhooks (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    secret TEXT NOT NULL,
    events TEXT NOT NULL,
    active BOOLEAN DEFAULT 1,
    failure_count INTEGER DEFAULT 0,
    last_triggered_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Social Features
CREATE TABLE IF NOT EXISTS profile_tags (
    profile_id TEXT REFERENCES profiles(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (profile_id, tag)
);

CREATE TABLE IF NOT EXISTS featured_profiles (
    profile_id TEXT PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    featured_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    featured_until TIMESTAMP,
    reason TEXT
);

CREATE TABLE IF NOT EXISTS profile_verification (
    profile_id TEXT PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    verified BOOLEAN DEFAULT 0,
    verification_type TEXT CHECK (verification_type IN ('email', 'domain', 'social', 'manual')),
    verification_data TEXT,
    verified_at TIMESTAMP,
    verified_by TEXT REFERENCES users(id)
);

-- Organizations/Teams
CREATE TABLE IF NOT EXISTS organizations (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    owner_id TEXT REFERENCES users(id),
    logo_url TEXT,
    settings TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS organization_members (
    org_id TEXT REFERENCES organizations(id) ON DELETE CASCADE,
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    role TEXT CHECK (role IN ('owner', 'admin', 'editor', 'viewer')),
    invited_by TEXT REFERENCES users(id),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (org_id, user_id)
);

CREATE TABLE IF NOT EXISTS organization_invites (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    org_id TEXT REFERENCES organizations(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    role TEXT CHECK (role IN ('admin', 'editor', 'viewer')),
    token TEXT UNIQUE NOT NULL,
    invited_by TEXT REFERENCES users(id),
    expires_at TIMESTAMP,
    accepted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Content Moderation
CREATE TABLE IF NOT EXISTS blocked_patterns (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    pattern TEXT NOT NULL,
    pattern_type TEXT CHECK (pattern_type IN ('domain', 'url', 'word')),
    reason TEXT,
    severity TEXT CHECK (severity IN ('warning', 'block')),
    created_by TEXT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reported_content (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    content_type TEXT CHECK (content_type IN ('profile', 'link')),
    content_id TEXT NOT NULL,
    reporter_ip_hash TEXT,
    reporter_email TEXT,
    reason TEXT CHECK (reason IN ('spam', 'inappropriate', 'phishing', 'copyright', 'other')),
    details TEXT,
    status TEXT CHECK (status IN ('pending', 'reviewing', 'resolved', 'dismissed')),
    moderator_id TEXT REFERENCES users(id),
    moderator_notes TEXT,
    action_taken TEXT CHECK (action_taken IN ('none', 'warning', 'edited', 'suspended', 'deleted')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP
);

-- Shortlinks
CREATE TABLE IF NOT EXISTS shortlinks (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    short_code TEXT UNIQUE NOT NULL,
    target_url TEXT NOT NULL,
    profile_id TEXT REFERENCES profiles(id) ON DELETE CASCADE,
    title TEXT,
    click_count INTEGER DEFAULT 0,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Compliance & Privacy
CREATE TABLE IF NOT EXISTS user_consent (
    user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    terms_version TEXT NOT NULL,
    terms_accepted_at TIMESTAMP NOT NULL,
    privacy_version TEXT NOT NULL,
    privacy_accepted_at TIMESTAMP NOT NULL,
    cookies_accepted BOOLEAN DEFAULT 0,
    cookies_accepted_at TIMESTAMP,
    marketing_consent BOOLEAN DEFAULT 0,
    marketing_consent_at TIMESTAMP,
    data_export_requested_at TIMESTAMP,
    deletion_requested_at TIMESTAMP,
    deletion_scheduled_for TIMESTAMP
);

CREATE TABLE IF NOT EXISTS data_exports (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    status TEXT CHECK (status IN ('pending', 'processing', 'completed', 'expired')),
    file_path TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_log (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    user_id TEXT REFERENCES users(id),
    action TEXT NOT NULL,
    resource_type TEXT,
    resource_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Profile Maintenance
CREATE TABLE IF NOT EXISTS profile_maintenance (
    profile_id TEXT PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    status TEXT CHECK (status IN ('active', 'maintenance', 'suspended')),
    message TEXT,
    bypass_token TEXT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    estimated_end TIMESTAMP
);

CREATE TABLE IF NOT EXISTS email_verification_tokens (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Server Admins (separate from users per spec — different DB table)
CREATE TABLE IF NOT EXISTS server_admins (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT 0,
    two_factor_enabled BOOLEAN DEFAULT 0,
    two_factor_secret TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP
);

-- Sessions (server-side session storage, HttpOnly cookie auth)
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    user_type TEXT NOT NULL,
    username TEXT NOT NULL,
    role TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Profile Views (raw event log for analytics)
CREATE TABLE IF NOT EXISTS profile_views (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_id TEXT NOT NULL,
    viewer_ip TEXT,
    referrer TEXT,
    user_agent TEXT,
    country TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Link Clicks (raw event log for link analytics)
CREATE TABLE IF NOT EXISTS link_clicks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    link_id TEXT NOT NULL,
    clicker_ip TEXT,
    referrer TEXT,
    user_agent TEXT,
    country TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Cluster Nodes (optional HA clustering support)
CREATE TABLE IF NOT EXISTS cluster_nodes (
    id TEXT PRIMARY KEY,
    hostname TEXT NOT NULL,
    address TEXT NOT NULL,
    port INTEGER NOT NULL,
    status TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT 0,
    last_heartbeat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_profiles_user_id ON profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_profiles_slug ON profiles(slug);
CREATE INDEX IF NOT EXISTS idx_links_profile_id ON links(profile_id);
CREATE INDEX IF NOT EXISTS idx_links_position ON links(profile_id, position);
CREATE INDEX IF NOT EXISTS idx_analytics_profile_id ON analytics(profile_id);
CREATE INDEX IF NOT EXISTS idx_analytics_created_at ON analytics(created_at);
CREATE INDEX IF NOT EXISTS idx_analytics_sessions_profile ON analytics_sessions(profile_id);
CREATE INDEX IF NOT EXISTS idx_services_category ON services(category);
CREATE INDEX IF NOT EXISTS idx_services_popularity ON services(popularity DESC);
