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

// TestEnable2FA_Success verifies that a correct TOTP code enables 2FA and persists the secret.
func TestEnable2FA_Success(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "enable2fasuccess", "enable2fasuccess@example.com", "ValidPass1")

	setup, err := a.Generate2FASecret(user)
	if err != nil {
		t.Fatalf("Generate2FASecret: %v", err)
	}

	// Pad and decode secret the same way verifyTOTP does.
	paddedSecret := setup.Secret
	if len(paddedSecret)%8 != 0 {
		paddedSecret += strings.Repeat("=", 8-len(paddedSecret)%8)
	}
	decoded, err := base32.StdEncoding.DecodeString(paddedSecret)
	if err != nil {
		t.Fatalf("base32 decode: %v", err)
	}
	counter := time.Now().Unix() / TOTPPeriod
	code := a.generateTOTP(decoded, counter)

	backupCodes, err := a.Enable2FA(user.ID, setup.Secret, code)
	if err != nil {
		t.Fatalf("Enable2FA valid code: %v", err)
	}
	if len(backupCodes) != 10 {
		t.Errorf("Enable2FA returned %d backup codes, want 10", len(backupCodes))
	}

	// Verify it was actually stored in the DB.
	got, err := a.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if !got.TwoFactorEnabled {
		t.Error("TwoFactorEnabled should be true after Enable2FA")
	}
	if got.TwoFactorSecret != setup.Secret {
		t.Errorf("TwoFactorSecret = %q, want %q", got.TwoFactorSecret, setup.Secret)
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

	_, err = a.Enable2FA(user.ID, setup.Secret, "000000")
	if err != ErrInvalidCredentials {
		t.Errorf("Enable2FA with invalid code: got %v, want ErrInvalidCredentials", err)
	}
}

// TestEnable2FA_UserNotFound verifies that an unknown userID returns an error.
func TestEnable2FA_UserNotFound(t *testing.T) {
	a := newTestAuth(t)
	_, err := a.Enable2FA("nonexistent-id", "JBSWY3DPEHPK3PXP", "123456")
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
	if _, err := a.db.ExecR(query, user.ID); err != nil {
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

// TestValidateBackupCode_Success verifies that a valid backup code is accepted and marked used.
func TestValidateBackupCode_Success(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "backupcode-user", "backupcode@example.com", "ValidPass1")

	codes, err := a.RegenerateBackupCodes(user.ID)
	if err != nil {
		t.Fatalf("RegenerateBackupCodes: %v", err)
	}

	// First use — should succeed
	valid, err := a.ValidateBackupCode(user.ID, codes[0])
	if err != nil {
		t.Fatalf("ValidateBackupCode: %v", err)
	}
	if !valid {
		t.Error("ValidateBackupCode should return true for a valid unused code")
	}

	// Second use of same code — should fail (marked used)
	valid, err = a.ValidateBackupCode(user.ID, codes[0])
	if err != nil {
		t.Fatalf("ValidateBackupCode second use: %v", err)
	}
	if valid {
		t.Error("ValidateBackupCode should return false for an already-used code")
	}
}

// TestValidateBackupCode_Invalid verifies that a wrong code is rejected.
func TestValidateBackupCode_Invalid(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "backupcode-invalid", "backupcode-invalid@example.com", "ValidPass1")

	if _, err := a.RegenerateBackupCodes(user.ID); err != nil {
		t.Fatalf("RegenerateBackupCodes: %v", err)
	}

	valid, err := a.ValidateBackupCode(user.ID, "wrong-code-xxx")
	if err != nil {
		t.Fatalf("ValidateBackupCode invalid: %v", err)
	}
	if valid {
		t.Error("ValidateBackupCode should return false for an invalid code")
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

// TestGenerate2FASecret_MissingSiteName verifies that Generate2FASecret falls back to
// "Cassocial" when site_name is missing from the settings table.
func TestGenerate2FASecret_MissingSiteName(t *testing.T) {
	a := newTestAuth(t)
	if _, err := a.db.Exec("DELETE FROM settings WHERE key = 'site_name'"); err != nil {
		t.Fatalf("DELETE settings: %v", err)
	}

	user := &model.User{
		ID:       "test-id",
		Username: "nositeuser",
	}
	setup, err := a.Generate2FASecret(user)
	if err != nil {
		t.Fatalf("Generate2FASecret with missing site_name: %v", err)
	}
	if setup.Secret == "" {
		t.Error("setup.Secret is empty")
	}
	// QR code should use the fallback site name "Cassocial".
	if !strings.Contains(setup.QRCodeURL, "Cassocial") {
		t.Errorf("QRCodeURL %q does not contain fallback site name 'Cassocial'", setup.QRCodeURL)
	}
}

// TestVerifyTOTP_WithPadding verifies that verifyTOTP correctly pads a secret whose
// length is not a multiple of 8 before base32 decoding.
func TestVerifyTOTP_WithPadding(t *testing.T) {
	a := newTestAuth(t)

	// "JBSWY3D" is 7 chars — not a multiple of 8, so padding must be added.
	// Re-encode to get a valid base32 string of non-padded length.
	// Use a raw 5-byte secret so base32 encoding produces 8 chars without trailing '='.
	// Instead, use base32 encode of some bytes that gives non-mult-of-8 when padding stripped.
	rawBytes := []byte("hello") // 5 bytes → base32 is "NBSWY3DPEB3W64TMMQ======" → strip = → "NBSWY3DPEB3W64TMMQ" (18 chars, 18%8=2, needs 6 padding)
	// Actually use raw bytes that produce a clean unpadded length not divisible by 8.
	// 5 bytes → base32 (8 chars before padding, actually ceil(5*8/5)=8) — let's compute:
	// 5 bytes = 40 bits / 5 = 8 base32 chars. That IS divisible by 8.
	// 4 bytes = 32 bits → ceil(32/5) = 7 base32 chars → needs 1 '='
	// Strip padding from 4 bytes:
	rawBytes = []byte("test") // 4 bytes → base32 "ORSXG5A=" → without padding "ORSXG5A" (7 chars, 7%8=7, needs 1 padding)
	secretNoPad := strings.TrimRight(base32.StdEncoding.EncodeToString(rawBytes), "=")

	// Verify that the secret length is not divisible by 8.
	if len(secretNoPad)%8 == 0 {
		t.Skipf("test setup error: %q has length %d divisible by 8", secretNoPad, len(secretNoPad))
	}

	// Generate the current code using the padded secret.
	paddedSecret := secretNoPad
	if len(paddedSecret)%8 != 0 {
		paddedSecret += strings.Repeat("=", 8-len(paddedSecret)%8)
	}
	decoded, err := base32.StdEncoding.DecodeString(paddedSecret)
	if err != nil {
		t.Fatalf("base32 decode: %v", err)
	}
	counter := time.Now().Unix() / TOTPPeriod
	code := a.generateTOTP(decoded, counter)

	// verifyTOTP must add the padding itself and accept the valid code.
	if !a.verifyTOTP(secretNoPad, code) {
		t.Error("verifyTOTP should return true for valid code with unpadded secret")
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

// TestEnable2FA_DBExecError exercises the db.Exec error path in Enable2FA.
// We register a user and get a valid TOTP code, then replace the users table
// with a trigger that raises an error on UPDATE, so db.Exec fails after
// GetUserByID and verifyTOTP both succeed.
func TestEnable2FA_DBExecError(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "enable2fadberr2", "enable2fadberr2@example.com", "ValidPass1")

	setup, err := a.Generate2FASecret(user)
	if err != nil {
		t.Fatalf("Generate2FASecret: %v", err)
	}
	paddedSecret := setup.Secret
	if len(paddedSecret)%8 != 0 {
		paddedSecret += strings.Repeat("=", 8-len(paddedSecret)%8)
	}
	decoded, err := base32.StdEncoding.DecodeString(paddedSecret)
	if err != nil {
		t.Fatalf("base32 decode: %v", err)
	}
	code := a.generateTOTP(decoded, time.Now().Unix()/TOTPPeriod)

	// Install a BEFORE UPDATE trigger that raises an error, so the UPDATE in
	// Enable2FA fails while SELECT (GetUserByID) still works.
	if _, err := a.db.Exec(`
		CREATE TRIGGER block_2fa_update BEFORE UPDATE ON users
		BEGIN SELECT RAISE(FAIL, 'blocked by test trigger'); END
	`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}
	defer a.db.Exec("DROP TRIGGER IF EXISTS block_2fa_update")

	_, err = a.Enable2FA(user.ID, setup.Secret, code)
	if err == nil {
		t.Error("Enable2FA with blocked UPDATE trigger should return error")
	}
}

// TestDisable2FA_DBError exercises the db.Exec error path in Disable2FA.
func TestDisable2FA_DBError(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "disable2fadberr", "disable2fadberr@example.com", "ValidPass1")

	// Close the DB before the UPDATE executes.
	a.db.Close()

	err := a.Disable2FA(user.ID)
	if err == nil {
		t.Error("Disable2FA with closed DB should return error")
	}
}

// TestDisable2FA_PostgresQueryBranch exercises the postgres query branch.
// We set db.Driver to "postgres" so the $1-style query is used; modernc SQLite
// accepts both placeholder styles so the operation succeeds.
func TestDisable2FA_PostgresQueryBranch(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "disable2fapg", "disable2fapg@example.com", "ValidPass1")

	// Enable 2FA first so there is something to disable.
	if _, err := a.db.Exec("UPDATE users SET two_factor_enabled = 1, two_factor_secret = 'SECRET' WHERE id = ?", user.ID); err != nil {
		t.Fatalf("enabling 2FA: %v", err)
	}

	// Fake the driver so the postgres branch is taken.
	origDriver := a.db.Driver
	a.db.Driver = "postgres"
	defer func() { a.db.Driver = origDriver }()

	if err := a.Disable2FA(user.ID); err != nil {
		t.Errorf("Disable2FA with postgres query branch should succeed on SQLite: %v", err)
	}
}

// TestEnable2FA_PostgresQueryBranch exercises the postgres query branch in Enable2FA.
func TestEnable2FA_PostgresQueryBranch(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "enable2fapg", "enable2fapg@example.com", "ValidPass1")

	setup, err := a.Generate2FASecret(user)
	if err != nil {
		t.Fatalf("Generate2FASecret: %v", err)
	}
	paddedSecret := setup.Secret
	if len(paddedSecret)%8 != 0 {
		paddedSecret += strings.Repeat("=", 8-len(paddedSecret)%8)
	}
	decoded, err := base32.StdEncoding.DecodeString(paddedSecret)
	if err != nil {
		t.Fatalf("base32 decode: %v", err)
	}
	code := a.generateTOTP(decoded, time.Now().Unix()/TOTPPeriod)

	// Fake the driver so the postgres branch is taken.
	origDriver := a.db.Driver
	a.db.Driver = "postgres"
	defer func() { a.db.Driver = origDriver }()

	// modernc SQLite accepts $1 placeholders, so this should succeed.
	if _, err = a.Enable2FA(user.ID, setup.Secret, code); err != nil {
		t.Errorf("Enable2FA with postgres query branch should succeed on SQLite: %v", err)
	}
}
