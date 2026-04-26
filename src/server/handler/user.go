package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/model"
	"github.com/casapps/cassocial/src/server/store"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	config *config.Config
	db     *store.DB
}

// NewUserHandler creates a new user handler
func NewUserHandler(cfg *config.Config, db *store.DB) *UserHandler {
	return &UserHandler{
		config: cfg,
		db:     db,
	}
}

// RegistrationRequest represents a user registration request
type RegistrationRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// HandleRegister handles user registration
func (h *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if registration is allowed
	if !h.config.Cassocial.AllowRegistration {
		h.renderError(w, http.StatusForbidden, "Registration is disabled")
		return
	}

	// Parse request
	var req RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate input
	if len(req.Username) < 3 || len(req.Username) > 30 {
		h.renderError(w, http.StatusBadRequest, "Username must be between 3 and 30 characters")
		return
	}

	if len(req.Password) < 8 {
		h.renderError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// TODO: Validate email format

	// Hash password using Argon2id
	passwordHash, err := server.HashPassword(req.Password)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Create user
	user := &store.User{
		Username:      req.Username,
		Email:         req.Email,
		PasswordHash:  passwordHash,
		Role:          "user",
		Status:        "pending",
		EmailVerified: false,
	}

	// TODO: Call db.CreateUser(user)
	// For now, return success
	_ = user

	// Generate email verification token
	verificationToken, err := h.generateVerificationToken(req.Email)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to generate verification token")
		return
	}

	// TODO: Send verification email
	_ = verificationToken

	h.renderJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  "success",
		"message": "Registration successful. Please check your email to verify your account.",
		"user_id": "temp-id", // TODO: Return actual user ID
	})
}

// HandleVerifyEmail handles email verification
func (h *UserHandler) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		h.renderError(w, http.StatusBadRequest, "Verification token required")
		return
	}

	// TODO: Validate token and mark email as verified
	// For now, return success
	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Email verified successfully. You can now login.",
	})
}

// HandleRequestPasswordReset handles password reset requests
func (h *UserHandler) HandleRequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Generate reset token
	resetToken, err := h.generatePasswordResetToken(req.Email)
	if err != nil {
		// Don't reveal if email exists - always return success
		h.renderJSON(w, http.StatusOK, map[string]string{
			"status":  "success",
			"message": "If your email is registered, you will receive a password reset link.",
		})
		return
	}

	// TODO: Send reset email
	_ = resetToken

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "If your email is registered, you will receive a password reset link.",
	})
}

// HandleResetPassword handles password reset with token
func (h *UserHandler) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate new password
	if len(req.NewPassword) < 8 {
		h.renderError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// TODO: Validate token
	// TODO: Hash new password with Argon2id
	// TODO: Update user password

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Password reset successfully. You can now login with your new password.",
	})
}

// HandleAccountSettings handles user account settings
func (h *UserHandler) HandleAccountSettings(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user from session
	// TODO: Return user settings

	if r.Method == http.MethodGet {
		// Return current settings
		settings := map[string]interface{}{
			"email":              "user@example.com",
			"username":           "user",
			"two_factor_enabled": false,
		}
		h.renderJSON(w, http.StatusOK, settings)
		return
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		// Update settings
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			h.renderError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		// TODO: Validate and save settings
		h.renderJSON(w, http.StatusOK, map[string]string{
			"status":  "success",
			"message": "Settings updated successfully",
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// Handle2FASetup handles 2FA setup
func (h *UserHandler) Handle2FASetup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Generate TOTP secret and QR code
		// TODO: Generate secret, QR code URL
		response := map[string]interface{}{
			"secret":  "TEMP-SECRET",
			"qr_code": "data:image/png;base64,...",
		}
		h.renderJSON(w, http.StatusOK, response)
		return
	}

	if r.Method == http.MethodPost {
		// Verify and enable 2FA
		var req struct {
			Code string `json:"code"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.renderError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		// TODO: Verify TOTP code and enable 2FA
		h.renderJSON(w, http.StatusOK, map[string]interface{}{
			"status":       "success",
			"message":      "2FA enabled successfully",
			"backup_codes": []string{"code1", "code2", "code3"},
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// Helper functions

// generateVerificationToken generates an email verification token
func (h *UserHandler) generateVerificationToken(email string) (string, error) {
	token, err := server.GenerateRandomToken(32)
	if err != nil {
		return "", err
	}

	// TODO: Store token in database with expiry
	return token, nil
}

// generatePasswordResetToken generates a password reset token
func (h *UserHandler) generatePasswordResetToken(email string) (string, error) {
	// TODO: Check if email exists
	// Don't reveal if email doesn't exist for security

	token, err := server.GenerateRandomToken(32)
	if err != nil {
		return "", err
	}

	// TODO: Store token in database with expiry (1 hour)
	return token, nil
}

// renderJSON renders a JSON response
func (h *UserHandler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// renderError renders an error response
func (h *UserHandler) renderError(w http.ResponseWriter, status int, message string) {
	h.renderJSON(w, status, map[string]string{
		"error": message,
	})
}
