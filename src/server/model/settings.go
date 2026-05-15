package model

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/casapps/cassocial/src/config"
)

// Setting represents a key-value configuration setting
type Setting struct {
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

var (
	ErrSettingNotFound = errors.New("setting not found")
	ErrInvalidValue    = errors.New("invalid value for setting")
)

// GetString returns the value as a string
func (s *Setting) GetString() string {
	return s.Value
}

// GetInt returns the value as an integer
func (s *Setting) GetInt() (int, error) {
	return strconv.Atoi(s.Value)
}

// GetBool returns the value as a boolean
func (s *Setting) GetBool() (bool, error) {
	return config.ParseBool(s.Value), nil
}

// GetFloat returns the value as a float64
func (s *Setting) GetFloat() (float64, error) {
	return strconv.ParseFloat(s.Value, 64)
}

// GetJSON unmarshals the value into the provided interface
func (s *Setting) GetJSON(v interface{}) error {
	return json.Unmarshal([]byte(s.Value), v)
}

// SetString sets the value from a string
func (s *Setting) SetString(value string) {
	s.Value = value
	s.UpdatedAt = time.Now()
}

// SetInt sets the value from an integer
func (s *Setting) SetInt(value int) {
	s.Value = strconv.Itoa(value)
	s.UpdatedAt = time.Now()
}

// SetBool sets the value from a boolean
func (s *Setting) SetBool(value bool) {
	s.Value = strconv.FormatBool(value)
	s.UpdatedAt = time.Now()
}

// SetFloat sets the value from a float64
func (s *Setting) SetFloat(value float64) {
	s.Value = strconv.FormatFloat(value, 'f', -1, 64)
	s.UpdatedAt = time.Now()
}

// SetJSON sets the value from a JSON-encodable interface
func (s *Setting) SetJSON(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	s.Value = string(b)
	s.UpdatedAt = time.Now()
	return nil
}

// Common setting keys as constants
const (
	// System Configuration
	SettingSiteName              = "site_name"
	SettingSiteURL               = "site_url"
	SettingInitialized           = "initialized"
	SettingSetupCompleted        = "setup_completed"
	SettingMaintenanceMode       = "maintenance_mode"
	SettingMaintenanceMessage    = "maintenance_message"
	SettingMaintenanceBypassIPs  = "maintenance_bypass_ips"

	// Registration & Authentication
	SettingRegistrationEnabled          = "registration_enabled"
	SettingRegistrationRequiresApproval = "registration_requires_approval"
	SettingEmailVerificationRequired    = "email_verification_required"
	SettingSessionTimeoutMinutes        = "session_timeout_minutes"
	SettingTwoFactorEnabled             = "two_factor_enabled"

	// Password Requirements
	SettingPasswordMinLength        = "password_min_length"
	SettingPasswordRequireUppercase = "password_require_uppercase"
	SettingPasswordRequireNumber    = "password_require_number"
	SettingPasswordRequireSpecial   = "password_require_special"

	// Limits
	SettingMaxLinksPerProfile   = "max_links_per_profile"
	SettingMaxProfilesPerUser   = "max_profiles_per_user"
	SettingUploadMaxSizeMB      = "upload_max_size_mb"
	SettingAllowedImageTypes    = "allowed_image_types"

	// Performance
	SettingCacheTTLSeconds          = "cache_ttl_seconds"
	SettingRateLimitRequests        = "rate_limit_requests"
	SettingRateLimitWindowSeconds   = "rate_limit_window_seconds"

	// Backups
	SettingBackupEnabled        = "backup_enabled"
	SettingBackupRetentionDays  = "backup_retention_days"
	SettingBackupTime           = "backup_time"

	// Analytics
	SettingAnalyticsRetentionDays = "analytics_retention_days"
	SettingAnalyticsAnonymousMode = "analytics_anonymous_mode"
	SettingAnalyticsSamplingRate  = "analytics_sampling_rate"

	// Features
	SettingDefaultThemeID         = "default_theme_id"
	SettingEnableCustomCSS        = "enable_custom_css"
	SettingEnableCustomDomains    = "enable_custom_domains"
	SettingQRCodeSize             = "qr_code_size"

	// SSL
	SettingSSLAutoRenew        = "ssl_auto_renew"
	SettingSSLRenewDaysBefore  = "ssl_renew_days_before"

	// SMTP Configuration
	SettingSMTPProvider      = "smtp_provider"
	SettingSMTPHost          = "smtp_host"
	SettingSMTPPort          = "smtp_port"
	SettingSMTPSecurity      = "smtp_security"
	SettingSMTPUser          = "smtp_user"
	SettingSMTPPassword      = "smtp_password"
	SettingSMTPFromName      = "smtp_from_name"
	SettingSMTPFromAddress   = "smtp_from_address"
	SettingAdminEmail        = "admin_email"
	SettingSMTPEnabled       = "smtp_enabled"
	SettingSMTPRetryCount    = "smtp_retry_count"
	SettingSMTPRetryDelay    = "smtp_retry_delay"

	// Notifications
	SettingNotifyEmergency          = "notify_emergency"
	SettingNotifyCertificate        = "notify_certificate"
	SettingNotifyBugReport          = "notify_bug_report"
	SettingNotifyUserRegistration   = "notify_user_registration"
	SettingNotifyDomainVerification = "notify_domain_verification"
	SettingNotifyBackupStatus       = "notify_backup_status"
	SettingNotifyHighTraffic        = "notify_high_traffic"
	SettingNotificationBatchDelay   = "notification_batch_delay"

	// Footer & Gradients
	SettingDefaultFooter    = "default_footer"
	SettingGradientPresets  = "gradient_presets"
)

// SMTPConfig represents SMTP configuration
type SMTPConfig struct {
	Provider    string `json:"provider"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Security    string `json:"security"`
	User        string `json:"user"`
	Password    string `json:"password"`
	FromName    string `json:"from_name"`
	FromAddress string `json:"from_address"`
	AdminEmail  string `json:"admin_email"`
	Enabled     bool   `json:"enabled"`
	RetryCount  int    `json:"retry_count"`
	RetryDelay  int    `json:"retry_delay"`
}

// NotificationPreferences represents notification settings
type NotificationPreferences struct {
	Emergency          bool `json:"notify_emergency"`
	Certificate        bool `json:"notify_certificate"`
	BugReport          bool `json:"notify_bug_report"`
	UserRegistration   bool `json:"notify_user_registration"`
	DomainVerification bool `json:"notify_domain_verification"`
	BackupStatus       bool `json:"notify_backup_status"`
	HighTraffic        bool `json:"notify_high_traffic"`
	BatchDelay         int  `json:"notification_batch_delay"`
}

// GradientPreset represents a gradient preset
type GradientPreset struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Validate validates SMTP configuration
func (sc *SMTPConfig) Validate() error {
	if sc.Host == "" {
		return errors.New("SMTP host is required")
	}
	if sc.Port <= 0 || sc.Port > 65535 {
		return errors.New("invalid SMTP port")
	}
	if sc.FromAddress == "" {
		return errors.New("SMTP from address is required")
	}
	if sc.User != "" && sc.Password == "" {
		return errors.New("SMTP password is required when user is set")
	}
	return nil
}
