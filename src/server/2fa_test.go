package server

import (
	"encoding/base32"
	"strings"
	"testing"
	"time"

	"github.com/casapps/cassocial/src/server/model"
)

// TestGenerate2FASecret_ReturnsSetup verifies that a setup struct with non-empty fields is returned.
func TestGenerate2FASecret_ReturnsSetup(t *testing.T) {
	a := newTestAuth(t)
	user := &model.User{
		ID:       "test-id",
		Username: "2fatestuser",
	}

	setup, err := a.Generate2FASecret(user)
	if err != nil {
		t.Fatalf("Generate2FASecret: %v", err)
	}
	if setup.Secret == "" {
		t.Error("setup.Secret is empty")
	}
	if setup.QRCodeURL == "" {
		t.Error("setup.QRCodeURL is empty")
	}
	if len(setup.BackupCodes) != 10 {
		t.Errorf("len(BackupCodes) = %d, want 10", len(setup.BackupCodes))
	}
}

// TestGenerate2FASecret_QRCodeContainsUsername verifies the QR code URL contains the username.
func TestGenerate2FASecret_QRCodeContainsUsername(t *testing.T) {
	a := newTestAuth(t)
	user := &model.User{
		ID:       "test-id",
		Username: "qruser",
	}

	setup, err := a.Generate2FASecret(user)
	if err != nil {
		t.Fatalf("Generate2FASecret: %v", err)
	}
	if !strings.Contains(setup.QRCodeURL, "qruser") {
		t.Errorf("QRCodeURL %q does not contain username", setup.QRCodeURL)
	}
	if !strings.Contains(setup.QRCodeURL, "otpauth://totp/") {
		t.Errorf("QRCodeURL %q does not contain otpauth://totp/", setup.QRCodeURL)
	}
}

// TestGenerate2FASecret_UniqueBackupCodes verifies that backup codes are unique.
func TestGenerate2FASecret_UniqueBackupCodes(t *testing.T) {
	a := newTestAuth(t)
	user := &model.User{ID: "test-id", Username: "uniqueuser"}

	setup, err := a.Generate2FASecret(user)
	if err != nil {
		t.Fatalf("Generate2FASecret: %v", err)
	}

	seen := make(map[string]bool)
	for _, code := range setup.BackupCodes {
		if code == "" {
			t.Error("backup code is empty")
		}
		if seen[code] {
			t.Errorf("duplicate backup code: %q", code)
		}
		seen[code] = true
	}
}

// TestEnable2FA_InvalidCode verifies that an invalid TOTP code rejects enabling.
func TestEnable2FA_InvalidCode(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "enable2fauser", "enable2fa@example.com", "ValidPass1")

	setup, err := a.Generate2FASecret(user)
	if err != nil {
		t.Fatalf("Generate2FASecret: %v", err)
	}

	err = a.Enable2FA(user.ID, setup.Secret, "000000")
	if err != ErrInvalidCredentials {
		t.Errorf("Enable2FA with invalid code: got %v, want ErrInvalidCredentials", err)
	}
}

// TestEnable2FA_UserNotFound verifies that an unknown userID returns an error.
func TestEnable2FA_UserNotFound(t *testing.T) {
	a := newTestAuth(t)
	err := a.Enable2FA("nonexistent-id", "JBSWY3DPEHPK3PXP", "123456")
	if err == nil {
		t.Error("Enable2FA with unknown user should return error")
	}
}

// TestDisable2FA_Success verifies that 2FA can be disabled for an existing user.
func TestDisable2FA_Success(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "disable2fauser", "disable2fa@example.com", "ValidPass1")

	// Enable 2FA directly in DB so Disable2FA has something to clear.
	query := "UPDATE users SET two_factor_enabled = 1, two_factor_secret = 'JBSWY3DPEHPK3PXP' WHERE id = ?"
	if _, err := a.db.Exec(query, user.ID); err != nil {
		t.Fatalf("enabling 2FA in DB: %v", err)
	}

	if err := a.Disable2FA(user.ID); err != nil {
		t.Fatalf("Disable2FA: %v", err)
	}

	// Verify 2FA is disabled.
	got, err := a.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.TwoFactorEnabled {
		t.Error("TwoFactorEnabled should be false after Disable2FA")
	}
}

// TestDisable2FA_UnknownUser verifies that Disable2FA is a no-op for unknown user.
func TestDisable2FA_UnknownUser(t *testing.T) {
	a := newTestAuth(t)
	// Disable2FA on nonexistent user should not error (SQL UPDATE with no rows is fine).
	if err := a.Disable2FA("nonexistent-id"); err != nil {
		t.Errorf("Disable2FA unknown user: %v", err)
	}
}

// TestVerify2FACode_NotEnabled verifies that verification fails when 2FA is not enabled.
func TestVerify2FACode_NotEnabled(t *testing.T) {
	a := newTestAuth(t)
	user := &model.User{
		ID:               "test-id",
		Username:         "notenabled",
		TwoFactorEnabled: false,
		TwoFactorSecret:  "",
	}

	valid, err := a.Verify2FACode(user, "123456")
	if err == nil {
		t.Error("Verify2FACode should return error when 2FA not enabled")
	}
	if valid {
		t.Error("Verify2FACode should return false when 2FA not enabled")
	}
}

// TestVerify2FACode_ValidCode verifies that a real TOTP code passes.
func TestVerify2FACode_ValidCode(t *testing.T) {
	a := newTestAuth(t)

	// Use a known secret and generate a real code.
	const rawSecret = "JBSWY3DPEHPK3PXP"
	user := &model.User{
		ID:               "test-id",
		Username:         "totp-user",
		TwoFactorEnabled: true,
		TwoFactorSecret:  rawSecret,
	}

	// Pad the secret exactly as verifyTOTP does.
	paddedSecret := rawSecret
	if len(paddedSecret)%8 != 0 {
		paddedSecret += strings.Repeat("=", 8-len(paddedSecret)%8)
	}

	decoded, err := base32.StdEncoding.DecodeString(paddedSecret)
	if err != nil {
		t.Fatalf("base32 decode: %v", err)
	}

	// Generate the current TOTP code.
	counter := time.Now().Unix() / TOTPPeriod
	code := a.generateTOTP(decoded, counter)

	valid, err := a.Verify2FACode(user, code)
	if err != nil {
		t.Fatalf("Verify2FACode: %v", err)
	}
	if !valid {
		t.Error("Verify2FACode should return true for valid TOTP code")
	}
}

// TestVerifyTOTP_InvalidBase32 verifies that an invalid base32 secret returns false.
func TestVerifyTOTP_InvalidBase32(t *testing.T) {
	a := newTestAuth(t)
	// '!' is not a valid base32 character.
	result := a.verifyTOTP("NOT!VALID!BASE32", "123456")
	if result {
		t.Error("verifyTOTP with invalid base32 secret should return false")
	}
}

// TestGenerateTOTP_Deterministic verifies that the same inputs produce the same output.
func TestGenerateTOTP_Deterministic(t *testing.T) {
	a := newTestAuth(t)

	secret := []byte("hello-world-test")
	counter := int64(12345678)

	code1 := a.generateTOTP(secret, counter)
	code2 := a.generateTOTP(secret, counter)

	if code1 != code2 {
		t.Errorf("generateTOTP not deterministic: %q != %q", code1, code2)
	}
}

// TestGenerateTOTP_SixDigits verifies that the output is always 6 digits.
func TestGenerateTOTP_SixDigits(t *testing.T) {
	a := newTestAuth(t)
	secret := []byte("test-secret")

	for counter := int64(0); counter < 10; counter++ {
		code := a.generateTOTP(secret, counter)
		if len(code) != TOTPDigits {
			t.Errorf("generateTOTP counter=%d: len=%d, want %d", counter, len(code), TOTPDigits)
		}
	}
}

// TestGenerateTOTP_VariesWithCounter verifies different counters produce different codes.
func TestGenerateTOTP_VariesWithCounter(t *testing.T) {
	a := newTestAuth(t)
	secret := []byte("test-secret-value")

	codes := make(map[string]bool)
	for counter := int64(0); counter < 5; counter++ {
		code := a.generateTOTP(secret, counter)
		codes[code] = true
	}
	// With 5 different counters we expect at least 2 distinct codes.
	if len(codes) < 2 {
		t.Error("generateTOTP produced identical codes for all counters")
	}
}

// TestValidateBackupCode_NotImplemented verifies the placeholder returns false and an error.
func TestValidateBackupCode_NotImplemented(t *testing.T) {
	a := newTestAuth(t)
	valid, err := a.ValidateBackupCode("some-user-id", "some-code")
	if valid {
		t.Error("ValidateBackupCode should return false (not implemented)")
	}
	if err == nil {
		t.Error("ValidateBackupCode should return error (not implemented)")
	}
}

// TestRegenerateBackupCodes_Success verifies that 10 non-empty backup codes are returned.
func TestRegenerateBackupCodes_Success(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "rbc-user", "rbc@example.com", "ValidPass1")

	codes, err := a.RegenerateBackupCodes(user.ID)
	if err != nil {
		t.Fatalf("RegenerateBackupCodes: %v", err)
	}
	if len(codes) != 10 {
		t.Errorf("len(codes) = %d, want 10", len(codes))
	}
	for _, c := range codes {
		if c == "" {
			t.Error("backup code is empty")
		}
	}
}

// TestRegenerateBackupCodes_UnknownUser verifies that an unknown user returns an error.
func TestRegenerateBackupCodes_UnknownUser(t *testing.T) {
	a := newTestAuth(t)
	_, err := a.RegenerateBackupCodes("nonexistent-id")
	if err == nil {
		t.Error("RegenerateBackupCodes with unknown user should return error")
	}
}

// TestGet2FAStatus_Disabled verifies that a newly registered user reports 2FA as off.
func TestGet2FAStatus_Disabled(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "status2fa", "status2fa@example.com", "ValidPass1")

	enabled, err := a.Get2FAStatus(user.ID)
	if err != nil {
		t.Fatalf("Get2FAStatus: %v", err)
	}
	if enabled {
		t.Error("Get2FAStatus should return false for newly registered user")
	}
}

// TestGet2FAStatus_Enabled verifies that a user with 2FA enabled reports true.
func TestGet2FAStatus_Enabled(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "status2faon", "status2faon@example.com", "ValidPass1")

	query := "UPDATE users SET two_factor_enabled = 1, two_factor_secret = 'JBSWY3DPEHPK3PXP' WHERE id = ?"
	if _, err := a.db.Exec(query, user.ID); err != nil {
		t.Fatalf("enabling 2FA: %v", err)
	}

	enabled, err := a.Get2FAStatus(user.ID)
	if err != nil {
		t.Fatalf("Get2FAStatus: %v", err)
	}
	if !enabled {
		t.Error("Get2FAStatus should return true after enabling 2FA")
	}
}

// TestGet2FAStatus_UnknownUser verifies that an unknown user returns an error.
func TestGet2FAStatus_UnknownUser(t *testing.T) {
	a := newTestAuth(t)
	_, err := a.Get2FAStatus("nonexistent-id")
	if err == nil {
		t.Error("Get2FAStatus with unknown user should return error")
	}
}
