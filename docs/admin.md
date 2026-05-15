# Admin Panel Guide

The admin panel is available at `/admin` after completing the setup wizard.

## Admin API Endpoints

All admin API endpoints require an authenticated session with the `admin` role. They are available under `/api/admin/`:

```
GET    /api/admin/users                       # List all users
GET    /api/admin/users/{id}                  # Get user details
PUT    /api/admin/users/{id}                  # Update user (role/status)
DELETE /api/admin/users/{id}                  # Delete user
GET    /api/admin/stats                       # System statistics
POST   /api/admin/backup                      # Trigger manual backup
GET    /api/admin/settings                    # Get all settings
PUT    /api/admin/settings                    # Update settings
POST   /api/admin/services/import             # Import services database
POST   /api/admin/cache/clear                 # Clear system cache
GET    /api/admin/smtp/config                 # Get SMTP configuration
PUT    /api/admin/smtp/config                 # Update SMTP configuration
POST   /api/admin/smtp/test                   # Test SMTP connection
GET    /api/admin/notifications/preferences   # Get notification preferences
PUT    /api/admin/notifications/preferences   # Update notification preferences
```

## Dashboard

The dashboard shows:
- Total users, profiles, links
- Total views and clicks
- Recent activity
- System information

## User Management

**Location**: `/admin/users`

- View all users
- Edit user details (role, status)
- Suspend/activate users
- Delete users
- Assign roles (admin, user)

## Settings

**Location**: `/admin/settings`

All `server.yml` settings are editable through the web UI:

### Server Settings
- Listen address and port
- Application mode (production/development)
- Debug mode

### Database Settings
- Database driver (SQLite, PostgreSQL, MySQL)
- Connection details
- Connection pool settings

### Email Settings
- SMTP host, port, credentials
- From address
- TLS settings
- Test email functionality

### Features & Limits
- Allow registration (enable/disable)
- Max profiles per user (default: 5)
- Max links per profile (default: 100)

## Profile Management

**Location**: `/admin/profiles`

- View all profiles (public and private)
- Search profiles by slug or user
- View profile analytics
- Moderate profiles

## Analytics

**Location**: `/admin/analytics`

System-wide analytics:
- Total views and clicks
- Top profiles by views
- Geographic data
- Device/browser statistics

Export analytics in CSV, JSON, or PDF format.

## Services Database

**Location**: `/admin/services`

Manage the database of 5000+ supported services:
- View all services
- Add custom services
- Edit service icons
- Enable/disable services

## Themes

**Location**: `/admin/themes`

- Manage global themes
- Create custom themes
- Set default theme

## SMTP Configuration

**Location**: `/admin/smtp`

- Configure SMTP settings
- Test email sending
- View email queue status

## Backup & Restore

**Location**: `/admin/backup`

### Automated Backups

- Scheduled daily at 2:30am
- Keeps last 4 backups automatically
- Includes databases, config, and uploads

### Manual Backup

1. Click "Create Backup"
2. Wait for backup to complete
3. Download backup file if needed

### Restore

1. Select backup from list
2. Click "Restore"
3. Confirm restoration
4. Server will restart with restored data

**⚠️ Warning**: Restore will overwrite current data. Download a current backup first!

## Maintenance Mode

**Location**: `/admin/maintenance`

Enable maintenance mode to:
- Display maintenance message to visitors
- Keep admin panel accessible
- Bypass maintenance for specific IPs (localhost always bypassed)
- Perform updates or maintenance safely

### Custom Message

Set a custom message shown to visitors during maintenance:

```
We're making improvements! Back online soon.
```

## Security Settings

**Location**: `/admin/security`

- SSL/TLS configuration
- Rate limiting rules
- Blocked patterns and domains
- Security headers
- IP whitelisting/blacklisting

## Moderation Queue

**Location**: `/admin/moderation`

- Review reported content
- Approve or remove profiles/links
- Ban users
- Manage blocked patterns
- View moderation history

## API Keys

**Location**: `/admin/api-keys`

- Generate API keys for external access
- View existing keys
- Revoke keys
- Set key permissions and expiration
