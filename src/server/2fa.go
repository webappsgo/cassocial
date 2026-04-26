package server

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/casapps/cassocial/src/server/model"
)

const (
	// TOTPPeriod is the time period for TOTP codes (30 seconds)
	TOTPPeriod = 30
	// TOTPDigits is the number of digits in TOTP codes
	TOTPDigits = 6
)

// TwoFactorSetup contains the setup information for 2FA
type TwoFactorSetup struct {
	Secret     string   `json:"secret"`
	QRCodeURL  string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// Generate2FASecret generates a new 2FA secret for a user
func (a *Auth) Generate2FASecret(user *model.User) (*TwoFactorSetup, error) {
	// Generate a random secret (160 bits = 20 bytes)
	secret := generateRandomString(20)

	// Encode to base32 for TOTP compatibility
	encodedSecret := base32.StdEncoding.EncodeToString([]byte(secret))

	// Remove padding
	encodedSecret = strings.TrimRight(encodedSecret, "=")

	// Generate backup codes
	backupCodes := make([]string, 10)
	for i := 0; i < 10; i++ {
		backupCodes[i] = generateRandomString(8)
	}

	// Get site name from settings
	siteName, err := a.db.GetSetting("site_name")
	if err != nil || siteName == "" {
		siteName = "Cassocial"
	}

	// Generate QR code URL for authenticator apps
	// Format: otpauth://totp/SITE:USERNAME?secret=SECRET&issuer=SITE
	qrCodeURL := fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s",
		siteName,
		user.Username,
		encodedSecret,
		siteName,
	)

	return &TwoFactorSetup{
		Secret:      encodedSecret,
		QRCodeURL:   qrCodeURL,
		BackupCodes: backupCodes,
	}, nil
}

// Enable2FA enables two-factor authentication for a user
func (a *Auth) Enable2FA(userID, secret, code string) error {
	// Get user
	_, err := a.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Verify the code with the secret
	valid := a.verifyTOTP(secret, code)
	if !valid {
		return ErrInvalidCredentials
	}

	// Update user in database
	query := "UPDATE users SET two_factor_enabled = ?, two_factor_secret = ?, updated_at = ? WHERE id = ?"
	if a.db.Driver == "postgres" {
		query = "UPDATE users SET two_factor_enabled = $1, two_factor_secret = $2, updated_at = $3 WHERE id = $4"
	}

	_, err = a.db.Exec(query, true, secret, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	return nil
}

// Disable2FA disables two-factor authentication for a user
func (a *Auth) Disable2FA(userID, password string) error {
	// Get user
	user, err := a.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Verify password
	if err := a.ComparePassword(user.PasswordHash, password); err != nil {
		return ErrInvalidCredentials
	}

	// Update user in database
	query := "UPDATE users SET two_factor_enabled = ?, two_factor_secret = ?, updated_at = ? WHERE id = ?"
	if a.db.Driver == "postgres" {
		query = "UPDATE users SET two_factor_enabled = $1, two_factor_secret = $2, updated_at = $3 WHERE id = $4"
	}

	_, err = a.db.Exec(query, false, "", time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	return nil
}

// Verify2FACode verifies a 2FA code for a user
func (a *Auth) Verify2FACode(user *model.User, code string) (bool, error) {
	if !user.TwoFactorEnabled || user.TwoFactorSecret == "" {
		return false, fmt.Errorf("2FA not enabled for user")
	}

	// Verify TOTP code
	return a.verifyTOTP(user.TwoFactorSecret, code), nil
}

// verifyTOTP verifies a TOTP code against a secret
func (a *Auth) verifyTOTP(secret, code string) bool {
	// Add padding back to secret if needed
	if len(secret)%8 != 0 {
		secret = secret + strings.Repeat("=", 8-len(secret)%8)
	}

	// Decode secret from base32
	decodedSecret, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return false
	}

	// Current time counter
	counter := time.Now().Unix() / TOTPPeriod

	// Check current code and codes from previous and next time windows
	// This provides a buffer for clock skew
	for i := -1; i <= 1; i++ {
		expectedCode := a.generateTOTP(decodedSecret, counter+int64(i))
		if expectedCode == code {
			return true
		}
	}

	return false
}

// generateTOTP generates a TOTP code for a given secret and counter
func (a *Auth) generateTOTP(secret []byte, counter int64) string {
	// Convert counter to byte array
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	// Generate HMAC-SHA1
	h := hmac.New(sha1.New, secret)
	h.Write(buf)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0F
	truncatedHash := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF

	// Generate OTP
	otp := truncatedHash % uint32(math.Pow10(TOTPDigits))

	// Format with leading zeros
	return fmt.Sprintf("%0*d", TOTPDigits, otp)
}

// ValidateBackupCode validates a backup code
// Note: In a production system, backup codes should be hashed and stored in the database
func (a *Auth) ValidateBackupCode(userID, code string) (bool, error) {
	// This is a placeholder implementation
	// In production, you would:
	// 1. Query backup_codes table for the user
	// 2. Hash the provided code and compare with stored hashes
	// 3. Mark the code as used if valid
	// 4. Return whether the code was valid

	// For now, return false as backup codes need database implementation
	return false, fmt.Errorf("backup code validation not implemented - requires database table")
}

// RegenerateBackupCodes generates new backup codes for a user
func (a *Auth) RegenerateBackupCodes(userID string) ([]string, error) {
	// Get user to verify they exist
	_, err := a.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Generate new backup codes
	backupCodes := make([]string, 10)
	for i := 0; i < 10; i++ {
		backupCodes[i] = generateRandomString(8)
	}

	// In production, you would:
	// 1. Hash each backup code
	// 2. Delete old backup codes from database
	// 3. Store new hashed backup codes in database

	return backupCodes, nil
}

// Get2FAStatus returns the 2FA status for a user
func (a *Auth) Get2FAStatus(userID string) (bool, error) {
	user, err := a.GetUserByID(userID)
	if err != nil {
		return false, err
	}

	return user.TwoFactorEnabled, nil
}
