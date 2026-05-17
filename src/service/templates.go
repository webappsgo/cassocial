package service

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

// EmailTemplate represents a base email template
type EmailTemplate struct {
	Subject     string
	HTMLBody    string
	TextBody    string
	Preheader   string
}

// TemplateData represents data passed to email templates
type TemplateData struct {
	SiteName    string
	SiteURL     string
	RecipientName string
	Year        int
	Data        map[string]interface{}
}

// NewTemplateData creates a new template data instance
func NewTemplateData(siteName, siteURL, recipientName string) *TemplateData {
	return &TemplateData{
		SiteName:      siteName,
		SiteURL:       siteURL,
		RecipientName: recipientName,
		Year:          time.Now().Year(),
		Data:          make(map[string]interface{}),
	}
}

// baseHTMLTemplate provides the base HTML structure for all emails.
// Declared as var (not const) so tests can substitute a broken template to
// exercise the error paths in renderHTML.
var baseHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="x-apple-disable-message-reformatting">
    <title>{{.Subject}}</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            background-color: #f4f4f7;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #ffffff;
            padding: 30px 20px;
            text-align: center;
            border-radius: 8px 8px 0 0;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content {
            background-color: #ffffff;
            padding: 30px 40px;
            border-radius: 0 0 8px 8px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }
        .content p {
            color: #51545e;
            margin: 16px 0;
        }
        .button {
            display: inline-block;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #ffffff;
            padding: 14px 32px;
            text-decoration: none;
            border-radius: 6px;
            margin: 20px 0;
            font-weight: 600;
        }
        .footer {
            text-align: center;
            padding: 20px;
            color: #6c757d;
            font-size: 14px;
        }
        .footer a {
            color: #667eea;
            text-decoration: none;
        }
        .divider {
            height: 1px;
            background-color: #e9ecef;
            margin: 24px 0;
        }
        .info-box {
            background-color: #f8f9fa;
            border-left: 4px solid #667eea;
            padding: 16px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .warning-box {
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 16px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .danger-box {
            background-color: #f8d7da;
            border-left: 4px solid #dc3545;
            padding: 16px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .code {
            background-color: #f4f4f7;
            padding: 2px 6px;
            border-radius: 4px;
            font-family: monospace;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.SiteName}}</h1>
        </div>
        <div class="content">
            {{.Content}}
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} {{.SiteName}}. All rights reserved.</p>
            <p>
                <a href="{{.SiteURL}}/privacy">Privacy Policy</a> &middot;
                <a href="{{.SiteURL}}/terms">Terms of Service</a>
            </p>
        </div>
    </div>
</body>
</html>`

// WelcomeEmail generates a welcome email for new users
func WelcomeEmail(data *TemplateData, verificationURL string) *EmailTemplate {
	data.Data["VerificationURL"] = verificationURL

	htmlContent := fmt.Sprintf(`
		<h2>Welcome to %s!</h2>
		<p>Hi %s,</p>
		<p>Thank you for signing up! We're excited to have you on board.</p>
		<p>To get started, please verify your email address by clicking the button below:</p>
		<p style="text-align: center;">
			<a href="%s" class="button">Verify Email Address</a>
		</p>
		<p>Or copy and paste this link into your browser:</p>
		<p class="code">%s</p>
		<div class="info-box">
			<p><strong>Note:</strong> This verification link will expire in 24 hours.</p>
		</div>
		<p>If you didn't create an account, please ignore this email.</p>
	`, data.SiteName, data.RecipientName, verificationURL, verificationURL)

	textContent := fmt.Sprintf(`Welcome to %s!

Hi %s,

Thank you for signing up! We're excited to have you on board.

To get started, please verify your email address by clicking this link:
%s

Note: This verification link will expire in 24 hours.

If you didn't create an account, please ignore this email.

---
© %d %s. All rights reserved.`,
		data.SiteName, data.RecipientName, verificationURL, data.Year, data.SiteName)

	return &EmailTemplate{
		Subject:   fmt.Sprintf("Welcome to %s - Verify Your Email", data.SiteName),
		HTMLBody:  renderHTML(data, htmlContent),
		TextBody:  textContent,
		Preheader: "Please verify your email address to get started",
	}
}

// PasswordResetEmail generates a password reset email
func PasswordResetEmail(data *TemplateData, resetURL string) *EmailTemplate {
	data.Data["ResetURL"] = resetURL

	htmlContent := fmt.Sprintf(`
		<h2>Password Reset Request</h2>
		<p>Hi %s,</p>
		<p>We received a request to reset your password for your %s account.</p>
		<p>Click the button below to reset your password:</p>
		<p style="text-align: center;">
			<a href="%s" class="button">Reset Password</a>
		</p>
		<p>Or copy and paste this link into your browser:</p>
		<p class="code">%s</p>
		<div class="warning-box">
			<p><strong>Important:</strong> This password reset link will expire in 1 hour.</p>
		</div>
		<p>If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
	`, data.RecipientName, data.SiteName, resetURL, resetURL)

	textContent := fmt.Sprintf(`Password Reset Request

Hi %s,

We received a request to reset your password for your %s account.

Click this link to reset your password:
%s

Important: This password reset link will expire in 1 hour.

If you didn't request a password reset, please ignore this email.

---
© %d %s. All rights reserved.`,
		data.RecipientName, data.SiteName, resetURL, data.Year, data.SiteName)

	return &EmailTemplate{
		Subject:   fmt.Sprintf("%s - Password Reset Request", data.SiteName),
		HTMLBody:  renderHTML(data, htmlContent),
		TextBody:  textContent,
		Preheader: "Reset your password",
	}
}

// EmailVerificationEmail generates an email verification email
func EmailVerificationEmail(data *TemplateData, verificationURL string) *EmailTemplate {
	data.Data["VerificationURL"] = verificationURL

	htmlContent := fmt.Sprintf(`
		<h2>Verify Your Email Address</h2>
		<p>Hi %s,</p>
		<p>Please verify your email address to continue using %s.</p>
		<p style="text-align: center;">
			<a href="%s" class="button">Verify Email</a>
		</p>
		<p>Or copy and paste this link into your browser:</p>
		<p class="code">%s</p>
		<div class="info-box">
			<p><strong>Note:</strong> This verification link will expire in 24 hours.</p>
		</div>
	`, data.RecipientName, data.SiteName, verificationURL, verificationURL)

	textContent := fmt.Sprintf(`Verify Your Email Address

Hi %s,

Please verify your email address to continue using %s.

Verification link:
%s

Note: This verification link will expire in 24 hours.

---
© %d %s. All rights reserved.`,
		data.RecipientName, data.SiteName, verificationURL, data.Year, data.SiteName)

	return &EmailTemplate{
		Subject:   fmt.Sprintf("%s - Verify Your Email", data.SiteName),
		HTMLBody:  renderHTML(data, htmlContent),
		TextBody:  textContent,
		Preheader: "Verify your email address",
	}
}

// TwoFactorCodeEmail generates a 2FA code email
func TwoFactorCodeEmail(data *TemplateData, code string) *EmailTemplate {
	data.Data["Code"] = code

	htmlContent := fmt.Sprintf(`
		<h2>Two-Factor Authentication Code</h2>
		<p>Hi %s,</p>
		<p>Your two-factor authentication code is:</p>
		<p style="text-align: center; font-size: 32px; font-weight: bold; letter-spacing: 4px; margin: 30px 0;">
			%s
		</p>
		<div class="warning-box">
			<p><strong>Important:</strong> This code will expire in 10 minutes.</p>
		</div>
		<p>If you didn't request this code, please contact support immediately.</p>
	`, data.RecipientName, code)

	textContent := fmt.Sprintf(`Two-Factor Authentication Code

Hi %s,

Your two-factor authentication code is:

%s

Important: This code will expire in 10 minutes.

If you didn't request this code, please contact support immediately.

---
© %d %s. All rights reserved.`,
		data.RecipientName, code, data.Year, data.SiteName)

	return &EmailTemplate{
		Subject:   fmt.Sprintf("%s - Two-Factor Authentication Code", data.SiteName),
		HTMLBody:  renderHTML(data, htmlContent),
		TextBody:  textContent,
		Preheader: fmt.Sprintf("Your 2FA code is: %s", code),
	}
}

// NotificationEmail generates a generic notification email
func NotificationEmail(data *TemplateData, title, message string, severity string) *EmailTemplate {
	data.Data["Title"] = title
	data.Data["Message"] = message
	data.Data["Severity"] = severity

	var boxClass string
	switch severity {
	case "emergency", "critical":
		boxClass = "danger-box"
	case "warning":
		boxClass = "warning-box"
	default:
		boxClass = "info-box"
	}

	htmlContent := fmt.Sprintf(`
		<h2>%s</h2>
		<p>Hi %s,</p>
		<div class="%s">
			%s
		</div>
		<p>For more information, please visit your <a href="%s/admin/dashboard">admin dashboard</a>.</p>
	`, title, data.RecipientName, boxClass, message, data.SiteURL)

	textContent := fmt.Sprintf(`%s

Hi %s,

%s

For more information, please visit your admin dashboard:
%s/admin/dashboard

---
© %d %s. All rights reserved.`,
		title, data.RecipientName, stripHTML(message), data.SiteURL, data.Year, data.SiteName)

	prefix := ""
	switch severity {
	case "emergency", "critical":
		prefix = "[URGENT] "
	case "warning":
		prefix = "[WARNING] "
	}

	return &EmailTemplate{
		Subject:   fmt.Sprintf("%s%s - %s", prefix, data.SiteName, title),
		HTMLBody:  renderHTML(data, htmlContent),
		TextBody:  textContent,
		Preheader: title,
	}
}

// renderHTML renders the HTML template with content
func renderHTML(data *TemplateData, content string) string {
	tmpl, err := template.New("email").Parse(baseHTMLTemplate)
	if err != nil {
		return content // Fallback to plain content
	}

	var buf strings.Builder
	templateData := map[string]interface{}{
		"SiteName": data.SiteName,
		"SiteURL":  data.SiteURL,
		"Year":     data.Year,
		"Content":  template.HTML(content),
	}

	if err := tmpl.Execute(&buf, templateData); err != nil {
		return content // Fallback to plain content
	}

	return buf.String()
}

// stripHTML removes HTML tags from a string (basic implementation)
func stripHTML(input string) string {
	// Basic HTML tag removal
	result := input
	result = strings.ReplaceAll(result, "<br>", "\n")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br />", "\n")
	result = strings.ReplaceAll(result, "</p>", "\n")

	// Remove all remaining tags
	inTag := false
	var stripped strings.Builder
	for _, r := range result {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			stripped.WriteRune(r)
		}
	}

	return stripped.String()
}

// TeamInviteEmail generates a team invitation email
func TeamInviteEmail(data *TemplateData, orgName, role, inviteURL string) *EmailTemplate {
	data.Data["OrgName"] = orgName
	data.Data["Role"] = role
	data.Data["InviteURL"] = inviteURL

	htmlContent := fmt.Sprintf(`
		<h2>Team Invitation</h2>
		<p>Hi %s,</p>
		<p>You've been invited to join <strong>%s</strong> on %s as a <strong>%s</strong>.</p>
		<p style="text-align: center;">
			<a href="%s" class="button">Accept Invitation</a>
		</p>
		<p>Or copy and paste this link into your browser:</p>
		<p class="code">%s</p>
		<div class="info-box">
			<p><strong>Note:</strong> This invitation will expire in 7 days.</p>
		</div>
	`, data.RecipientName, orgName, data.SiteName, role, inviteURL, inviteURL)

	textContent := fmt.Sprintf(`Team Invitation

Hi %s,

You've been invited to join %s on %s as a %s.

Accept invitation:
%s

Note: This invitation will expire in 7 days.

---
© %d %s. All rights reserved.`,
		data.RecipientName, orgName, data.SiteName, role, inviteURL, data.Year, data.SiteName)

	return &EmailTemplate{
		Subject:   fmt.Sprintf("%s - Invitation to Join %s", data.SiteName, orgName),
		HTMLBody:  renderHTML(data, htmlContent),
		TextBody:  textContent,
		Preheader: fmt.Sprintf("You've been invited to join %s", orgName),
	}
}
