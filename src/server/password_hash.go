package server

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters (OWASP 2023 recommendations - NON-NEGOTIABLE per TEMPLATE.md)
const (
	ArgonTime    = 3         // iterations
	ArgonMemory  = 64 * 1024 // 64 MB
	ArgonThreads = 4         // parallelism
	ArgonKeyLen  = 32        // output length in bytes
	ArgonSaltLen = 16        // salt length in bytes
)

// HashPassword hashes a password using Argon2id
// Returns PHC string format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
func HashPassword(password string) (string, error) {
	// Generate random salt
	salt := make([]byte, ArgonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password using Argon2id
	hash := argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLen)

	// Encode as PHC string format
	return encodeArgon2Hash(salt, hash), nil
}

// VerifyPassword verifies a password against an Argon2id hash.
func VerifyPassword(password, hash string) bool {
	if !strings.HasPrefix(hash, "$argon2id$") {
		return false
	}
	return verifyArgon2Password(password, hash)
}

// verifyArgon2Password verifies Argon2id hashed password
func verifyArgon2Password(password, encodedHash string) bool {
	salt, hash, err := decodeArgon2Hash(encodedHash)
	if err != nil {
		return false
	}

	// Hash the input password with the same salt
	inputHash := argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLen)

	// Compare hashes (constant-time comparison)
	return subtle.ConstantTimeCompare(hash, inputHash) == 1
}

// encodeArgon2Hash encodes salt and hash into PHC string format
func encodeArgon2Hash(salt, hash []byte) string {
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		ArgonMemory, ArgonTime, ArgonThreads, saltB64, hashB64)
}

// decodeArgon2Hash decodes PHC string format into salt and hash
func decodeArgon2Hash(encodedHash string) (salt, hash []byte, err error) {
	// Expected format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, fmt.Errorf("not an argon2id hash")
	}

	// Decode salt
	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Decode hash
	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	return salt, hash, nil
}

// HashToken hashes an API token using SHA-256
// Used for API tokens and session tokens (fast lookup needed)
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash[:])
}

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return fmt.Sprintf("%x", bytes), nil
}
