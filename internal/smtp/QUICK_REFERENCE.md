# SMTP System Quick Reference

## Quick Setup

```go
// 1. Create SMTP configuration
config := &models.SMTPConfig{
    Provider:    "Gmail",
    Host:        "smtp.gmail.com",
    Port:        587,
    Security:    "STARTTLS",
    User:        "your-email@gmail.com",
    Password:    "app-password",
    FromName:    "Cassocial",
    FromAddress: "your-email@gmail.com",
    AdminEmail:  "admin@example.com",
    Enabled:     true,
    RetryCount:  3,
    RetryDelay:  60,
}

// 2. Create mailer
mailer, _ := smtp.NewMailer(config, "Cassocial", "https://casjay.link")

// 3. Create notification manager
notifMgr := smtp.NewNotificationManager(mailer, prefs, "admin@example.com")
notifMgr.Start()
defer notifMgr.Stop()
```

## Common Operations

### Send Welcome Email
```go
mailer.SendWelcome("user@example.com", "username", "https://verify-url")
```

### Send Password Reset
```go
mailer.SendPasswordReset("user@example.com", "username", "https://reset-url")
```

### Send 2FA Code
```go
mailer.SendTwoFactorCode("user@example.com", "username", "123456")
```

### Send Notification
```go
mailer.SendNotification(
    "admin@example.com",
    "Admin",
    "Backup Complete",
    "<p>Backup finished successfully.</p>",
    "info",
    1, // retries
)
```

### Queue Notification (Batched)
```go
notifMgr.NotifyUserRegistration("newuser", "newuser@example.com")
notifMgr.NotifyBackupStatus("success", "Completed at 03:00")
notifMgr.NotifyHighTraffic(1500, 1000)
```

### Emergency Alert (Immediate)
```go
notifMgr.NotifyEmergency("Critical Error", "Database connection lost!")
```

## Providers & Ports

### Provider Hosts
- **Gmail**: smtp.gmail.com
- **Yahoo**: smtp.mail.yahoo.com
- **Outlook**: smtp-mail.outlook.com
- **Custom**: Manual entry

### Port/Security Combinations
- **25 (NONE)**: Plain SMTP (not recommended)
- **587 (STARTTLS)**: STARTTLS (recommended)
- **465 (SSL/TLS)**: SSL/TLS
- **2525 (STARTTLS)**: Alternative STARTTLS

## Notification Types

| Type | Priority | Retries | Immediate |
|------|----------|---------|-----------|
| Emergency | 5 | 5 | Yes |
| Certificate | 1-3 | 1-3 | No |
| Bug Report | 1 | 1 | No |
| User Registration | 1 | 1 | No |
| Domain Verification | 1 | 1 | No |
| Backup Status | 1-3 | 1-3 | No |
| High Traffic | 1 | 1 | No |

## Error Handling

```go
if err := mailer.SendWelcome(...); err != nil {
    if errors.Is(err, smtp.ErrAuthFailed) {
        // Authentication failed - check credentials
    } else if errors.Is(err, smtp.ErrConnectionFailed) {
        // Connection failed - check host/port
    } else {
        // Other error
    }
}
```

## Testing Connection

```go
if err := mailer.TestConnection(); err != nil {
    log.Printf("SMTP test failed: %v", err)
} else {
    log.Println("SMTP connection successful!")
}
```

## Configuration Validation

```go
if err := config.Validate(); err != nil {
    log.Printf("Invalid config: %v", err)
}
```

## Check If Enabled

```go
if mailer.IsEnabled() {
    // Send email
} else {
    // Log instead
}
```

## File Locations

- **smtp.go**: Core SMTP client
- **mailer.go**: Email sending functions
- **templates.go**: HTML email templates
- **notifications.go**: Notification manager
- **example_usage.go**: Usage examples
- **README.md**: Full documentation
- **QUICK_REFERENCE.md**: This file
