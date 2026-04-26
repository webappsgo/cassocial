package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/casapps/cassocial/src/server/model"
)

var (
	ErrMailerNotConfigured = errors.New("mailer is not configured")
	ErrRecipientRequired   = errors.New("recipient email is required")
)

// Mailer handles email sending operations
type Mailer struct {
	client   *Client
	siteName string
	siteURL  string
	enabled  bool
}

// NewMailer creates a new mailer instance
func NewMailer(config *models.SMTPConfig, siteName, siteURL string) (*Mailer, error) {
	if config == nil {
		return &Mailer{
			siteName: siteName,
			siteURL:  siteURL,
			enabled:  false,
		}, nil
	}

	// Create SMTP client
	client, err := NewClient(config)
	if err != nil {
		log.Printf("Failed to create SMTP client: %v", err)
		return &Mailer{
			siteName: siteName,
			siteURL:  siteURL,
			enabled:  false,
		}, nil
	}

	return &Mailer{
		client:   client,
		siteName: siteName,
		siteURL:  siteURL,
		enabled:  config.Enabled,
	}, nil
}

// IsEnabled returns whether the mailer is enabled
func (m *Mailer) IsEnabled() bool {
	return m.enabled && m.client != nil
}

// SendWelcome sends a welcome email to a new user
func (m *Mailer) SendWelcome(to, username, verificationURL string) error {
	if !m.IsEnabled() {
		log.Printf("Mailer disabled: would have sent welcome email to %s", to)
		return nil
	}

	if to == "" {
		return ErrRecipientRequired
	}

	// Prepare template data
	data := NewTemplateData(m.siteName, m.siteURL, username)
	tmpl := WelcomeEmail(data, verificationURL)

	// Send email
	return m.client.Send([]string{to}, tmpl.Subject, tmpl.HTMLBody, true)
}

// SendPasswordReset sends a password reset email
func (m *Mailer) SendPasswordReset(to, username, resetURL string) error {
	if !m.IsEnabled() {
		log.Printf("Mailer disabled: would have sent password reset email to %s", to)
		return nil
	}

	if to == "" {
		return ErrRecipientRequired
	}

	// Prepare template data
	data := NewTemplateData(m.siteName, m.siteURL, username)
	tmpl := PasswordResetEmail(data, resetURL)

	// Send email with retry (high priority)
	return m.client.SendWithRetry([]string{to}, tmpl.Subject, tmpl.HTMLBody, true, 3)
}

// SendEmailVerification sends an email verification link
func (m *Mailer) SendEmailVerification(to, username, verificationURL string) error {
	if !m.IsEnabled() {
		log.Printf("Mailer disabled: would have sent verification email to %s", to)
		return nil
	}

	if to == "" {
		return ErrRecipientRequired
	}

	// Prepare template data
	data := NewTemplateData(m.siteName, m.siteURL, username)
	tmpl := EmailVerificationEmail(data, verificationURL)

	// Send email
	return m.client.Send([]string{to}, tmpl.Subject, tmpl.HTMLBody, true)
}

// SendTwoFactorCode sends a 2FA code email
func (m *Mailer) SendTwoFactorCode(to, username, code string) error {
	if !m.IsEnabled() {
		log.Printf("Mailer disabled: would have sent 2FA code to %s", to)
		return nil
	}

	if to == "" {
		return ErrRecipientRequired
	}

	// Prepare template data
	data := NewTemplateData(m.siteName, m.siteURL, username)
	tmpl := TwoFactorCodeEmail(data, code)

	// Send email with retry (high priority)
	return m.client.SendWithRetry([]string{to}, tmpl.Subject, tmpl.HTMLBody, true, 3)
}

// SendTeamInvite sends a team invitation email
func (m *Mailer) SendTeamInvite(to, recipientName, orgName, role, inviteURL string) error {
	if !m.IsEnabled() {
		log.Printf("Mailer disabled: would have sent team invite to %s", to)
		return nil
	}

	if to == "" {
		return ErrRecipientRequired
	}

	// Prepare template data
	data := NewTemplateData(m.siteName, m.siteURL, recipientName)
	tmpl := TeamInviteEmail(data, orgName, role, inviteURL)

	// Send email
	return m.client.Send([]string{to}, tmpl.Subject, tmpl.HTMLBody, true)
}

// SendNotification sends a generic notification email
func (m *Mailer) SendNotification(to, recipientName, title, message, severity string, retries int) error {
	if !m.IsEnabled() {
		log.Printf("Mailer disabled: would have sent notification to %s: %s", to, title)
		return nil
	}

	if to == "" {
		return ErrRecipientRequired
	}

	// Prepare template data
	data := NewTemplateData(m.siteName, m.siteURL, recipientName)
	tmpl := NotificationEmail(data, title, message, severity)

	// Send email with retry based on priority
	return m.client.SendWithRetry([]string{to}, tmpl.Subject, tmpl.HTMLBody, true, retries)
}

// SendPlainText sends a plain text email
func (m *Mailer) SendPlainText(to []string, subject, body string) error {
	if !m.IsEnabled() {
		log.Printf("Mailer disabled: would have sent email to %v: %s", to, subject)
		return nil
	}

	if len(to) == 0 {
		return ErrRecipientRequired
	}

	return m.client.Send(to, subject, body, false)
}

// SendHTML sends an HTML email
func (m *Mailer) SendHTML(to []string, subject, body string) error {
	if !m.IsEnabled() {
		log.Printf("Mailer disabled: would have sent HTML email to %v: %s", to, subject)
		return nil
	}

	if len(to) == 0 {
		return ErrRecipientRequired
	}

	return m.client.Send(to, subject, body, true)
}

// TestConnection tests the SMTP connection
func (m *Mailer) TestConnection() error {
	if m.client == nil {
		return ErrMailerNotConfigured
	}

	return m.client.TestConnection()
}

// Email templates for common scenarios

// SendUserRegistrationNotification notifies admin of new user registration
func (m *Mailer) SendUserRegistrationNotification(adminEmail, username, userEmail string) error {
	if !m.IsEnabled() {
		return nil
	}

	message := fmt.Sprintf(`
		<p>A new user has registered on %s:</p>
		<ul>
			<li><strong>Username:</strong> %s</li>
			<li><strong>Email:</strong> %s</li>
		</ul>
		<p>Please review and approve their account if required.</p>
	`, m.siteName, username, userEmail)

	return m.SendNotification(
		adminEmail,
		"Administrator",
		"New User Registration",
		message,
		"info",
		1, // Normal priority
	)
}

// SendDomainVerificationNotification notifies admin of domain verification status
func (m *Mailer) SendDomainVerificationNotification(adminEmail, domain, status string) error {
	if !m.IsEnabled() {
		return nil
	}

	var message string
	var severity string

	if status == "verified" {
		message = fmt.Sprintf("<p>The custom domain <strong>%s</strong> has been successfully verified.</p>", domain)
		severity = "info"
	} else {
		message = fmt.Sprintf("<p>Domain verification failed for <strong>%s</strong>. Please check DNS settings.</p>", domain)
		severity = "warning"
	}

	return m.SendNotification(
		adminEmail,
		"Administrator",
		fmt.Sprintf("Domain Verification: %s", domain),
		message,
		severity,
		1,
	)
}

// SendBackupStatusNotification notifies admin of backup status
func (m *Mailer) SendBackupStatusNotification(adminEmail, status, details string) error {
	if !m.IsEnabled() {
		return nil
	}

	var severity string
	var message string

	if status == "success" {
		severity = "info"
		message = fmt.Sprintf("<p>Database backup completed successfully.</p><p>%s</p>", details)
	} else {
		severity = "warning"
		message = fmt.Sprintf("<p>Database backup failed.</p><p><strong>Error:</strong> %s</p>", details)
	}

	return m.SendNotification(
		adminEmail,
		"Administrator",
		fmt.Sprintf("Backup %s", status),
		message,
		severity,
		1,
	)
}

// SendCertificateRenewalNotification notifies admin of SSL certificate renewal
func (m *Mailer) SendCertificateRenewalNotification(adminEmail, domain string, daysUntilExpiry int) error {
	if !m.IsEnabled() {
		return nil
	}

	var severity string
	if daysUntilExpiry <= 7 {
		severity = "warning"
	} else {
		severity = "info"
	}

	message := fmt.Sprintf(`
		<p>The SSL certificate for <strong>%s</strong> will expire in <strong>%d days</strong>.</p>
		<p>Automatic renewal will be attempted. If it fails, manual intervention may be required.</p>
	`, domain, daysUntilExpiry)

	return m.SendNotification(
		adminEmail,
		"Administrator",
		fmt.Sprintf("SSL Certificate Expiring: %s", domain),
		message,
		severity,
		3, // High priority
	)
}

// SendEmergencyAlert sends an emergency alert to admin
func (m *Mailer) SendEmergencyAlert(adminEmail, title, details string) error {
	if !m.IsEnabled() {
		return nil
	}

	message := fmt.Sprintf(`
		<p><strong>EMERGENCY ALERT</strong></p>
		<p>%s</p>
		<p>Immediate action may be required.</p>
	`, details)

	return m.SendNotification(
		adminEmail,
		"Administrator",
		title,
		message,
		"emergency",
		5, // Emergency priority - max retries
	)
}

// SendHighTrafficNotification notifies admin of high traffic
func (m *Mailer) SendHighTrafficNotification(adminEmail string, currentLoad, threshold int) error {
	if !m.IsEnabled() {
		return nil
	}

	message := fmt.Sprintf(`
		<p>Your %s instance is experiencing high traffic.</p>
		<ul>
			<li><strong>Current Load:</strong> %d requests/minute</li>
			<li><strong>Threshold:</strong> %d requests/minute</li>
		</ul>
		<p>System performance may be affected. Consider scaling resources if this persists.</p>
	`, m.siteName, currentLoad, threshold)

	return m.SendNotification(
		adminEmail,
		"Administrator",
		"High Traffic Alert",
		message,
		"warning",
		1,
	)
}

// SendBugReport sends a bug report notification
func (m *Mailer) SendBugReport(adminEmail, reportedBy, description string) error {
	if !m.IsEnabled() {
		return nil
	}

	message := fmt.Sprintf(`
		<p>A new bug report has been submitted:</p>
		<ul>
			<li><strong>Reported By:</strong> %s</li>
		</ul>
		<p><strong>Description:</strong></p>
		<p>%s</p>
	`, reportedBy, description)

	return m.SendNotification(
		adminEmail,
		"Administrator",
		"New Bug Report",
		message,
		"info",
		1,
	)
}

// SendProfileSuspensionNotice sends a notification about profile suspension
func (m *Mailer) SendProfileSuspensionNotice(userEmail, username, reason string) error {
	if !m.IsEnabled() {
		return nil
	}

	message := fmt.Sprintf(`
		<p>Your profile on %s has been suspended.</p>
		<p><strong>Reason:</strong> %s</p>
		<p>If you believe this is an error, please contact support.</p>
	`, m.siteName, reason)

	return m.SendNotification(
		userEmail,
		username,
		"Profile Suspended",
		message,
		"warning",
		3,
	)
}

// SendDataExportReady notifies user that their data export is ready
func (m *Mailer) SendDataExportReady(userEmail, username, downloadURL string) error {
	if !m.IsEnabled() {
		return nil
	}

	message := fmt.Sprintf(`
		<p>Your data export is ready for download.</p>
		<p style="text-align: center;">
			<a href="%s" class="button">Download Data Export</a>
		</p>
		<div class="info-box">
			<p><strong>Note:</strong> This download link will expire in 7 days.</p>
		</div>
	`, downloadURL)

	return m.SendNotification(
		userEmail,
		username,
		"Data Export Ready",
		message,
		"info",
		1,
	)
}
