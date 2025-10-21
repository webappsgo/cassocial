package auth

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	// PasswordResetTokenExpiry is the duration a password reset token is valid
	PasswordResetTokenExpiry = 1 * time.Hour
	// EmailVerificationTokenExpiry is the duration an email verification token is valid
	EmailVerificationTokenExpiry = 24 * time.Hour
)

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetConfirm represents a password reset confirmation
type PasswordResetConfirm struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// PasswordChangeRequest represents a password change request
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// RequestPasswordReset creates a password reset token for a user
func (a *Auth) RequestPasswordReset(email string) (string, error) {
	// Get user by email
	user, err := a.GetUserByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal if email exists or not for security
			return "", nil
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Generate reset token
	token := generateRandomString(32)

	// Set token expiry
	expiry := time.Now().Add(PasswordResetTokenExpiry)

	// Store token in database
	query := `UPDATE users
			  SET password_reset_token = ?, password_reset_expires = ?, updated_at = ?
			  WHERE id = ?`

	if a.db.Driver == "postgres" {
		query = `UPDATE users
				 SET password_reset_token = $1, password_reset_expires = $2, updated_at = $3
				 WHERE id = $4`
	}

	_, err = a.db.Exec(query, token, expiry, time.Now(), user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to store reset token: %w", err)
	}

	return token, nil
}

// ValidatePasswordResetToken validates a password reset token
func (a *Auth) ValidatePasswordResetToken(token string) (string, error) {
	var userID string
	var expires sql.NullTime

	query := `SELECT id, password_reset_expires
			  FROM users
			  WHERE password_reset_token = ?`

	if a.db.Driver == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
	}

	err := a.db.QueryRow(query, token).Scan(&userID, &expires)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrInvalidToken
		}
		return "", fmt.Errorf("failed to validate token: %w", err)
	}

	// Check if token has expired
	if !expires.Valid || time.Now().After(expires.Time) {
		return "", ErrInvalidToken
	}

	return userID, nil
}

// ResetPassword resets a user's password using a reset token
func (a *Auth) ResetPassword(token, newPassword string) error {
	// Validate token and get user ID
	userID, err := a.ValidatePasswordResetToken(token)
	if err != nil {
		return err
	}

	// Validate new password
	if err := a.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	passwordHash, err := a.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password and clear reset token
	query := `UPDATE users
			  SET password_hash = ?, password_reset_token = NULL, password_reset_expires = NULL, updated_at = ?
			  WHERE id = ?`

	if a.db.Driver == "postgres" {
		query = `UPDATE users
				 SET password_hash = $1, password_reset_token = NULL, password_reset_expires = NULL, updated_at = $2
				 WHERE id = $3`
	}

	_, err = a.db.Exec(query, passwordHash, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	return nil
}

// ChangePassword changes a user's password (requires current password)
func (a *Auth) ChangePassword(userID, currentPassword, newPassword string) error {
	// Get user
	user, err := a.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := a.ComparePassword(user.PasswordHash, currentPassword); err != nil {
		return ErrInvalidCredentials
	}

	// Validate new password
	if err := a.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Check if new password is same as current
	if err := a.ComparePassword(user.PasswordHash, newPassword); err == nil {
		return fmt.Errorf("new password must be different from current password")
	}

	// Hash new password
	passwordHash, err := a.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`

	if a.db.Driver == "postgres" {
		query = `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
	}

	_, err = a.db.Exec(query, passwordHash, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	return nil
}

// GenerateEmailVerificationToken generates an email verification token for a user
func (a *Auth) GenerateEmailVerificationToken(userID string) (string, error) {
	// Get user to verify they exist
	_, err := a.GetUserByID(userID)
	if err != nil {
		return "", err
	}

	// Generate verification token
	token := generateRandomString(32)

	// In production, you would store this token in a separate table with expiry
	// For now, we'll use the password_reset_token field as a placeholder
	// A proper implementation would have a separate email_verification_tokens table

	expiry := time.Now().Add(EmailVerificationTokenExpiry)

	query := `UPDATE users
			  SET password_reset_token = ?, password_reset_expires = ?, updated_at = ?
			  WHERE id = ?`

	if a.db.Driver == "postgres" {
		query = `UPDATE users
				 SET password_reset_token = $1, password_reset_expires = $2, updated_at = $3
				 WHERE id = $4`
	}

	_, err = a.db.Exec(query, "EMAIL_"+token, expiry, time.Now(), userID)
	if err != nil {
		return "", fmt.Errorf("failed to store verification token: %w", err)
	}

	return token, nil
}

// VerifyEmail verifies a user's email using a verification token
func (a *Auth) VerifyEmail(token string) error {
	var userID string
	var expires sql.NullTime

	// Prefix token to distinguish from password reset tokens
	token = "EMAIL_" + token

	query := `SELECT id, password_reset_expires
			  FROM users
			  WHERE password_reset_token = ?`

	if a.db.Driver == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
	}

	err := a.db.QueryRow(query, token).Scan(&userID, &expires)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrInvalidVerificationToken
		}
		return fmt.Errorf("failed to validate token: %w", err)
	}

	// Check if token has expired
	if !expires.Valid || time.Now().After(expires.Time) {
		return ErrInvalidVerificationToken
	}

	// Mark email as verified and clear token
	updateQuery := `UPDATE users
					SET email_verified = ?, password_reset_token = NULL, password_reset_expires = NULL, updated_at = ?
					WHERE id = ?`

	if a.db.Driver == "postgres" {
		updateQuery = `UPDATE users
					   SET email_verified = $1, password_reset_token = NULL, password_reset_expires = NULL, updated_at = $2
					   WHERE id = $3`
	}

	_, err = a.db.Exec(updateQuery, true, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	return nil
}

// ResendVerificationEmail generates a new verification token and returns it
func (a *Auth) ResendVerificationEmail(email string) (string, error) {
	// Get user by email
	user, err := a.GetUserByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal if email exists or not
			return "", nil
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Check if already verified
	if user.EmailVerified {
		return "", fmt.Errorf("email already verified")
	}

	// Generate new verification token
	return a.GenerateEmailVerificationToken(user.ID)
}

// InvalidateAllPasswordResetTokens invalidates all password reset tokens for a user
func (a *Auth) InvalidateAllPasswordResetTokens(userID string) error {
	query := `UPDATE users
			  SET password_reset_token = NULL, password_reset_expires = NULL, updated_at = ?
			  WHERE id = ?`

	if a.db.Driver == "postgres" {
		query = `UPDATE users
				 SET password_reset_token = NULL, password_reset_expires = NULL, updated_at = $1
				 WHERE id = $2`
	}

	_, err := a.db.Exec(query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate tokens: %w", err)
	}

	return nil
}

// CheckPasswordStrength provides feedback on password strength
func (a *Auth) CheckPasswordStrength(password string) map[string]interface{} {
	strength := map[string]interface{}{
		"score":        0,
		"length":       len(password),
		"has_upper":    strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"),
		"has_lower":    strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz"),
		"has_number":   strings.ContainsAny(password, "0123456789"),
		"has_special":  strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?"),
		"feedback":     []string{},
	}

	score := 0
	feedback := []string{}

	// Length scoring
	if len(password) >= 8 {
		score += 1
	} else {
		feedback = append(feedback, "Password should be at least 8 characters")
	}

	if len(password) >= 12 {
		score += 1
	}

	if len(password) >= 16 {
		score += 1
	}

	// Character variety scoring
	if strength["has_upper"].(bool) {
		score += 1
	} else {
		feedback = append(feedback, "Add uppercase letters")
	}

	if strength["has_lower"].(bool) {
		score += 1
	} else {
		feedback = append(feedback, "Add lowercase letters")
	}

	if strength["has_number"].(bool) {
		score += 1
	} else {
		feedback = append(feedback, "Add numbers")
	}

	if strength["has_special"].(bool) {
		score += 1
	} else {
		feedback = append(feedback, "Add special characters")
	}

	// Cap score at 5
	if score > 5 {
		score = 5
	}

	strength["score"] = score
	strength["feedback"] = feedback

	// Add strength label
	switch score {
	case 0, 1:
		strength["label"] = "Very Weak"
	case 2:
		strength["label"] = "Weak"
	case 3:
		strength["label"] = "Fair"
	case 4:
		strength["label"] = "Strong"
	case 5:
		strength["label"] = "Very Strong"
	}

	return strength
}

// ForcePasswordChange marks a user as requiring a password change on next login
// This would require an additional field in the database
func (a *Auth) ForcePasswordChange(userID string) error {
	// This is a placeholder for future implementation
	// Would require adding a force_password_change column to users table
	return fmt.Errorf("force password change not implemented - requires database column")
}
