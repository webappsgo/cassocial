package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/casapps/cassocial/src/server/store"
	"github.com/casapps/cassocial/src/server/model"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrUserNotFound           = errors.New("user not found")
	ErrUserNotActive          = errors.New("user account is not active")
	ErrEmailNotVerified       = errors.New("email not verified")
	ErrUsernameExists         = errors.New("username already exists")
	ErrEmailExists            = errors.New("email already exists")
	ErrWeakPassword           = errors.New("password does not meet requirements")
	ErrInvalidToken           = errors.New("invalid or expired token")
	Err2FARequired            = errors.New("two-factor authentication required")
	ErrInvalidVerificationToken = errors.New("invalid verification token")
)

// Auth handles authentication operations
type Auth struct {
	db        *store.DB
	jwtSecret []byte
}

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// PasswordRequirements stores password validation rules
type PasswordRequirements struct {
	MinLength       int
	RequireUpper    bool
	RequireNumber   bool
	RequireSpecial  bool
}

// NewAuth creates a new Auth instance
func NewAuth(db *store.DB, jwtSecret string) *Auth {
	if jwtSecret == "" {
		// Generate a random secret if none provided
		jwtSecret = generateRandomString(32)
	}
	return &Auth{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register creates a new user account
func (a *Auth) Register(username, email, password string) (*model.User, error) {
	// Validate password requirements
	if err := a.ValidatePassword(password); err != nil {
		return nil, err
	}

	// Check if username exists
	exists, err := a.usernameExists(username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, ErrUsernameExists
	}

	// Check if email exists
	exists, err = a.emailExists(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password using Argon2id
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Determine if email verification is required
	emailVerificationRequired, _ := a.db.GetSetting("email_verification_required")
	emailVerified := emailVerificationRequired != "true"

	// Determine default status
	registrationRequiresApproval, _ := a.db.GetSetting("registration_requires_approval")
	status := model.StatusActive
	if registrationRequiresApproval == "true" {
		status = model.StatusPending
	}

	// Create user
	user := &model.User{
		Username:         username,
		Email:            email,
		PasswordHash:     passwordHash,
		Role:             model.RoleUser,
		Status:           status,
		EmailVerified:    emailVerified,
		TwoFactorEnabled: false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Validate user model
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Insert user into database
	query := `
		INSERT INTO users (id, username, email, password_hash, role, status, created_at, updated_at, email_verified, two_factor_enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	if a.db.Driver == "postgres" {
		query = `
			INSERT INTO users (username, email, password_hash, role, status, created_at, updated_at, email_verified, two_factor_enabled)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`
		err = a.db.QueryRow(query, user.Username, user.Email, user.PasswordHash, user.Role, user.Status,
			user.CreatedAt, user.UpdatedAt, user.EmailVerified, user.TwoFactorEnabled).Scan(&user.ID)
	} else {
		userID := generateUUID()
		user.ID = userID
		_, err = a.db.Exec(query, userID, user.Username, user.Email, user.PasswordHash, user.Role, user.Status,
			user.CreatedAt, user.UpdatedAt, user.EmailVerified, user.TwoFactorEnabled)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (a *Auth) Login(username, password string) (string, *model.User, error) {
	// Get user by username or email
	user, err := a.GetUserByUsernameOrEmail(username)
	if err != nil {
		if err == ErrUserNotFound || err == sql.ErrNoRows {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive() {
		return "", nil, ErrUserNotActive
	}

	// Check if email is verified (if required)
	emailVerificationRequired, _ := a.db.GetSetting("email_verification_required")
	if emailVerificationRequired == "true" && !user.EmailVerified {
		return "", nil, ErrEmailNotVerified
	}

	// Verify password
	if !VerifyPassword(password, user.PasswordHash) {
		return "", nil, ErrInvalidCredentials
	}

	// Check if 2FA is enabled
	if user.TwoFactorEnabled {
		// Return special error indicating 2FA is required
		// The caller should then prompt for 2FA code
		return "", user, Err2FARequired
	}

	// Generate JWT token
	token, err := a.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login
	user.UpdateLastLogin()
	a.updateLastLogin(user.ID)

	return token, user, nil
}

// LoginWith2FA authenticates a user with 2FA code
func (a *Auth) LoginWith2FA(userID, code string) (string, *model.User, error) {
	// Get user
	user, err := a.GetUserByID(userID)
	if err != nil {
		return "", nil, err
	}

	// Verify 2FA code
	valid, err := a.Verify2FACode(user, code)
	if err != nil || !valid {
		return "", nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := a.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login
	user.UpdateLastLogin()
	a.updateLastLogin(user.ID)

	return token, user, nil
}

// GenerateToken creates a JWT token for a user
func (a *Auth) GenerateToken(user *model.User) (string, error) {
	// Get session timeout from settings (default 1440 minutes = 24 hours)
	sessionTimeoutStr, err := a.db.GetSetting("session_timeout_minutes")
	if err != nil {
		sessionTimeoutStr = "1440"
	}
	sessionTimeout, _ := strconv.Atoi(sessionTimeoutStr)
	if sessionTimeout == 0 {
		sessionTimeout = 1440
	}

	expirationTime := time.Now().Add(time.Duration(sessionTimeout) * time.Minute)

	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "cassocial",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (a *Auth) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken generates a new token from an existing valid token
func (a *Auth) RefreshToken(tokenString string) (string, error) {
	claims, err := a.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Get user to ensure they still exist and are active
	user, err := a.GetUserByID(claims.UserID)
	if err != nil {
		return "", err
	}

	if !user.IsActive() {
		return "", ErrUserNotActive
	}

	// Generate new token
	return a.GenerateToken(user)
}

// ValidatePassword validates a password against requirements
func (a *Auth) ValidatePassword(password string) error {
	reqs, err := a.GetPasswordRequirements()
	if err != nil {
		// Use defaults if can't get from database
		reqs = &PasswordRequirements{
			MinLength:      8,
			RequireUpper:   true,
			RequireNumber:  true,
			RequireSpecial: false,
		}
	}

	if len(password) < reqs.MinLength {
		return fmt.Errorf("%w: password must be at least %d characters", ErrWeakPassword, reqs.MinLength)
	}

	if reqs.RequireUpper && !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return fmt.Errorf("%w: password must contain at least one uppercase letter", ErrWeakPassword)
	}

	if reqs.RequireNumber && !strings.ContainsAny(password, "0123456789") {
		return fmt.Errorf("%w: password must contain at least one number", ErrWeakPassword)
	}

	if reqs.RequireSpecial && !strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?") {
		return fmt.Errorf("%w: password must contain at least one special character", ErrWeakPassword)
	}

	return nil
}

// GetPasswordRequirements retrieves password requirements from settings
func (a *Auth) GetPasswordRequirements() (*PasswordRequirements, error) {
	minLengthStr, _ := a.db.GetSetting("password_min_length")
	requireUpperStr, _ := a.db.GetSetting("password_require_uppercase")
	requireNumberStr, _ := a.db.GetSetting("password_require_number")
	requireSpecialStr, _ := a.db.GetSetting("password_require_special")

	minLength, _ := strconv.Atoi(minLengthStr)
	if minLength == 0 {
		minLength = 8
	}

	return &PasswordRequirements{
		MinLength:      minLength,
		RequireUpper:   requireUpperStr == "true",
		RequireNumber:  requireNumberStr == "true",
		RequireSpecial: requireSpecialStr == "true",
	}, nil
}

// GetUserByID retrieves a user by ID
func (a *Auth) GetUserByID(id string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, email, password_hash, role, status, created_at, updated_at,
			  last_login, email_verified, two_factor_enabled, two_factor_secret
			  FROM users WHERE id = ?`

	if a.db.Driver == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
	}

	var twoFactorSecret sql.NullString
	err := a.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.EmailVerified,
		&user.TwoFactorEnabled, &twoFactorSecret,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user.TwoFactorSecret = twoFactorSecret.String
	return user, nil
}

// GetUserByUsername retrieves a user by username
func (a *Auth) GetUserByUsername(username string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, email, password_hash, role, status, created_at, updated_at,
			  last_login, email_verified, two_factor_enabled, two_factor_secret
			  FROM users WHERE username = ?`

	if a.db.Driver == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
	}

	var twoFactorSecret sql.NullString
	err := a.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.EmailVerified,
		&user.TwoFactorEnabled, &twoFactorSecret,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user.TwoFactorSecret = twoFactorSecret.String
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (a *Auth) GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, email, password_hash, role, status, created_at, updated_at,
			  last_login, email_verified, two_factor_enabled, two_factor_secret
			  FROM users WHERE email = ?`

	if a.db.Driver == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
	}

	var twoFactorSecret sql.NullString
	err := a.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.EmailVerified,
		&user.TwoFactorEnabled, &twoFactorSecret,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user.TwoFactorSecret = twoFactorSecret.String
	return user, nil
}

// GetUserByUsernameOrEmail retrieves a user by username or email
func (a *Auth) GetUserByUsernameOrEmail(usernameOrEmail string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, email, password_hash, role, status, created_at, updated_at,
			  last_login, email_verified, two_factor_enabled, two_factor_secret
			  FROM users WHERE username = ? OR email = ?`

	if a.db.Driver == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
		query = strings.Replace(query, "?", "$2", 1)
	}

	var twoFactorSecret sql.NullString
	err := a.db.QueryRow(query, usernameOrEmail, usernameOrEmail).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.EmailVerified,
		&user.TwoFactorEnabled, &twoFactorSecret,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user.TwoFactorSecret = twoFactorSecret.String
	return user, nil
}

// usernameExists checks if a username already exists
func (a *Auth) usernameExists(username string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE username = ?"

	if a.db.Driver == "postgres" {
		query = "SELECT COUNT(*) FROM users WHERE username = $1"
	}

	err := a.db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// emailExists checks if an email already exists
func (a *Auth) emailExists(email string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE email = ?"

	if a.db.Driver == "postgres" {
		query = "SELECT COUNT(*) FROM users WHERE email = $1"
	}

	err := a.db.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// updateLastLogin updates the last login timestamp for a user
func (a *Auth) updateLastLogin(userID string) error {
	query := "UPDATE users SET last_login = ? WHERE id = ?"

	if a.db.Driver == "postgres" {
		query = "UPDATE users SET last_login = $1 WHERE id = $2"
	}

	_, err := a.db.Exec(query, time.Now(), userID)
	return err
}

// generateRandomString generates a random hex string of specified length
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

// generateUUID generates a simple UUID
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
