# SMTP and Notification System

Complete SMTP and notification implementation for Cassocial v1.0.0.

## Overview

This package provides a comprehensive email and notification system with support for multiple SMTP providers, retry logic, notification batching, and HTML email templates.

## Components

### 1. **smtp.go** - SMTP Client
Core SMTP functionality with support for multiple providers and security methods.

**Features:**
- Multiple provider support (Gmail, Yahoo, Outlook, Custom)
- Multiple security methods (NONE, STARTTLS, SSL/TLS)
- Predefined port configurations matching SPEC
- Connection testing
- Retry logic with exponential backoff
- TLS/SSL support

**Supported Providers:**
- `CUSTOM` - Custom SMTP server (manual configuration)
- `Gmail` - Google Gmail (smtp.gmail.com)
- `Yahoo` - Yahoo Mail (smtp.mail.yahoo.com)
- `Outlook` - Microsoft Outlook (smtp-mail.outlook.com)

**Port/Security Combinations:**
- `25 (NONE)` - Plain SMTP (not recommended)
- `587 (STARTTLS)` - STARTTLS (recommended)
- `465 (SSL/TLS)` - SSL/TLS
- `2525 (STARTTLS)` - Alternative STARTTLS

### 2. **mailer.go** - Email Sending Functions
High-level email sending interface with pre-built templates.

**Features:**
- Welcome emails
- Password reset emails
- Email verification
- 2FA code emails
- Team invitations
- Generic notifications
- Plain text and HTML support
- Connection testing

**Admin Notification Helpers:**
- User registration notifications
- Domain verification status
- Backup status alerts
- Certificate renewal reminders
- Emergency alerts
- High traffic warnings
- Bug report notifications
- Profile suspension notices
- Data export ready notifications

### 3. **templates.go** - HTML Email Templates
Beautiful, responsive HTML email templates.

**Features:**
- Responsive design (mobile-friendly)
- Dracula-inspired color scheme matching app theme
- Gradient header design
- Clean, professional layout
- Info/Warning/Danger boxes
- Button CTAs
- Plain text fallbacks
- Footer with privacy/terms links

**Available Templates:**
- Welcome email with verification link
- Password reset with security warnings
- Email verification
- 2FA code with expiry notice
- Team invitations
- Generic notification with severity levels
- Emergency alerts

### 4. **notifications.go** - Notification Management
Advanced notification system with queuing, batching, and priority handling.

**Features:**
- Notification queuing and batching
- Priority-based retry logic
- Configurable batch delays (default: 5 minutes)
- Emergency notifications bypass batching
- Notification type filtering
- Background processing
- Thread-safe queue management

**Notification Types:**
- `emergency` - Critical system failures (sent immediately, 5 retries)
- `certificate` - SSL certificate expiry warnings (3 retries)
- `bug_report` - User-submitted bug reports (1 retry)
- `user_registration` - New user signups (1 retry)
- `domain_verification` - Custom domain verification (1 retry)
- `backup_status` - Backup success/failure (1-3 retries)
- `high_traffic` - Traffic threshold exceeded (1 retry)

**Priority Levels:**
- `PriorityNormal` (1) - 1 retry
- `PriorityHigh` (3) - 3 retries
- `PriorityEmergency` (5) - 5 retries, sent immediately

## Configuration

### SMTP Configuration (from settings table)

```go
type SMTPConfig struct {
    Provider    string // "CUSTOM", "Gmail", "Yahoo", "Outlook"
    Host        string // SMTP hostname
    Port        int    // SMTP port (25, 587, 465, 2525)
    Security    string // "NONE", "STARTTLS", "SSL/TLS"
    User        string // SMTP username (optional)
    Password    string // SMTP password (required if user set)
    FromName    string // Sender name
    FromAddress string // Sender email (required)
    AdminEmail  string // Admin email for notifications
    Enabled     bool   // Enable/disable SMTP
    RetryCount  int    // Default retry count
    RetryDelay  int    // Delay between retries (seconds)
}
```

### Notification Preferences (from settings table)

```go
type NotificationPreferences struct {
    Emergency          bool // Emergency alerts enabled
    Certificate        bool // Certificate renewal alerts
    BugReport          bool // Bug report notifications
    UserRegistration   bool // New user registration alerts
    DomainVerification bool // Domain verification alerts
    BackupStatus       bool // Backup status notifications
    HighTraffic        bool // High traffic warnings
    BatchDelay         int  // Batch delay in seconds (default: 300)
}
```

## Usage Examples

See `example_usage.go` for comprehensive examples covering:
1. Setting up SMTP client
2. Sending emails with Mailer
3. Using Notification Manager
4. Provider-specific configurations
5. Port/Security configurations
6. Admin notification helpers
7. Getting provider hosts
8. Retry logic
9. Checking if mailer is enabled
10. Testing SMTP configuration

## Validation Rules

### SMTP Configuration
- `Host` is required
- `Port` must be between 1 and 65535
- `FromAddress` is required and must be valid email
- `Password` is required if `User` is set
- Security type must be one of: NONE, STARTTLS, SSL/TLS

### Email Addresses
- Must conform to RFC 5322 standard
- Validated before sending

## Security Features

- **TLS Support**: Full support for STARTTLS and SSL/TLS
- **Encrypted Storage**: Credentials encrypted using master key
- **No Plaintext Logging**: Passwords never logged
- **Connection Timeouts**: 10-second timeout for connections
- **Auth Failure Detection**: No retry on authentication failures
- **Minimum TLS Version**: TLS 1.2+

## Error Handling

### Error Types
- `ErrInvalidConfig` - Invalid SMTP configuration
- `ErrConnectionFailed` - Failed to connect to SMTP server
- `ErrAuthFailed` - Authentication failed
- `ErrSendFailed` - Failed to send email
- `ErrInvalidProvider` - Invalid provider specified
- `ErrMailerNotConfigured` - Mailer not properly configured
- `ErrRecipientRequired` - Recipient email required

### Retry Behavior
- Exponential backoff for retries
- Configurable retry count based on priority
- No retry on authentication failures
- Logs all failures with details

## Operational Behavior

### When SMTP is Disabled
- All send functions return gracefully without error
- Operations are logged for debugging
- System continues to function normally

### Notification Batching
- Normal/High priority notifications are batched
- Emergency notifications sent immediately
- Batch window configurable (default: 5 minutes)
- Queue processed when batch window expires

### Connection Testing
- Test connection before saving configuration
- Validates host, port, and credentials
- Tests authentication if credentials provided
- Returns detailed error messages

## Integration Points

### Database Settings
All SMTP and notification settings stored in `settings` table:
- `smtp_provider`, `smtp_host`, `smtp_port`, `smtp_security`
- `smtp_user`, `smtp_password` (encrypted)
- `smtp_from_name`, `smtp_from_address`
- `admin_email`, `smtp_enabled`
- `smtp_retry_count`, `smtp_retry_delay`
- `notify_emergency`, `notify_certificate`, etc.
- `notification_batch_delay`

### Admin Panel Integration
- SMTP configuration page with provider dropdown
- Port/Security dropdown with predefined combinations
- Test connection button
- Notification preferences toggles
- Real-time validation

### Encryption
- SMTP passwords encrypted using `CASSOCIAL_MASTER_KEY`
- Decrypted only when needed for sending
- Never stored in plaintext

## Testing

### Unit Tests
Test coverage should include:
- SMTP connection testing
- Email template rendering
- Notification queuing and batching
- Retry logic
- Error handling
- Provider host resolution

### Integration Tests
- End-to-end email sending
- Notification manager lifecycle
- Priority-based sending
- Batch processing

### Manual Testing
1. Configure SMTP settings in admin panel
2. Test connection
3. Send test email
4. Verify email received
5. Test notification batching
6. Test emergency notification (immediate)
7. Verify retry logic on failures

## Performance Considerations

### Email Sending
- Asynchronous sending recommended for user-facing operations
- Connection pooling not implemented (SMTP is lightweight)
- Timeout prevents hanging connections

### Notification Manager
- Background goroutine for processing
- Thread-safe queue operations
- Minimal memory footprint
- Automatic cleanup on shutdown

### Resource Usage
- ~100KB memory per notification manager
- ~10KB per queued notification
- Negligible CPU usage

## Future Enhancements

Potential improvements not in v1.0.0:
- Connection pooling for high-volume sending
- Template customization via admin panel
- SMS notification support
- Webhook notification support
- Email bounce handling
- Delivery status tracking
- Rate limiting per recipient

## Compliance

### GDPR
- Email addresses handled according to privacy policy
- Consent tracked in `user_consent` table
- Data export includes email notification history
- Deletion includes all email records

### CAN-SPAM
- Unsubscribe links in marketing emails
- Clear sender identification
- Honest subject lines
- Physical address in footer (configurable)

## Troubleshooting

### Common Issues

**SMTP connection fails:**
- Verify host and port are correct
- Check firewall allows outbound SMTP
- Ensure credentials are valid
- Try different port/security combination

**Gmail authentication fails:**
- Use App Password, not regular password
- Enable "Less secure app access" if needed
- Check 2FA settings

**Emails not sending:**
- Check SMTP enabled in settings
- Verify mailer.IsEnabled() returns true
- Check logs for error messages
- Test connection in admin panel

**Notification delays:**
- Check batch delay setting
- Emergency notifications bypass batching
- Verify notification manager is running

## File Structure

```
internal/smtp/
├── smtp.go              # Core SMTP client
├── mailer.go            # Email sending functions
├── templates.go         # HTML email templates
├── notifications.go     # Notification management
├── example_usage.go     # Usage examples
└── README.md           # This file
```

## Dependencies

- Standard library only:
  - `net/smtp` - SMTP protocol
  - `crypto/tls` - TLS/SSL support
  - `html/template` - Template rendering
  - `sync` - Thread-safe operations
  - `time` - Scheduling and delays

- Internal packages:
  - `github.com/casapps/cassocial/internal/models` - Data models

## License

MIT License - See LICENSE.md

## Version

1.0.0 - Production Ready
