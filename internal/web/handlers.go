package web

import (
	"fmt"
	"net/http"

	"github.com/casapps/cassocial/internal/auth"
	"github.com/casapps/cassocial/internal/models"
)

// Index handler - landing page or profile directory
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	cookie, err := r.Cookie("session_id")
	var userData *UserData
	if err == nil {
		if session, valid := s.sessionManager.ValidateSession(cookie.Value); valid {
			userData = &UserData{
				ID:       session.UserID,
				Username: session.Username,
				Role:     session.Role,
			}
		}
	}

	data := &PageData{
		Title:       "Cassocial - Self-hosted Link Aggregator",
		Description: "Create your own link-in-bio landing page",
		User:        userData,
		Flash:       GetFlash(r),
	}

	s.templates.Render(w, "index.html", data)
}

// Setup Wizard Handlers

func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	// Check if already initialized
	initialized, _ := s.isInitialized()
	if initialized {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Redirect to welcome page
	http.Redirect(w, r, "/setup/welcome", http.StatusTemporaryRedirect)
}

func (s *Server) handleSetupWelcome(w http.ResponseWriter, r *http.Request) {
	data := &PageData{
		Title: "Welcome to Cassocial - Setup Wizard",
		Meta: map[string]interface{}{
			"Step":        1,
			"TotalSteps":  7,
			"StepName":    "Welcome",
		},
	}

	s.templates.Render(w, "setup/welcome.html", data)
}

func (s *Server) handleSetupBasic(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Process form
		instanceName := r.FormValue("instance_name")
		instanceURL := r.FormValue("instance_url")
		supportEmail := r.FormValue("support_email")
		timezone := r.FormValue("timezone")

		// Save settings
		s.db.SetSetting(models.SettingSiteName, instanceName)
		s.db.SetSetting(models.SettingSiteURL, instanceURL)
		s.db.SetSetting(models.SettingAdminEmail, supportEmail)
		s.db.SetSetting("timezone", timezone)

		// Redirect to next step
		http.Redirect(w, r, "/setup/domain", http.StatusSeeOther)
		return
	}

	// Auto-detect instance URL
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}
	defaultURL := fmt.Sprintf("%s://%s", protocol, r.Host)

	data := &PageData{
		Title: "Basic Configuration - Setup Wizard",
		Meta: map[string]interface{}{
			"Step":       2,
			"TotalSteps": 7,
			"StepName":   "Basic Configuration",
			"DefaultURL": defaultURL,
		},
	}

	s.templates.Render(w, "setup/basic.html", data)
}

func (s *Server) handleSetupDomain(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		profileStructure := r.FormValue("profile_structure")
		customDomainsEnabled := r.FormValue("custom_domains_enabled") == "on"
		sslConfig := r.FormValue("ssl_config")

		s.db.SetSetting("profile_url_structure", profileStructure)
		s.db.SetSetting(models.SettingEnableCustomDomains, fmt.Sprintf("%t", customDomainsEnabled))
		s.db.SetSetting("ssl_configuration", sslConfig)

		http.Redirect(w, r, "/setup/email", http.StatusSeeOther)
		return
	}

	data := &PageData{
		Title: "Domain & Access - Setup Wizard",
		Meta: map[string]interface{}{
			"Step":       3,
			"TotalSteps": 7,
			"StepName":   "Domain & Access",
		},
	}

	s.templates.Render(w, "setup/domain.html", data)
}

func (s *Server) handleSetupEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		provider := r.FormValue("smtp_provider")
		host := r.FormValue("smtp_host")
		port := r.FormValue("smtp_port")
		security := r.FormValue("smtp_security")
		user := r.FormValue("smtp_user")
		password := r.FormValue("smtp_password")
		fromName := r.FormValue("smtp_from_name")
		fromAddress := r.FormValue("smtp_from_address")

		// Save SMTP settings
		s.db.SetSetting(models.SettingSMTPProvider, provider)
		s.db.SetSetting(models.SettingSMTPHost, host)
		s.db.SetSetting(models.SettingSMTPPort, port)
		s.db.SetSetting(models.SettingSMTPSecurity, security)
		s.db.SetSetting(models.SettingSMTPUser, user)
		s.db.SetSetting(models.SettingSMTPPassword, password) // TODO: Encrypt
		s.db.SetSetting(models.SettingSMTPFromName, fromName)
		s.db.SetSetting(models.SettingSMTPFromAddress, fromAddress)
		s.db.SetSetting(models.SettingSMTPEnabled, "true")

		http.Redirect(w, r, "/setup/features", http.StatusSeeOther)
		return
	}

	data := &PageData{
		Title: "Email Configuration - Setup Wizard",
		Meta: map[string]interface{}{
			"Step":       4,
			"TotalSteps": 7,
			"StepName":   "Email Configuration",
			"Providers":  []string{"CUSTOM", "Gmail", "Yahoo", "Outlook"},
		},
	}

	s.templates.Render(w, "setup/email.html", data)
}

func (s *Server) handleSetupFeatures(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Registration settings
		s.db.SetSetting(models.SettingRegistrationEnabled, r.FormValue("registration_enabled"))
		s.db.SetSetting(models.SettingRegistrationRequiresApproval, r.FormValue("registration_requires_approval"))
		s.db.SetSetting(models.SettingEmailVerificationRequired, r.FormValue("email_verification_required"))

		// Limits
		s.db.SetSetting(models.SettingMaxProfilesPerUser, r.FormValue("max_profiles_per_user"))
		s.db.SetSetting(models.SettingMaxLinksPerProfile, r.FormValue("max_links_per_profile"))

		// Features
		s.db.SetSetting("analytics_enabled", r.FormValue("analytics_enabled"))
		s.db.SetSetting("qr_codes_enabled", r.FormValue("qr_codes_enabled"))
		s.db.SetSetting(models.SettingEnableCustomCSS, r.FormValue("custom_css_enabled"))
		s.db.SetSetting("api_enabled", r.FormValue("api_enabled"))

		// Privacy
		s.db.SetSetting("default_profile_public", r.FormValue("default_profile_public"))
		s.db.SetSetting(models.SettingAnalyticsAnonymousMode, r.FormValue("analytics_anonymous"))

		http.Redirect(w, r, "/setup/database", http.StatusSeeOther)
		return
	}

	data := &PageData{
		Title: "Features & Limits - Setup Wizard",
		Meta: map[string]interface{}{
			"Step":       5,
			"TotalSteps": 7,
			"StepName":   "Features & Limits",
		},
	}

	s.templates.Render(w, "setup/features.html", data)
}

func (s *Server) handleSetupDatabase(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Backup settings
		s.db.SetSetting(models.SettingBackupEnabled, r.FormValue("backup_enabled"))
		s.db.SetSetting(models.SettingBackupTime, r.FormValue("backup_time"))
		s.db.SetSetting(models.SettingBackupRetentionDays, r.FormValue("backup_retention_days"))

		http.Redirect(w, r, "/setup/complete", http.StatusSeeOther)
		return
	}

	data := &PageData{
		Title: "Database & Backup - Setup Wizard",
		Meta: map[string]interface{}{
			"Step":        6,
			"TotalSteps":  7,
			"StepName":    "Database & Backup",
			"DBType":      s.db.Driver,
		},
	}

	s.templates.Render(w, "setup/database.html", data)
}

func (s *Server) handleSetupComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Create first admin user
		username := r.FormValue("admin_username")
		email := r.FormValue("admin_email")
		password := r.FormValue("admin_password")

		user, err := s.auth.Register(username, email, password)
		if err != nil {
			data := &PageData{
				Title: "Complete Setup - Setup Wizard",
				Flash: &FlashMessage{
					Type:    "error",
					Message: fmt.Sprintf("Failed to create admin user: %v", err),
				},
				Meta: map[string]interface{}{
					"Step":       7,
					"TotalSteps": 7,
					"StepName":   "Complete Setup",
				},
			}
			s.templates.Render(w, "setup/complete.html", data)
			return
		}

		// Update user to admin role and active status
		query := "UPDATE users SET role = ?, status = ?, email_verified = ? WHERE id = ?"
		if s.db.Driver == "postgres" {
			query = "UPDATE users SET role = $1, status = $2, email_verified = $3 WHERE id = $4"
		}
		s.db.Exec(query, models.RoleAdmin, models.StatusActive, true, user.ID)

		// Mark as initialized
		s.db.SetSetting(models.SettingInitialized, "true")
		s.db.SetSetting(models.SettingSetupCompleted, "true")

		// Create session for admin user
		session, _ := s.sessionManager.CreateSession(user.ID, user.Username, models.RoleAdmin, getIPAddress(r), r.UserAgent())
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    session.ID,
			Path:     "/",
			MaxAge:   86400,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to admin dashboard
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	data := &PageData{
		Title: "Complete Setup - Setup Wizard",
		Meta: map[string]interface{}{
			"Step":       7,
			"TotalSteps": 7,
			"StepName":   "Complete Setup",
		},
	}

	s.templates.Render(w, "setup/complete.html", data)
}

// Authentication Handlers

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Attempt login
		token, user, err := s.auth.Login(username, password)
		if err != nil {
			if err == auth.Err2FARequired {
				// Redirect to 2FA page
				http.Redirect(w, r, "/login/2fa?user_id="+user.ID, http.StatusSeeOther)
				return
			}

			data := &PageData{
				Title: "Login - Cassocial",
				Flash: &FlashMessage{
					Type:    "error",
					Message: "Invalid username or password",
				},
			}
			s.templates.Render(w, "login.html", data)
			return
		}

		// Create session
		session, err := s.sessionManager.CreateSession(user.ID, user.Username, user.Role, getIPAddress(r), r.UserAgent())
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    session.ID,
			Path:     "/",
			MaxAge:   86400,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		// Store token (optional, for API access)
		_ = token

		// Redirect to dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := &PageData{
		Title: "Login - Cassocial",
		Flash: GetFlash(r),
	}

	s.templates.Render(w, "login.html", data)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	// Check if registration is enabled
	registrationEnabled, _ := s.db.GetSetting(models.SettingRegistrationEnabled)
	if registrationEnabled != "true" {
		http.Error(w, "Registration is disabled", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		if password != confirmPassword {
			data := &PageData{
				Title: "Register - Cassocial",
				Flash: &FlashMessage{
					Type:    "error",
					Message: "Passwords do not match",
				},
			}
			s.templates.Render(w, "register.html", data)
			return
		}

		_, err := s.auth.Register(username, email, password)
		if err != nil {
			data := &PageData{
				Title: "Register - Cassocial",
				Flash: &FlashMessage{
					Type:    "error",
					Message: fmt.Sprintf("Registration failed: %v", err),
				},
			}
			s.templates.Render(w, "register.html", data)
			return
		}

		SetFlash(w, "success", "Registration successful! Please log in.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := &PageData{
		Title: "Register - Cassocial",
		Flash: GetFlash(r),
	}

	s.templates.Render(w, "register.html", data)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		s.sessionManager.DestroySession(cookie.Value)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement password reset flow
	data := &PageData{
		Title: "Forgot Password - Cassocial",
	}
	s.templates.Render(w, "forgot-password.html", data)
}

func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement password reset flow
	data := &PageData{
		Title: "Reset Password - Cassocial",
	}
	s.templates.Render(w, "reset-password.html", data)
}

// Dashboard Handlers

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(auth.ContextKeyUserID).(string)
	username, _ := r.Context().Value(auth.ContextKeyUsername).(string)
	role, _ := r.Context().Value(auth.ContextKeyRole).(string)

	data := &PageData{
		Title: "Dashboard - Cassocial",
		User: &UserData{
			ID:       userID,
			Username: username,
			Role:     role,
		},
		Flash: GetFlash(r),
	}

	s.templates.Render(w, "dashboard/index.html", data)
}

func (s *Server) handleProfilesList(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(auth.ContextKeyUserID).(string)
	username, _ := r.Context().Value(auth.ContextKeyUsername).(string)
	role, _ := r.Context().Value(auth.ContextKeyRole).(string)

	// TODO: Fetch user's profiles from database

	data := &PageData{
		Title: "My Profiles - Cassocial",
		User: &UserData{
			ID:       userID,
			Username: username,
			Role:     role,
		},
		Meta: map[string]interface{}{
			"Profiles": []interface{}{}, // TODO: Add actual profiles
		},
	}

	s.templates.Render(w, "dashboard/profiles.html", data)
}

func (s *Server) handleProfileNew(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement profile creation
	data := &PageData{
		Title: "Create Profile - Cassocial",
	}
	s.templates.Render(w, "dashboard/profile-new.html", data)
}

func (s *Server) handleProfileEdit(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement profile editing
	data := &PageData{
		Title: "Edit Profile - Cassocial",
	}
	s.templates.Render(w, "dashboard/profile-edit.html", data)
}

func (s *Server) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement analytics dashboard
	data := &PageData{
		Title: "Analytics - Cassocial",
	}
	s.templates.Render(w, "dashboard/analytics.html", data)
}

func (s *Server) handleUserSettings(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement user settings
	data := &PageData{
		Title: "Settings - Cassocial",
	}
	s.templates.Render(w, "dashboard/settings.html", data)
}

// Admin Handlers

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	data := &PageData{
		Title: "Admin Dashboard - Cassocial",
		Meta: map[string]interface{}{
			"Stats": map[string]interface{}{
				"TotalUsers":    0, // TODO: Fetch from DB
				"TotalProfiles": 0,
				"TotalLinks":    0,
				"ActiveUsers":   s.sessionManager.GetActiveUsers(),
			},
		},
	}

	s.templates.Render(w, "admin/index.html", data)
}

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement user management
	data := &PageData{
		Title: "User Management - Admin",
	}
	s.templates.Render(w, "admin/users.html", data)
}

func (s *Server) handleAdminSettings(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement system settings
	data := &PageData{
		Title: "System Settings - Admin",
	}
	s.templates.Render(w, "admin/settings.html", data)
}

func (s *Server) handleAdminSMTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Update SMTP settings
		s.db.SetSetting(models.SettingSMTPHost, r.FormValue("smtp_host"))
		s.db.SetSetting(models.SettingSMTPPort, r.FormValue("smtp_port"))
		s.db.SetSetting(models.SettingSMTPUser, r.FormValue("smtp_user"))
		// TODO: Encrypt password before saving
		if password := r.FormValue("smtp_password"); password != "" {
			s.db.SetSetting(models.SettingSMTPPassword, password)
		}

		SetFlash(w, "success", "SMTP settings updated successfully")
		http.Redirect(w, r, "/admin/smtp", http.StatusSeeOther)
		return
	}

	// Get current SMTP settings
	smtpHost, _ := s.db.GetSetting(models.SettingSMTPHost)
	smtpPort, _ := s.db.GetSetting(models.SettingSMTPPort)
	smtpUser, _ := s.db.GetSetting(models.SettingSMTPUser)

	data := &PageData{
		Title: "SMTP Configuration - Admin",
		Meta: map[string]interface{}{
			"SMTPHost": smtpHost,
			"SMTPPort": smtpPort,
			"SMTPUser": smtpUser,
		},
		Flash: GetFlash(r),
	}

	s.templates.Render(w, "admin/smtp.html", data)
}

func (s *Server) handleAdminNotifications(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement notification preferences
	data := &PageData{
		Title: "Notification Preferences - Admin",
	}
	s.templates.Render(w, "admin/notifications.html", data)
}

func (s *Server) handleAdminBackup(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement backup management
	data := &PageData{
		Title: "Backup Management - Admin",
	}
	s.templates.Render(w, "admin/backup.html", data)
}

// Public profile handler
func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	// Extract username from URL path
	slug := r.URL.Path[len("/p/"):]
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	// TODO: Fetch profile from database
	// For now, return 404
	http.NotFound(w, r)
}
