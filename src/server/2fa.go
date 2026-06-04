package server

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
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
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
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

	// Generate plaintext backup codes (shown to user once, stored as hashes)
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

// Enable2FA enables two-factor authentication for a user.
// It verifies the TOTP code, stores the secret, generates and stores hashed backup codes,
// and returns the plaintext backup codes (shown to user exactly once).
func (a *Auth) Enable2FA(userID, secret, code string) ([]string, error) {
	// Get user
	_, err := a.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Verify the code with the secret
	valid := a.verifyTOTP(secret, code)
	if !valid {
		return nil, ErrInvalidCredentials
	}

	// Update user in database
	query := "UPDATE users SET two_factor_enabled = ?, two_factor_secret = ?, updated_at = ? WHERE id = ?"
	if a.db.Driver == "postgres" {
		query = "UPDATE users SET two_factor_enabled = $1, two_factor_secret = $2, updated_at = $3 WHERE id = $4"
	}

	_, err = a.db.ExecR(query, true, secret, a.db.BindTime(time.Now()), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to enable 2FA: %w", err)
	}

	// Generate 10 backup codes, store hashes, return plaintext (shown once)
	backupCodes := make([]string, 10)
	hashes := make([]string, 10)
	for i := 0; i < 10; i++ {
		backupCodes[i] = generateRandomString(8)
		hashes[i] = hashBackupCode(backupCodes[i])
	}
	if err := a.db.StoreBackupCodes(userID, hashes); err != nil {
		return nil, fmt.Errorf("failed to store backup codes: %w", err)
	}

	return backupCodes, nil
}

// Disable2FA disables two-factor authentication for a user.
// The caller is responsible for verifying user identity before calling this.
func (a *Auth) Disable2FA(userID string) error {
	// Update user in database
	query := "UPDATE users SET two_factor_enabled = ?, two_factor_secret = ?, updated_at = ? WHERE id = ?"
	if a.db.Driver == "postgres" {
		query = "UPDATE users SET two_factor_enabled = $1, two_factor_secret = $2, updated_at = $3 WHERE id = $4"
	}

	if _, err := a.db.ExecR(query, false, "", a.db.BindTime(time.Now()), userID); err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	// Remove stored backup codes
	if err := a.db.DeleteBackupCodes(userID); err != nil {
		return fmt.Errorf("failed to delete backup codes: %w", err)
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

// ValidateBackupCode validates a backup code for a user, marking it used on success
func (a *Auth) ValidateBackupCode(userID, code string) (bool, error) {
	codes, err := a.db.GetUnusedBackupCodes(userID)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve backup codes: %w", err)
	}

	codeHash := hashBackupCode(code)
	for _, bc := range codes {
		if hmac.Equal([]byte(bc.CodeHash), []byte(codeHash)) {
			if err := a.db.MarkBackupCodeUsed(bc.ID); err != nil {
				return false, fmt.Errorf("failed to mark backup code used: %w", err)
			}
			return true, nil
		}
	}

	return false, nil
}

// RegenerateBackupCodes generates new backup codes for a user, replaces old ones
func (a *Auth) RegenerateBackupCodes(userID string) ([]string, error) {
	// Get user to verify they exist
	_, err := a.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Generate new plaintext backup codes
	backupCodes := make([]string, 10)
	for i := 0; i < 10; i++ {
		backupCodes[i] = generateRandomString(8)
	}

	// Hash and store them
	hashes := make([]string, len(backupCodes))
	for i, bc := range backupCodes {
		hashes[i] = hashBackupCode(bc)
	}
	if err := a.db.StoreBackupCodes(userID, hashes); err != nil {
		return nil, fmt.Errorf("failed to store backup codes: %w", err)
	}

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

// hashBackupCode hashes a backup code with SHA-256
func hashBackupCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}

