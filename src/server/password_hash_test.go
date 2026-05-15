package server

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("correcthorsebatterystaple")
	if err != nil {
		t.Fatalf("HashPassword returned unexpected error: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("HashPassword returned %q, want prefix $argon2id$", hash)
	}
}

func TestVerifyPasswordCorrect(t *testing.T) {
	const password = "correcthorsebatterystaple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned unexpected error: %v", err)
	}

	if !VerifyPassword(password, hash) {
		t.Errorf("VerifyPassword(%q, hash) = false, want true", password)
	}
}

func TestVerifyPasswordWrong(t *testing.T) {
	hash, err := HashPassword("correcthorsebatterystaple")
	if err != nil {
		t.Fatalf("HashPassword returned unexpected error: %v", err)
	}

	if VerifyPassword("wrongpassword", hash) {
		t.Errorf("VerifyPassword(\"wrongpassword\", hash) = true, want false")
	}
}

func TestVerifyPasswordInvalidFormat(t *testing.T) {
	// Must not panic and must return false for a malformed hash
	result := VerifyPassword("password", "invalid-hash-format")
	if result {
		t.Errorf("VerifyPassword(\"password\", \"invalid-hash-format\") = true, want false")
	}
}

func TestHashPasswordRandomSalt(t *testing.T) {
	const password = "correcthorsebatterystaple"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("first HashPassword returned unexpected error: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("second HashPassword returned unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("HashPassword returned identical hashes for two separate calls; random salt is not being applied")
	}
}

func TestVerifyPasswordEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{"empty hash", "password", "", false},
		{"wrong prefix", "password", "$bcrypt$something", false},
		{"only prefix", "password", "$argon2id$", false},
		{"six-part but garbage", "password", "$argon2id$v=19$m=65536,t=3,p=4$notbase64!!!$notbase64!!!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyPassword(tt.password, tt.hash)
			if got != tt.want {
				t.Errorf("VerifyPassword(%q, %q) = %v, want %v", tt.password, tt.hash, got, tt.want)
			}
		})
	}
}
