package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server"
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

	if len(req.Email) == 0 || len(req.Email) > 254 {
		h.renderError(w, http.StatusBadRequest, "Valid email address required")
		return
	}

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

	if err := h.db.CreateUser(user); err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate email verification token
	verificationToken, err := h.generateVerificationToken(req.Email)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to generate verification token")
		return
	}

	_ = verificationToken

	h.renderJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  "success",
		"message": "Registration successful. Please check your email to verify your account.",
		"user_id": user.ID,
	})
}

// HandleVerifyEmail handles email verification
func (h *UserHandler) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		h.renderError(w, http.StatusBadRequest, "Verification token required")
		return
	}

	record, err := h.db.GetEmailVerificationToken(token)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid or expired verification token")
		return
	}

	user, err := h.db.GetUserByID(record.UserID)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	user.EmailVerified = true
	if err := h.db.UpdateUser(user); err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to verify email")
		return
	}

	h.db.DeleteEmailVerificationToken(token)

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

	record, err := h.db.GetPasswordResetToken(req.Token)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid or expired reset token")
		return
	}

	newHash, err := server.HashPassword(req.NewPassword)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	user, err := h.db.GetUserByID(record.UserID)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	user.PasswordHash = newHash
	if err := h.db.UpdateUser(user); err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	h.db.DeletePasswordResetToken(req.Token)

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Password reset successfully. You can now login with your new password.",
	})
}

// HandleAccountSettings handles user account settings
func (h *UserHandler) HandleAccountSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		h.renderError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if r.Method == http.MethodGet {
		user, err := h.db.GetUserByID(userID)
		if err != nil {
			h.renderError(w, http.StatusInternalServerError, "Failed to retrieve user")
			return
		}
		settings := map[string]interface{}{
			"email":              user.Email,
			"username":           user.Username,
			"two_factor_enabled": user.TwoFactorEnabled,
		}
		h.renderJSON(w, http.StatusOK, settings)
		return
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			h.renderError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		user, err := h.db.GetUserByID(userID)
		if err != nil {
			h.renderError(w, http.StatusInternalServerError, "Failed to retrieve user")
			return
		}

		if email, ok := updates["email"].(string); ok && email != "" {
			user.Email = email
		}

		if err := h.db.UpdateUser(user); err != nil {
			h.renderError(w, http.StatusInternalServerError, "Failed to save settings")
			return
		}

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
	http.Error(w, "2FA setup not yet implemented", http.StatusNotImplemented)
}

// Helper functions

// generateVerificationToken generates an email verification token and stores it in the database.
func (h *UserHandler) generateVerificationToken(email string) (string, error) {
	user, err := h.db.GetUserByEmail(email)
	if err != nil {
		return "", err
	}

	token, err := server.GenerateRandomToken(32)
	if err != nil {
		return "", err
	}

	record := &store.EmailVerificationToken{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := h.db.CreateEmailVerificationToken(record); err != nil {
		return "", err
	}

	return token, nil
}

// generatePasswordResetToken generates a password reset token and stores it in the database.
func (h *UserHandler) generatePasswordResetToken(email string) (string, error) {
	user, err := h.db.GetUserByEmail(email)
	if err != nil {
		return "", err
	}

	token, err := server.GenerateRandomToken(32)
	if err != nil {
		return "", err
	}

	record := &store.PasswordResetToken{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	if err := h.db.CreatePasswordResetToken(record); err != nil {
		return "", err
	}

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
