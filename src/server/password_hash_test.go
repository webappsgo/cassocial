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

func TestDecodeArgon2Hash_WrongParts(t *testing.T) {
	// Only 3 $ delimiters instead of 5 — must return error
	_, _, err := decodeArgon2Hash("$argon2id$garbage")
	if err == nil {
		t.Error("decodeArgon2Hash with wrong part count should return error")
	}
}

func TestDecodeArgon2Hash_WrongAlgorithm(t *testing.T) {
	// Correct number of parts but algorithm is not argon2id
	_, _, err := decodeArgon2Hash("$bcrypt$v=19$m=65536,t=3,p=4$salt$hash")
	if err == nil {
		t.Error("decodeArgon2Hash with non-argon2id algorithm should return error")
	}
}

func TestDecodeArgon2Hash_BadSaltEncoding(t *testing.T) {
	// Valid structure but salt has invalid base64
	_, _, err := decodeArgon2Hash("$argon2id$v=19$m=65536,t=3,p=4$!!!invalid!!!$AAAA")
	if err == nil {
		t.Error("decodeArgon2Hash with invalid salt encoding should return error")
	}
}

func TestDecodeArgon2Hash_BadHashEncoding(t *testing.T) {
	validSalt := "AAAAAAAAAAAAAAAAAAAAAA" // valid base64 salt (16 bytes raw std)
	// Valid salt but invalid hash
	_, _, err := decodeArgon2Hash("$argon2id$v=19$m=65536,t=3,p=4$" + validSalt + "$!!!invalid!!!")
	if err == nil {
		t.Error("decodeArgon2Hash with invalid hash encoding should return error")
	}
}

func TestGenerateRandomToken_ZeroLength(t *testing.T) {
	token, err := GenerateRandomToken(0)
	if err != nil {
		t.Fatalf("GenerateRandomToken(0): %v", err)
	}
	if len(token) != 0 {
		t.Errorf("GenerateRandomToken(0) = %q, want empty string", token)
	}
}
