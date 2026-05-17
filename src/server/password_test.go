package server

import (
	"strings"
	"testing"
)

// TestRequestPasswordReset_UnknownEmail verifies that an unknown email returns no error
// and no token (security: don't reveal whether email exists).
func TestRequestPasswordReset_UnknownEmail(t *testing.T) {
	a := newTestAuth(t)
	token, err := a.RequestPasswordReset("nobody@example.com")
	if err != nil {
		t.Fatalf("RequestPasswordReset unknown email: %v", err)
	}
	if token != "" {
		t.Error("token should be empty for unknown email")
	}
}

func TestRequestPasswordReset_KnownEmail(t *testing.T) {
	a := newTestAuth(t)
	registerTestUser(t, a, "alice", "alice@example.com", "ValidPass1!")

	token, err := a.RequestPasswordReset("alice@example.com")
	if err != nil {
		t.Fatalf("RequestPasswordReset: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token for known email")
	}
}

func TestValidatePasswordResetToken_Invalid(t *testing.T) {
	a := newTestAuth(t)
	_, err := a.ValidatePasswordResetToken("nonexistent-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
	if err != ErrInvalidToken {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestValidatePasswordResetToken_Valid(t *testing.T) {
	a := newTestAuth(t)
	registerTestUser(t, a, "bob", "bob@example.com", "ValidPass1!")

	token, err := a.RequestPasswordReset("bob@example.com")
	if err != nil || token == "" {
		t.Fatalf("RequestPasswordReset: err=%v token=%q", err, token)
	}

	userID, err := a.ValidatePasswordResetToken(token)
	if err != nil {
		t.Fatalf("ValidatePasswordResetToken: %v", err)
	}
	if userID == "" {
		t.Error("expected non-empty userID")
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	a := newTestAuth(t)
	err := a.ResetPassword("bad-token", "NewValidPass1!")
	if err == nil {
		t.Error("ResetPassword with invalid token should error")
	}
}

func TestResetPassword_Success(t *testing.T) {
	a := newTestAuth(t)
	registerTestUser(t, a, "carol", "carol@example.com", "OldPass1!")

	token, err := a.RequestPasswordReset("carol@example.com")
	if err != nil || token == "" {
		t.Fatalf("RequestPasswordReset: %v / %q", err, token)
	}

	if err := a.ResetPassword(token, "NewPass1!"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	// Token should now be invalid (used)
	if err := a.ResetPassword(token, "AnotherPass1!"); err == nil {
		t.Error("token should be invalidated after use")
	}
}

func TestChangePassword_Success(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "dave", "dave@example.com", "OldPass1!")

	if err := a.ChangePassword(user.ID, "OldPass1!", "NewPass2!"); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "eve", "eve@example.com", "OldPass1!")

	err := a.ChangePassword(user.ID, "WrongPass99!", "NewPass2!")
	if err == nil {
		t.Error("ChangePassword with wrong current password should error")
	}
	if err != ErrInvalidCredentials {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestChangePassword_SamePassword(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "frank", "frank@example.com", "SamePass1!")

	err := a.ChangePassword(user.ID, "SamePass1!", "SamePass1!")
	if err == nil {
		t.Error("ChangePassword with same password should error")
	}
}

func TestGenerateEmailVerificationToken(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "grace", "grace@example.com", "ValidPass1!")

	token, err := a.GenerateEmailVerificationToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateEmailVerificationToken: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestGenerateEmailVerificationToken_UnknownUser(t *testing.T) {
	a := newTestAuth(t)
	_, err := a.GenerateEmailVerificationToken("nonexistent-user-id")
	if err == nil {
		t.Error("expected error for unknown user ID")
	}
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	a := newTestAuth(t)
	err := a.VerifyEmail("invalid-token")
	if err == nil {
		t.Error("VerifyEmail with invalid token should error")
	}
	if err != ErrInvalidVerificationToken {
		t.Errorf("err = %v, want ErrInvalidVerificationToken", err)
	}
}

func TestVerifyEmail_Success(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "hank", "hank@example.com", "ValidPass1!")

	token, err := a.GenerateEmailVerificationToken(user.ID)
	if err != nil {
		t.Fatalf("GenerateEmailVerificationToken: %v", err)
	}

	if err := a.VerifyEmail(token); err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}
}

func TestResendVerificationEmail_UnknownEmail(t *testing.T) {
	a := newTestAuth(t)
	token, err := a.ResendVerificationEmail("nobody@example.com")
	if err != nil {
		t.Fatalf("ResendVerificationEmail unknown: %v", err)
	}
	if token != "" {
		t.Error("token should be empty for unknown email")
	}
}

func TestResendVerificationEmail_AlreadyVerified(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "iris", "iris@example.com", "ValidPass1!")

	// Verify first
	tok, _ := a.GenerateEmailVerificationToken(user.ID)
	a.VerifyEmail(tok)

	_, err := a.ResendVerificationEmail("iris@example.com")
	if err == nil {
		t.Error("ResendVerificationEmail should error for already-verified email")
	}
}

func TestInvalidateAllPasswordResetTokens(t *testing.T) {
	a := newTestAuth(t)
	user := registerTestUser(t, a, "jack", "jack@example.com", "ValidPass1!")

	// Create a token first
	token, _ := a.RequestPasswordReset("jack@example.com")
	if token == "" {
		t.Skip("no token generated, skipping invalidation test")
	}

	if err := a.InvalidateAllPasswordResetTokens(user.ID); err != nil {
		t.Fatalf("InvalidateAllPasswordResetTokens: %v", err)
	}

	// Token should now be invalid
	_, err := a.ValidatePasswordResetToken(token)
	if err == nil {
		t.Error("token should be invalid after invalidation")
	}
}

func TestCheckPasswordStrength_VeryWeak(t *testing.T) {
	a := newTestAuth(t)
	result := a.CheckPasswordStrength("abc")

	if result["score"].(int) > 1 {
		t.Errorf("score = %v, want <= 1 for weak password", result["score"])
	}
	if result["label"].(string) != "Very Weak" {
		t.Errorf("label = %q, want Very Weak", result["label"])
	}
}

func TestCheckPasswordStrength_Strong(t *testing.T) {
	a := newTestAuth(t)
	result := a.CheckPasswordStrength("MyStr0ng!Pass#Word")

	if result["score"].(int) < 4 {
		t.Errorf("score = %v, want >= 4 for strong password", result["score"])
	}
}

func TestCheckPasswordStrength_AllFields(t *testing.T) {
	a := newTestAuth(t)
	result := a.CheckPasswordStrength("Test1!")

	required := []string{"score", "length", "has_upper", "has_lower", "has_number", "has_special", "feedback", "label"}
	for _, key := range required {
		if _, ok := result[key]; !ok {
			t.Errorf("CheckPasswordStrength missing key %q", key)
		}
	}
}

func TestCheckPasswordStrength_Feedback(t *testing.T) {
	a := newTestAuth(t)
	// Short, no upper, no number, no special
	result := a.CheckPasswordStrength("abc")
	feedback := result["feedback"].([]string)
	if len(feedback) == 0 {
		t.Error("expected feedback for weak password")
	}
}

func TestCheckPasswordStrength_MaxScore(t *testing.T) {
	a := newTestAuth(t)
	// Very long with all character types — score should cap at 5
	result := a.CheckPasswordStrength("ThisIsAVeryLongPassword123!@#$%^&*")
	if result["score"].(int) > 5 {
		t.Errorf("score = %v, want <= 5 (capped)", result["score"])
	}
}

func TestForcePasswordChange(t *testing.T) {
	a := newTestAuth(t)
	err := a.ForcePasswordChange("any-user-id")
	if err == nil {
		t.Error("ForcePasswordChange should return error (not implemented)")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("error = %q, want 'not implemented' message", err.Error())
	}
}

func TestHashToken(t *testing.T) {
	h1 := HashToken("my-api-token")
	h2 := HashToken("my-api-token")
	if h1 != h2 {
		t.Error("HashToken should be deterministic")
	}
	if h1 == "my-api-token" {
		t.Error("HashToken should not return the plaintext")
	}
	if len(h1) != 64 {
		t.Errorf("HashToken length = %d, want 64 hex chars (SHA-256)", len(h1))
	}
}

func TestHashToken_Different(t *testing.T) {
	h1 := HashToken("token-a")
	h2 := HashToken("token-b")
	if h1 == h2 {
		t.Error("different tokens should have different hashes")
	}
}

func TestGenerateRandomToken_Length(t *testing.T) {
	token, err := GenerateRandomToken(16)
	if err != nil {
		t.Fatalf("GenerateRandomToken: %v", err)
	}
	// 16 random bytes = 32 hex chars
	if len(token) != 32 {
		t.Errorf("token length = %d, want 32 (16 bytes as hex)", len(token))
	}
}

func TestGenerateRandomToken_Unique(t *testing.T) {
	t1, _ := GenerateRandomToken(16)
	t2, _ := GenerateRandomToken(16)
	if t1 == t2 {
		t.Error("GenerateRandomToken should produce unique tokens")
	}
}
