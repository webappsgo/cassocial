package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
	"github.com/casapps/cassocial/src/service"
)

// AuthHandlers handles authentication-related HTTP requests
type AuthHandlers struct {
	auth    *Auth
	db      *store.DB
	mailer  *service.Mailer
	siteURL string
}

// NewAuthHandlers creates a new AuthHandlers instance. mailer may be a
// disabled Mailer (IsEnabled() == false) when SMTP is not configured — per
// AI.md PART 18, callers must never attempt to send in that case; every
// send call below is already guarded by mailer.IsEnabled().
func NewAuthHandlers(authService *Auth, db *store.DB, mailer *service.Mailer, siteURL string) *AuthHandlers {
	return &AuthHandlers{
		auth:    authService,
		db:      db,
		mailer:  mailer,
		siteURL: siteURL,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login2FARequest represents a 2FA login request
type Login2FARequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// Enable2FAResponse represents the response for enabling 2FA
type Enable2FAResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

// Verify2FARequest represents a 2FA verification request
type Verify2FARequest struct {
	Code   string `json:"code"`
	Secret string `json:"secret,omitempty"`
}

// Register handles user registration
// POST /api/auth/register
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Check if registration is enabled
	registrationEnabled, _ := h.db.GetSetting("registration_enabled")
	if registrationEnabled == "false" {
		respondError(w, http.StatusForbidden, "registration is currently disabled")
		return
	}

	// Create user
	user, err := h.auth.Register(req.Username, req.Email, req.Password)
	if err != nil {
		switch err {
		case server.ErrUsernameExists:
			respondError(w, http.StatusConflict, "username already exists")
		case server.ErrEmailExists:
			respondError(w, http.StatusConflict, "email already exists")
		case server.ErrWeakPassword:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to create user")
		}
		return
	}

	// Generate token if user is active
	var token string
	if user.CanLogin() {
		token, _ = h.auth.GenerateToken(user)
	}

	// Send the welcome email (with an email-verification link when the new
	// user isn't already auto-verified). Per AI.md PART 18, no SMTP means no
	// email is attempted at all — mailer.IsEnabled() already reflects a live
	// connection test performed once at startup.
	if h.mailer.IsEnabled() {
		verificationURL := ""
		if !user.EmailVerified {
			if vToken, terr := h.auth.GenerateEmailVerificationToken(user.ID); terr == nil {
				verificationURL = h.siteURL + "/api/v1/auth/verify-email/" + vToken
			} else {
				log.Printf("failed to generate email verification token for %s: %v", user.Email, terr)
			}
		}
		if err := h.mailer.SendWelcome(user.Email, user.Username, verificationURL); err != nil {
			log.Printf("failed to send welcome email to %s: %v", user.Email, err)
		}
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"user":    user.SanitizeForJSON(),
		"token":   token,
		"message": "user created successfully",
	})
}

// Login handles user login
// POST /api/auth/login
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Attempt login
	token, user, err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		switch err {
		case server.ErrInvalidCredentials:
			respondError(w, http.StatusUnauthorized, "invalid credentials")
		case server.ErrUserNotActive:
			respondError(w, http.StatusForbidden, "user account is not active")
		case server.ErrEmailNotVerified:
			respondError(w, http.StatusForbidden, "email not verified")
		case server.Err2FARequired:
			// Return user ID for 2FA flow
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"requires_2fa": true,
				"user_id":      user.ID,
			})
		default:
			respondError(w, http.StatusInternalServerError, "login failed")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user.SanitizeForJSON(),
	})
}

// LoginWith2FA handles 2FA login
// POST /api/auth/login/2fa
func (h *AuthHandlers) LoginWith2FA(w http.ResponseWriter, r *http.Request) {
	var req Login2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Verify 2FA and login
	token, user, err := h.auth.LoginWith2FA(req.UserID, req.Code)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid 2FA code")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user.SanitizeForJSON(),
	})
}

// Logout handles user logout
// POST /api/auth/logout
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	// In a JWT-based system, logout is typically handled client-side
	// by removing the token. However, we can implement token blacklisting
	// if needed in the future.
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "logged out successfully",
	})
}

// RefreshToken handles token refresh
// POST /api/auth/refresh
func (h *AuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 {
		respondError(w, http.StatusUnauthorized, "missing authorization token")
		return
	}

	tokenString := authHeader[7:] // Remove "Bearer " prefix

	// Refresh token
	newToken, err := h.auth.RefreshToken(tokenString)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"token": newToken,
	})
}

// ForgotPassword handles forgot password request
// POST /api/auth/forgot-password
func (h *AuthHandlers) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Per AI.md PART 18: with no working SMTP, password reset is a hidden/
	// disabled feature — never attempt to generate or send a reset link.
	if !h.mailer.IsEnabled() {
		respondError(w, http.StatusServiceUnavailable, "password reset is unavailable: contact the administrator")
		return
	}

	// Generate password reset token
	token, err := h.auth.RequestPasswordReset(req.Email)
	if err != nil || token == "" {
		// Don't reveal if email exists or not
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"message": "if the email exists, a password reset link has been sent",
		})
		return
	}

	if user, uerr := h.auth.GetUserByEmail(req.Email); uerr == nil {
		resetURL := h.siteURL + "/reset-password?token=" + token
		if serr := h.mailer.SendPasswordReset(user.Email, user.Username, resetURL); serr != nil {
			log.Printf("failed to send password reset email to %s: %v", user.Email, serr)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "password reset link has been sent to your email",
	})
}

// ResetPassword handles password reset
// POST /api/auth/reset-password
func (h *AuthHandlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Resolve the user before the token is consumed, so a confirmation email
	// can be sent after a successful reset.
	userID, _ := h.auth.ValidatePasswordResetToken(req.Token)

	// Reset password
	err := h.auth.ResetPassword(req.Token, req.Password)
	if err != nil {
		switch err {
		case server.ErrInvalidToken:
			respondError(w, http.StatusBadRequest, "invalid or expired reset token")
		case server.ErrWeakPassword:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to reset password")
		}
		return
	}

	if h.mailer.IsEnabled() && userID != "" {
		if user, uerr := h.auth.GetUserByID(userID); uerr == nil {
			if serr := h.mailer.SendNotification(user.Email, user.Username,
				"Your Password Was Changed", "Your password was just changed. If this wasn't you, contact the administrator immediately.",
				"warning", 1); serr != nil {
				log.Printf("failed to send password-changed email to %s: %v", user.Email, serr)
			}
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "password reset successfully",
	})
}

// VerifyEmail handles email verification
// GET /api/auth/verify-email/{token}
func (h *AuthHandlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "missing verification token")
		return
	}

	// Verify email
	err := h.auth.VerifyEmail(token)
	if err != nil {
		switch err {
		case server.ErrInvalidVerificationToken:
			respondError(w, http.StatusBadRequest, "invalid or expired verification token")
		default:
			respondError(w, http.StatusInternalServerError, "failed to verify email")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "email verified successfully",
	})
}

// sendTwoFactorStateEmail notifies the user of a 2FA state change. Errors are
// logged, never returned — 2FA enable/disable must not fail on email issues.
func (h *AuthHandlers) sendTwoFactorStateEmail(userID, title, message string) {
	if !h.mailer.IsEnabled() {
		return
	}
	user, err := h.auth.GetUserByID(userID)
	if err != nil {
		return
	}
	if err := h.mailer.SendNotification(user.Email, user.Username, title, message, "warning", 1); err != nil {
		log.Printf("failed to send 2FA state-change email to %s: %v", user.Email, err)
	}
}

// Enable2FA handles enabling 2FA for the authenticated user
// POST /api/auth/2fa/enable
func (h *AuthHandlers) Enable2FA(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get user
	user, err := h.auth.GetUserByID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	// Check if 2FA is already enabled
	if user.TwoFactorEnabled {
		respondError(w, http.StatusBadRequest, "2FA is already enabled")
		return
	}

	// Generate 2FA secret
	setup, err := h.auth.Generate2FASecret(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate 2FA secret")
		return
	}

	respondJSON(w, http.StatusOK, Enable2FAResponse{
		Secret: setup.Secret,
		QRCode: setup.QRCodeURL,
	})
}

// Verify2FA handles verifying and enabling 2FA
// POST /api/auth/2fa/verify
func (h *AuthHandlers) Verify2FA(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req Verify2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Enable 2FA: verifies TOTP code, stores secret, generates and stores backup code hashes
	backupCodes, err := h.auth.Enable2FA(userID, req.Secret, req.Code)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid 2FA code")
		return
	}

	h.sendTwoFactorStateEmail(userID, "Two-Factor Authentication Enabled",
		"Two-factor authentication was just enabled on your account. If this wasn't you, contact the administrator immediately.")

	// backup_codes are shown to the user exactly once
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "2FA enabled successfully",
		"backup_codes": backupCodes,
	})
}

// Disable2FA handles disabling 2FA for the authenticated user
// POST /api/auth/2fa/disable
func (h *AuthHandlers) Disable2FA(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req Verify2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user
	user, err := h.auth.GetUserByID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify 2FA code before disabling
	valid, err := h.auth.Verify2FACode(user, req.Code)
	if err != nil || !valid {
		respondError(w, http.StatusBadRequest, "invalid 2FA code")
		return
	}

	// Disable 2FA
	err = h.auth.Disable2FA(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to disable 2FA")
		return
	}

	h.sendTwoFactorStateEmail(userID, "Two-Factor Authentication Disabled",
		"Two-factor authentication was just removed from your account. If this wasn't you, contact the administrator immediately.")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "2FA disabled successfully",
	})
}
