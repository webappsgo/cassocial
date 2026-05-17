package server

import (
	"encoding/base32"
	"strings"
	"testing"
	"time"

	"github.com/casapps/cassocial/src/server/model"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestAuth creates an Auth backed by an in-memory SQLite database.
func newTestAuth(t *testing.T) *Auth {
	t.Helper()
	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}
	return NewAuth(db, "test-jwt-secret-12345678")
}

// registerTestUser is a helper that registers a user and fails on error.
func registerTestUser(t *testing.T, a *Auth, username, email, password string) *model.User {
	t.Helper()
	user, err := a.Register(username, email, password)
	if err != nil {
		t.Fatalf("Register(%q, %q): %v", username, email, err)
	}
	return user
}

// TestNewAuth_WithSecret verifies that NewAuth stores the provided secret.
func TestNewAuth_WithSecret(t *testing.T) {
	a := newTestAuth(t)
	if a == nil {
		t.Fatal("NewAuth returned nil")
	}
	if len(a.jwtSecret) == 0 {
		t.Error("jwtSecret is empty")
	}
}

// TestNewAuth_EmptySecret verifies that NewAuth generates a secret when none is provided.
func TestNewAuth_EmptySecret(t *testing.T) {
	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	defer db.Close()
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}
	a := NewAuth(db, "")
	if len(a.jwtSecret) == 0 {
		t.Error("NewAuth with empty secret should have generated a random secret")
	}
}

// TestRegister_Success verifies that a valid registration returns a user with the expected fields.
func TestRegister_Success(t *testing.T) {
	a := newTestAuth(t)
	user, err := a.Register("testuser", "test@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if user.ID == "" {
		t.Error("Register returned user with empty ID")
	}
	if user.Username != "testuser" {
		t.Errorf("Username = %q, want %q", user.Username, "testuser")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "test@example.com")
	}
}

// TestRegister_DuplicateUsername verifies that duplicate usernames are rejected.
func TestRegister_DuplicateUsername(t *testing.T) {
	a := newTestAuth(t)
	registerTestUser(t, a, "dupuser", "first@example.com", "ValidPass1")

	_, err := a.Register("dupuser", "second@example.com", "ValidPass1")
	if err != ErrUsernameExists {
		t.Errorf("Register duplicate username: got %v, want ErrUsernameExists", err)
	}
}

// TestRegister_DuplicateEmail verifies that duplicate emails are rejected.
func TestRegister_DuplicateEmail(t *testing.T) {
	a := newTestAuth(t)
	registerTestUser(t, a, "firstuser", "same@example.com", "ValidPass1")

	_, err := a.Register("seconduser", "same@example.com", "ValidPass1")
	if err != ErrEmailExists {
		t.Errorf("Register duplicate email: got %v, want ErrEmailExists", err)
	}
}

// TestRegister_WeakPassword verifies that a weak password is rejected.
func TestRegister_WeakPassword(t *testing.T) {
	a := newTestAuth(t)
	_, err := a.Register("weakpwuser", "weakpw@example.com", "short")
	if err == nil {
		t.Error("Register with weak password should return an error")
	}
}

// TestLogin_Success verifies successful login returns a token.
func TestLogin_Success(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "loginuser", "login@example.com", "ValidPass1")

	// Disable email verification requirement so login can succeed.
	if _, err := a.db.Exec("UPDATE settings SET value = 'false' WHERE key = 'email_verification_required'"); err != nil {
		t.Fatalf("disabling email verification: %v", err)
	}
	// Mark user as email-verified.
	if _, err := a.db.Exec("UPDATE users SET email_verified = 1 WHERE id = ?", user.ID); err != nil {
		t.Fatalf("marking email verified: %v", err)
	}

	token, gotUser, err := a.Login("loginuser", "ValidPass1")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if token == "" {
		t.Error("Login returned empty token")
	}
	if gotUser == nil {
		t.Error("Login returned nil user")
	}
}

// TestLogin_WrongPassword verifies that a wrong password returns ErrInvalidCredentials.
func TestLogin_WrongPassword(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "pwuser", "pw@example.com", "ValidPass1")
	// Ensure email is verified so the wrong-password path is reached.
	if _, err := a.db.Exec("UPDATE users SET email_verified = 1 WHERE id = ?", user.ID); err != nil {
		t.Fatalf("marking email verified: %v", err)
	}
	if _, err := a.db.Exec("UPDATE settings SET value = 'false' WHERE key = 'email_verification_required'"); err != nil {
		t.Fatalf("disabling email verification: %v", err)
	}

	_, _, err := a.Login("pwuser", "WrongPass1")
	if err != ErrInvalidCredentials {
		t.Errorf("Login with wrong password: got %v, want ErrInvalidCredentials", err)
	}
}

// TestLogin_UnknownUser verifies that login with unknown user returns ErrInvalidCredentials.
func TestLogin_UnknownUser(t *testing.T) {
	a := newTestAuth(t)
	_, _, err := a.Login("nobody", "ValidPass1")
	if err != ErrInvalidCredentials {
		t.Errorf("Login with unknown user: got %v, want ErrInvalidCredentials", err)
	}
}

// TestLogin_2FARequired verifies that 2FA-enabled accounts return Err2FARequired.
func TestLogin_2FARequired(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "2fauser", "2fa@example.com", "ValidPass1")

	// Disable email verification so login can progress to 2FA check.
	if _, err := a.db.Exec("UPDATE settings SET value = 'false' WHERE key = 'email_verification_required'"); err != nil {
		t.Fatalf("disabling email verification: %v", err)
	}
	if _, err := a.db.Exec("UPDATE users SET email_verified = 1 WHERE id = ?", user.ID); err != nil {
		t.Fatalf("marking email verified: %v", err)
	}

	// Manually enable 2FA with a dummy secret so login detects it.
	if _, err := a.db.Exec("UPDATE users SET two_factor_enabled = 1, two_factor_secret = 'JBSWY3DPEHPK3PXP' WHERE id = ?", user.ID); err != nil {
		t.Fatalf("enabling 2FA in DB: %v", err)
	}

	_, _, err := a.Login("2fauser", "ValidPass1")
	if err != Err2FARequired {
		t.Errorf("Login with 2FA enabled: got %v, want Err2FARequired", err)
	}
}

// TestLoginWith2FA_InvalidCode verifies that a bad 2FA code returns ErrInvalidCredentials.
func TestLoginWith2FA_InvalidCode(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "2fauser2", "2fa2@example.com", "ValidPass1")

	query := "UPDATE users SET two_factor_enabled = 1, two_factor_secret = 'JBSWY3DPEHPK3PXP' WHERE id = ?"
	if _, err := a.db.Exec(query, user.ID); err != nil {
		t.Fatalf("enabling 2FA: %v", err)
	}

	_, _, err := a.LoginWith2FA(user.ID, "000000")
	if err != ErrInvalidCredentials {
		t.Errorf("LoginWith2FA bad code: got %v, want ErrInvalidCredentials", err)
	}
}

// TestLoginWith2FA_NotFoundUser verifies that an unknown userID returns an error.
func TestLoginWith2FA_NotFoundUser(t *testing.T) {
	a := newTestAuth(t)
	_, _, err := a.LoginWith2FA("nonexistent-id", "123456")
	if err == nil {
		t.Error("LoginWith2FA with nonexistent user should return error")
	}
}

// TestGenerateToken_ValidUser verifies that a token is generated correctly.
func TestGenerateToken_ValidUser(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "tokenuser", "token@example.com", "ValidPass1")

	token, err := a.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Error("GenerateToken returned empty token")
	}
}

// TestValidateToken_Valid verifies that a valid token is parsed correctly.
func TestValidateToken_Valid(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "valuser", "val@example.com", "ValidPass1")

	token, err := a.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := a.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("claims.UserID = %q, want %q", claims.UserID, user.ID)
	}
	if claims.Username != user.Username {
		t.Errorf("claims.Username = %q, want %q", claims.Username, user.Username)
	}
}

// TestValidateToken_InvalidString verifies that a garbage token returns ErrInvalidToken.
func TestValidateToken_InvalidString(t *testing.T) {
	a := newTestAuth(t)
	_, err := a.ValidateToken("this.is.not.a.valid.jwt")
	if err != ErrInvalidToken {
		t.Errorf("ValidateToken invalid: got %v, want ErrInvalidToken", err)
	}
}

// TestValidateToken_WrongSecret verifies that a token signed with a different secret fails.
func TestValidateToken_WrongSecret(t *testing.T) {
	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	defer db.Close()
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	a1 := NewAuth(db, "secret-one")
	a2 := NewAuth(db, "secret-two")

	user, err := a1.Register("wsuser", "ws@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	token, err := a1.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	_, err = a2.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Errorf("ValidateToken wrong secret: got %v, want ErrInvalidToken", err)
	}
}

// TestRefreshToken_Success verifies that a valid token can be refreshed.
func TestRefreshToken_Success(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "refreshuser", "refresh@example.com", "ValidPass1")

	token, err := a.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	newToken, err := a.RefreshToken(token)
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}
	if newToken == "" {
		t.Error("RefreshToken returned empty token")
	}
}

// TestRefreshToken_InvalidToken verifies that an invalid token cannot be refreshed.
func TestRefreshToken_InvalidToken(t *testing.T) {
	a := newTestAuth(t)
	_, err := a.RefreshToken("not-a-token")
	if err != ErrInvalidToken {
		t.Errorf("RefreshToken invalid: got %v, want ErrInvalidToken", err)
	}
}

// TestValidatePassword_Valid verifies that a valid password passes.
func TestValidatePassword_Valid(t *testing.T) {
	a := newTestAuth(t)
	if err := a.ValidatePassword("ValidPass1"); err != nil {
		t.Errorf("ValidatePassword valid password: %v", err)
	}
}

// TestValidatePassword_TooShort verifies that a too-short password is rejected.
func TestValidatePassword_TooShort(t *testing.T) {
	a := newTestAuth(t)
	err := a.ValidatePassword("Abc1")
	if err == nil {
		t.Error("ValidatePassword short password should return error")
	}
}

// TestValidatePassword_NoUppercase verifies that a password without uppercase fails when required.
// The default migration already sets password_require_uppercase=true, so this checks the default.
func TestValidatePassword_NoUppercase(t *testing.T) {
	a := newTestAuth(t)
	// Default: password_require_uppercase=true from migration.
	err := a.ValidatePassword("nouppercase1")
	if err == nil {
		t.Error("ValidatePassword without uppercase should fail when required")
	}
}

// TestValidatePassword_NoNumber verifies that a password without a digit fails when required.
// The default migration already sets password_require_number=true.
func TestValidatePassword_NoNumber(t *testing.T) {
	a := newTestAuth(t)
	// Turn off uppercase requirement via direct SQL so only number check applies.
	if _, err := a.db.Exec("UPDATE settings SET value = 'false' WHERE key = 'password_require_uppercase'"); err != nil {
		t.Fatalf("UPDATE settings: %v", err)
	}
	err := a.ValidatePassword("NoNumberHere")
	if err == nil {
		t.Error("ValidatePassword without number should fail when required")
	}
}

// TestValidatePassword_NoSpecial verifies that a password without a special char fails when required.
func TestValidatePassword_NoSpecial(t *testing.T) {
	a := newTestAuth(t)
	// Enable special, disable others via direct SQL.
	if _, err := a.db.Exec("UPDATE settings SET value = 'true' WHERE key = 'password_require_special'"); err != nil {
		t.Fatalf("UPDATE settings: %v", err)
	}
	if _, err := a.db.Exec("UPDATE settings SET value = 'false' WHERE key = 'password_require_uppercase'"); err != nil {
		t.Fatalf("UPDATE settings: %v", err)
	}
	if _, err := a.db.Exec("UPDATE settings SET value = 'false' WHERE key = 'password_require_number'"); err != nil {
		t.Fatalf("UPDATE settings: %v", err)
	}
	err := a.ValidatePassword("nospecialcharhere")
	if err == nil {
		t.Error("ValidatePassword without special char should fail when required")
	}
}

// TestGetPasswordRequirements_Defaults verifies that default requirements are returned correctly.
func TestGetPasswordRequirements_Defaults(t *testing.T) {
	a := newTestAuth(t)
	reqs, err := a.GetPasswordRequirements()
	if err != nil {
		t.Fatalf("GetPasswordRequirements: %v", err)
	}
	if reqs.MinLength < 1 {
		t.Errorf("MinLength = %d, want >= 1", reqs.MinLength)
	}
}

// TestGetPasswordRequirements_FromDB verifies that settings from DB are respected.
func TestGetPasswordRequirements_FromDB(t *testing.T) {
	a := newTestAuth(t)
	// Use direct SQL to set the value since SetSetting has a known SQLite bug.
	if _, err := a.db.Exec("UPDATE settings SET value = '12' WHERE key = 'password_min_length'"); err != nil {
		t.Fatalf("UPDATE settings: %v", err)
	}
	reqs, err := a.GetPasswordRequirements()
	if err != nil {
		t.Fatalf("GetPasswordRequirements: %v", err)
	}
	if reqs.MinLength != 12 {
		t.Errorf("MinLength = %d, want 12", reqs.MinLength)
	}
}

// TestGetUserByID_Found verifies that a registered user can be retrieved by ID.
func TestGetUserByID_Found(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "iduser", "id@example.com", "ValidPass1")

	got, err := a.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("GetUserByID ID = %q, want %q", got.ID, user.ID)
	}
}

// TestGetUserByID_NotFound verifies that a missing ID returns ErrUserNotFound.
func TestGetUserByID_NotFound(t *testing.T) {
	a := newTestAuth(t)
	_, err := a.GetUserByID("does-not-exist")
	if err != ErrUserNotFound {
		t.Errorf("GetUserByID missing: got %v, want ErrUserNotFound", err)
	}
}

// TestGetUserByUsername_Found verifies retrieval by username.
func TestGetUserByUsername_Found(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "unameuser", "uname@example.com", "ValidPass1")

	got, err := a.GetUserByUsername("unameuser")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("GetUserByUsername ID = %q, want %q", got.ID, user.ID)
	}
}

// TestGetUserByUsername_NotFound verifies missing username returns ErrUserNotFound.
func TestGetUserByUsername_NotFound(t *testing.T) {
	a := newTestAuth(t)
	_, err := a.GetUserByUsername("nobody")
	if err != ErrUserNotFound {
		t.Errorf("GetUserByUsername missing: got %v, want ErrUserNotFound", err)
	}
}

// TestGetUserByEmail_Found verifies retrieval by email.
func TestGetUserByEmail_Found(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "emailuser", "emailfind@example.com", "ValidPass1")

	got, err := a.GetUserByEmail("emailfind@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("GetUserByEmail ID = %q, want %q", got.ID, user.ID)
	}
}

// TestGetUserByEmail_NotFound verifies missing email returns ErrUserNotFound.
func TestGetUserByEmail_NotFound(t *testing.T) {
	a := newTestAuth(t)
	_, err := a.GetUserByEmail("nobody@example.com")
	if err != ErrUserNotFound {
		t.Errorf("GetUserByEmail missing: got %v, want ErrUserNotFound", err)
	}
}

// TestGetUserByUsernameOrEmail_ByUsername verifies lookup by username field.
func TestGetUserByUsernameOrEmail_ByUsername(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "oreuser", "ore@example.com", "ValidPass1")

	got, err := a.GetUserByUsernameOrEmail("oreuser")
	if err != nil {
		t.Fatalf("GetUserByUsernameOrEmail by username: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("ID = %q, want %q", got.ID, user.ID)
	}
}

// TestGetUserByUsernameOrEmail_ByEmail verifies lookup by email field.
func TestGetUserByUsernameOrEmail_ByEmail(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "oreuser2", "ore2@example.com", "ValidPass1")

	got, err := a.GetUserByUsernameOrEmail("ore2@example.com")
	if err != nil {
		t.Fatalf("GetUserByUsernameOrEmail by email: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("ID = %q, want %q", got.ID, user.ID)
	}
}

// TestGetUserByUsernameOrEmail_NotFound verifies missing entry returns ErrUserNotFound.
func TestGetUserByUsernameOrEmail_NotFound(t *testing.T) {
	a := newTestAuth(t)
	_, err := a.GetUserByUsernameOrEmail("nobody-at-all")
	if err != ErrUserNotFound {
		t.Errorf("GetUserByUsernameOrEmail missing: got %v, want ErrUserNotFound", err)
	}
}

// TestUsernameExists verifies that usernameExists returns true for an existing username.
func TestUsernameExists(t *testing.T) {
	a := newTestAuth(t)
	registerTestUser(t, a, "existuser", "exist@example.com", "ValidPass1")

	exists, err := a.usernameExists("existuser")
	if err != nil {
		t.Fatalf("usernameExists: %v", err)
	}
	if !exists {
		t.Error("usernameExists should return true for registered username")
	}

	exists, err = a.usernameExists("notregistered")
	if err != nil {
		t.Fatalf("usernameExists: %v", err)
	}
	if exists {
		t.Error("usernameExists should return false for unknown username")
	}
}

// TestEmailExists verifies that emailExists returns true for an existing email.
func TestEmailExists(t *testing.T) {
	a := newTestAuth(t)
	registerTestUser(t, a, "existemail", "existemail@example.com", "ValidPass1")

	exists, err := a.emailExists("existemail@example.com")
	if err != nil {
		t.Fatalf("emailExists: %v", err)
	}
	if !exists {
		t.Error("emailExists should return true for registered email")
	}

	exists, err = a.emailExists("notregistered@example.com")
	if err != nil {
		t.Fatalf("emailExists: %v", err)
	}
	if exists {
		t.Error("emailExists should return false for unknown email")
	}
}

// TestUpdateLastLogin verifies that updateLastLogin executes without error.
func TestUpdateLastLogin(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "lastloginuser", "lastlogin@example.com", "ValidPass1")

	if err := a.updateLastLogin(user.ID); err != nil {
		t.Errorf("updateLastLogin: %v", err)
	}
}

// TestLogin_InactiveUser verifies that a pending/disabled user cannot log in.
func TestLogin_InactiveUser(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "inactiveuser", "inactive@example.com", "ValidPass1")

	// Make user inactive (pending status).
	if _, err := a.db.Exec("UPDATE users SET status = ? WHERE id = ?", model.StatusPending, user.ID); err != nil {
		t.Fatalf("setting status pending: %v", err)
	}
	if _, err := a.db.Exec("UPDATE settings SET value = 'false' WHERE key = 'email_verification_required'"); err != nil {
		t.Fatalf("disabling email verification: %v", err)
	}

	_, _, err := a.Login("inactiveuser", "ValidPass1")
	if err != ErrUserNotActive {
		t.Errorf("Login inactive user: got %v, want ErrUserNotActive", err)
	}
}

// TestLogin_EmailNotVerified verifies that unverified email blocks login when required.
func TestLogin_EmailNotVerified(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "unverifuser", "unverif@example.com", "ValidPass1")

	// Require email verification and ensure user is not verified.
	if _, err := a.db.Exec("UPDATE settings SET value = 'true' WHERE key = 'email_verification_required'"); err != nil {
		t.Fatalf("enabling email verification: %v", err)
	}
	if _, err := a.db.Exec("UPDATE users SET email_verified = 0 WHERE id = ?", user.ID); err != nil {
		t.Fatalf("marking email unverified: %v", err)
	}

	_, _, err := a.Login("unverifuser", "ValidPass1")
	if err != ErrEmailNotVerified {
		t.Errorf("Login unverified email: got %v, want ErrEmailNotVerified", err)
	}
}

// TestLogin_ByEmail verifies that a user can log in using their email address.
func TestLogin_ByEmail(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "emailloginuser", "emaillogin@example.com", "ValidPass1")
	if _, err := a.db.Exec("UPDATE settings SET value = 'false' WHERE key = 'email_verification_required'"); err != nil {
		t.Fatalf("disabling email verification: %v", err)
	}
	if _, err := a.db.Exec("UPDATE users SET email_verified = 1 WHERE id = ?", user.ID); err != nil {
		t.Fatalf("marking email verified: %v", err)
	}

	token, gotUser, err := a.Login("emaillogin@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Login by email: %v", err)
	}
	if token == "" {
		t.Error("Login by email returned empty token")
	}
	if gotUser.ID != user.ID {
		t.Errorf("Login by email: user.ID = %q, want %q", gotUser.ID, user.ID)
	}
}

// TestLoginWith2FA_ValidCode verifies that a correct TOTP code completes login.
func TestLoginWith2FA_ValidCode(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "valid2fauser", "valid2fa@example.com", "ValidPass1")

	const rawSecret = "JBSWY3DPEHPK3PXP"
	if _, err := a.db.Exec("UPDATE users SET two_factor_enabled = 1, two_factor_secret = ? WHERE id = ?", rawSecret, user.ID); err != nil {
		t.Fatalf("enabling 2FA: %v", err)
	}

	// Derive a current TOTP code using the same algorithm as the server.
	paddedSecret := rawSecret
	if len(paddedSecret)%8 != 0 {
		paddedSecret += strings.Repeat("=", 8-len(paddedSecret)%8)
	}
	decoded, err := base32.StdEncoding.DecodeString(paddedSecret)
	if err != nil {
		t.Fatalf("base32 decode: %v", err)
	}
	counter := time.Now().Unix() / TOTPPeriod
	code := a.generateTOTP(decoded, counter)

	token, gotUser, err := a.LoginWith2FA(user.ID, code)
	if err != nil {
		t.Fatalf("LoginWith2FA valid code: %v", err)
	}
	if token == "" {
		t.Error("LoginWith2FA returned empty token")
	}
	if gotUser.ID != user.ID {
		t.Errorf("LoginWith2FA user.ID = %q, want %q", gotUser.ID, user.ID)
	}
}

// TestRefreshToken_InactiveUser verifies that refresh fails when the user is no longer active.
func TestRefreshToken_InactiveUser(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "inactiverefresh", "inactiverefresh@example.com", "ValidPass1")

	token, err := a.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	// Deactivate user after token was issued (suspended is not active).
	if _, err := a.db.Exec("UPDATE users SET status = ? WHERE id = ?", model.StatusSuspended, user.ID); err != nil {
		t.Fatalf("suspending user: %v", err)
	}

	_, err = a.RefreshToken(token)
	if err != ErrUserNotActive {
		t.Errorf("RefreshToken inactive user: got %v, want ErrUserNotActive", err)
	}
}

// TestRefreshToken_UnknownUser verifies that refresh fails when user no longer exists.
func TestRefreshToken_UnknownUser(t *testing.T) {
	a := newTestAuth(t)

	// Build a token for a user that doesn't exist in DB.
	fakeUser := &model.User{
		ID:       "fake-id-not-in-db",
		Username: "ghostuser",
		Role:     model.RoleUser,
		Status:   model.StatusActive,
	}
	token, err := a.GenerateToken(fakeUser)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	_, err = a.RefreshToken(token)
	if err == nil {
		t.Error("RefreshToken for nonexistent user should return error")
	}
}

// TestRegister_PendingWhenApprovalRequired verifies that users get pending status
// when registration_requires_approval is true.
func TestRegister_PendingWhenApprovalRequired(t *testing.T) {
	a := newTestAuth(t)
	if _, err := a.db.Exec("UPDATE settings SET value = 'true' WHERE key = 'registration_requires_approval'"); err != nil {
		t.Fatalf("enabling approval: %v", err)
	}

	user, err := a.Register("pendinguser", "pending@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if user.Status != model.StatusPending {
		t.Errorf("Status = %q, want %q", user.Status, model.StatusPending)
	}
}

// TestRegister_EmailUnverifiedWhenVerificationRequired verifies that EmailVerified is false
// when email_verification_required is true.
func TestRegister_EmailUnverifiedWhenVerificationRequired(t *testing.T) {
	a := newTestAuth(t)
	if _, err := a.db.Exec("UPDATE settings SET value = 'true' WHERE key = 'email_verification_required'"); err != nil {
		t.Fatalf("enabling email verification: %v", err)
	}

	user, err := a.Register("verifuser", "verif@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if user.EmailVerified {
		t.Error("EmailVerified should be false when email_verification_required is true")
	}
}

// TestGenerateToken_CustomSessionTimeout verifies that a session_timeout_minutes setting
// is respected (token is generated without error and can be validated).
func TestGenerateToken_CustomSessionTimeout(t *testing.T) {
	a := newTestAuth(t)
	if _, err := a.db.Exec("UPDATE settings SET value = '60' WHERE key = 'session_timeout_minutes'"); err != nil {
		t.Fatalf("setting session timeout: %v", err)
	}

	user := registerTestUser(t, a, "timeoutuser", "timeout@example.com", "ValidPass1")
	token, err := a.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken with custom timeout: %v", err)
	}
	claims, err := a.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken after GenerateToken with custom timeout: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("claims.UserID = %q, want %q", claims.UserID, user.ID)
	}
}

// TestValidateToken_WrongAlgorithm verifies that a token signed with a non-HMAC algorithm
// is rejected.
func TestValidateToken_WrongAlgorithm(t *testing.T) {
	a := newTestAuth(t)
	// "none" algorithm tokens are a known attack vector; jwt library produces them
	// when parsed with an empty secret check. We test that our parser rejects garbage alg.
	_, err := a.ValidateToken("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VyX2lkIjoiYWJjIn0.")
	if err != ErrInvalidToken {
		t.Errorf("ValidateToken none-alg: got %v, want ErrInvalidToken", err)
	}
}

// TestGenerateRandomString verifies that generateRandomString produces hex strings of expected length.
func TestGenerateRandomString(t *testing.T) {
	s := generateRandomString(16)
	// hex-encoded 16 bytes = 32 chars
	if len(s) != 32 {
		t.Errorf("generateRandomString(16) len = %d, want 32", len(s))
	}

	s2 := generateRandomString(16)
	if s == s2 {
		t.Error("generateRandomString produced identical output on consecutive calls")
	}
}

// TestGenerateUUID verifies that generateUUID returns a UUID-shaped string.
func TestGenerateUUID(t *testing.T) {
	u := generateUUID()
	if u == "" {
		t.Error("generateUUID returned empty string")
	}
	// UUID format: 8-4-4-4-12 hex chars separated by dashes
	parts := strings.Split(u, "-")
	if len(parts) != 5 {
		t.Errorf("generateUUID = %q, want 5 dash-separated parts, got %d", u, len(parts))
	}

	u2 := generateUUID()
	if u == u2 {
		t.Error("generateUUID produced identical values on consecutive calls")
	}
}
