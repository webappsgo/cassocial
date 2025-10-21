package smtp

// This file demonstrates example usage of the SMTP and notification system.
// It is not meant to be compiled, just for documentation purposes.

/*
Example 1: Setting up the SMTP client

	import (
		"github.com/casapps/cassocial/internal/models"
		"github.com/casapps/cassocial/internal/smtp"
	)

	// Create SMTP configuration
	config := &models.SMTPConfig{
		Provider:    "Gmail",
		Host:        "smtp.gmail.com",
		Port:        587,
		Security:    "STARTTLS",
		User:        "your-email@gmail.com",
		Password:    "your-app-password",
		FromName:    "Cassocial",
		FromAddress: "your-email@gmail.com",
		AdminEmail:  "admin@example.com",
		Enabled:     true,
		RetryCount:  3,
		RetryDelay:  60,
	}

	// Test the connection
	client, err := smtp.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create SMTP client: %v", err)
	}

	if err := client.TestConnection(); err != nil {
		log.Fatalf("SMTP connection test failed: %v", err)
	}

	log.Println("SMTP connection successful!")


Example 2: Sending emails using the Mailer

	// Create mailer
	mailer, err := smtp.NewMailer(config, "Cassocial", "https://casjay.link")
	if err != nil {
		log.Fatalf("Failed to create mailer: %v", err)
	}

	// Send welcome email
	err = mailer.SendWelcome(
		"user@example.com",
		"johndoe",
		"https://casjay.link/verify?token=abc123",
	)

	// Send password reset
	err = mailer.SendPasswordReset(
		"user@example.com",
		"johndoe",
		"https://casjay.link/reset?token=xyz789",
	)

	// Send 2FA code
	err = mailer.SendTwoFactorCode(
		"user@example.com",
		"johndoe",
		"123456",
	)


Example 3: Using the Notification Manager

	// Create notification preferences
	prefs := &models.NotificationPreferences{
		Emergency:          true,
		Certificate:        true,
		BugReport:          true,
		UserRegistration:   true,
		DomainVerification: true,
		BackupStatus:       true,
		HighTraffic:        true,
		BatchDelay:         300, // 5 minutes
	}

	// Create notification manager
	notifManager := smtp.NewNotificationManager(
		mailer,
		prefs,
		"admin@example.com",
	)

	// Start the notification processing loop
	notifManager.Start()
	defer notifManager.Stop()

	// Queue notifications (will be batched)
	notifManager.NotifyUserRegistration("newuser", "newuser@example.com")
	notifManager.NotifyBackupStatus("success", "Backup completed at 2025-09-30 03:00")
	notifManager.NotifyHighTraffic(1500, 1000)

	// Emergency notifications are sent immediately
	notifManager.NotifyEmergency(
		"Critical System Error",
		"Database connection lost. Immediate action required.",
	)


Example 4: Provider-specific configuration

	// Gmail setup
	gmailConfig := &models.SMTPConfig{
		Provider:    "Gmail",
		Host:        "smtp.gmail.com",
		Port:        587,
		Security:    "STARTTLS",
		User:        "your-email@gmail.com",
		Password:    "app-specific-password", // Not regular password!
		FromName:    "Cassocial",
		FromAddress: "your-email@gmail.com",
		Enabled:     true,
	}

	// Yahoo setup
	yahooConfig := &models.SMTPConfig{
		Provider:    "Yahoo",
		Host:        "smtp.mail.yahoo.com",
		Port:        587,
		Security:    "STARTTLS",
		User:        "your-email@yahoo.com",
		Password:    "app-password",
		FromName:    "Cassocial",
		FromAddress: "your-email@yahoo.com",
		Enabled:     true,
	}

	// Outlook setup
	outlookConfig := &models.SMTPConfig{
		Provider:    "Outlook",
		Host:        "smtp-mail.outlook.com",
		Port:        587,
		Security:    "STARTTLS",
		User:        "your-email@outlook.com",
		Password:    "your-password",
		FromName:    "Cassocial",
		FromAddress: "your-email@outlook.com",
		Enabled:     true,
	}

	// Custom SMTP server (e.g., private mail server)
	customConfig := &models.SMTPConfig{
		Provider:    "CUSTOM",
		Host:        "mail.example.com",
		Port:        465, // SSL/TLS
		Security:    "SSL/TLS",
		User:        "noreply@example.com",
		Password:    "secure-password",
		FromName:    "Cassocial Notifications",
		FromAddress: "noreply@example.com",
		Enabled:     true,
	}


Example 5: Port/Security configurations (as per SPEC)

	// Port 25 with no encryption (not recommended)
	config25 := &models.SMTPConfig{
		Host:     "mail.example.com",
		Port:     25,
		Security: "NONE",
	}

	// Port 587 with STARTTLS (recommended)
	config587 := &models.SMTPConfig{
		Host:     "mail.example.com",
		Port:     587,
		Security: "STARTTLS",
	}

	// Port 465 with SSL/TLS
	config465 := &models.SMTPConfig{
		Host:     "mail.example.com",
		Port:     465,
		Security: "SSL/TLS",
	}

	// Port 2525 with STARTTLS (alternative)
	config2525 := &models.SMTPConfig{
		Host:     "mail.example.com",
		Port:     2525,
		Security: "STARTTLS",
	}


Example 6: Admin notification helpers

	// Certificate expiry warning
	notifManager.NotifyCertificateExpiring("casjay.link", 7)

	// Bug report
	notifManager.NotifyBugReport(
		"user@example.com",
		"Links not rendering correctly on mobile devices",
	)

	// Domain verification
	notifManager.NotifyDomainVerification("custom.example.com", "verified")

	// High traffic alert
	notifManager.NotifyHighTraffic(2000, 1000) // current: 2000, threshold: 1000


Example 7: Getting provider host automatically

	// Get host for known providers
	gmailHost, err := smtp.GetProviderHost(smtp.ProviderGmail)
	// Returns: "smtp.gmail.com"

	yahooHost, err := smtp.GetProviderHost(smtp.ProviderYahoo)
	// Returns: "smtp.mail.yahoo.com"

	outlookHost, err := smtp.GetProviderHost(smtp.ProviderOutlook)
	// Returns: "smtp-mail.outlook.com"

	customHost, err := smtp.GetProviderHost(smtp.ProviderCustom)
	// Returns: "" (requires manual entry)


Example 8: Retry logic based on priority

	// Normal priority: 1 retry
	err := mailer.SendNotification(
		"admin@example.com",
		"Admin",
		"Info Message",
		"<p>This is informational.</p>",
		"info",
		1, // 1 retry
	)

	// High priority: 3 retries (password resets, certificates)
	err = mailer.SendPasswordReset(
		"user@example.com",
		"username",
		"reset-url",
	) // Internally uses 3 retries

	// Emergency priority: 5 retries
	err = mailer.SendEmergencyAlert(
		"admin@example.com",
		"System Critical",
		"Database connection lost!",
	) // Uses 5 retries


Example 9: Checking if mailer is enabled

	if mailer.IsEnabled() {
		// SMTP is configured and enabled
		err := mailer.SendWelcome(...)
	} else {
		// SMTP is disabled, log instead
		log.Println("SMTP disabled: would have sent welcome email")
	}


Example 10: Testing SMTP configuration from admin panel

	// Admin tests SMTP settings before saving
	testConfig := &models.SMTPConfig{
		Host:        "smtp.gmail.com",
		Port:        587,
		Security:    "STARTTLS",
		User:        "admin@example.com",
		Password:    "app-password",
		FromAddress: "admin@example.com",
		Enabled:     true,
	}

	// Validate configuration
	if err := testConfig.Validate(); err != nil {
		return fmt.Errorf("Invalid SMTP config: %v", err)
	}

	// Test connection
	testMailer, err := smtp.NewMailer(testConfig, "Cassocial", "https://casjay.link")
	if err != nil {
		return fmt.Errorf("Failed to create test mailer: %v", err)
	}

	if err := testMailer.TestConnection(); err != nil {
		return fmt.Errorf("SMTP connection test failed: %v", err)
	}

	// Connection successful, save configuration
	log.Println("SMTP test successful!")

*/
