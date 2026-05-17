package model

import (
	"strings"
	"testing"
	"time"
)

// ---- User tests ----

func validUser() *User {
	return &User{
		ID:       "u1",
		Username: "alice",
		Email:    "alice@example.com",
		Role:     RoleUser,
		Status:   StatusActive,
	}
}

func TestUser_Validate_Valid(t *testing.T) {
	u := validUser()
	if err := u.Validate(); err != nil {
		t.Errorf("valid user Validate() = %v, want nil", err)
	}
}

func TestUser_Validate_ShortUsername(t *testing.T) {
	u := validUser()
	u.Username = "ab"
	if err := u.Validate(); err != ErrInvalidUsername {
		t.Errorf("short username Validate() = %v, want ErrInvalidUsername", err)
	}
}

func TestUser_Validate_LongUsername(t *testing.T) {
	u := validUser()
	u.Username = strings.Repeat("a", 31)
	if err := u.Validate(); err != ErrInvalidUsername {
		t.Errorf("long username Validate() = %v, want ErrInvalidUsername", err)
	}
}

func TestUser_Validate_InvalidEmail(t *testing.T) {
	u := validUser()
	u.Email = "not-an-email"
	if err := u.Validate(); err != ErrInvalidEmail {
		t.Errorf("invalid email Validate() = %v, want ErrInvalidEmail", err)
	}
}

func TestUser_Validate_InvalidRole(t *testing.T) {
	u := validUser()
	u.Role = "superuser"
	if err := u.Validate(); err != ErrInvalidRole {
		t.Errorf("invalid role Validate() = %v, want ErrInvalidRole", err)
	}
}

func TestUser_Validate_InvalidStatus(t *testing.T) {
	u := validUser()
	u.Status = "banned"
	if err := u.Validate(); err != ErrInvalidStatus {
		t.Errorf("invalid status Validate() = %v, want ErrInvalidStatus", err)
	}
}

func TestUser_Validate_AllRoles(t *testing.T) {
	for _, role := range []string{RoleAdmin, RoleUser, RoleViewer} {
		u := validUser()
		u.Role = role
		if err := u.Validate(); err != nil {
			t.Errorf("Validate() with role %q = %v, want nil", role, err)
		}
	}
}

func TestUser_Validate_AllStatuses(t *testing.T) {
	for _, status := range []string{StatusActive, StatusSuspended, StatusPending} {
		u := validUser()
		u.Status = status
		if err := u.Validate(); err != nil {
			t.Errorf("Validate() with status %q = %v, want nil", status, err)
		}
	}
}

func TestUser_IsAdmin(t *testing.T) {
	u := validUser()
	u.Role = RoleAdmin
	if !u.IsAdmin() {
		t.Error("IsAdmin() = false for admin role, want true")
	}

	u.Role = RoleUser
	if u.IsAdmin() {
		t.Error("IsAdmin() = true for user role, want false")
	}
}

func TestUser_IsActive(t *testing.T) {
	u := validUser()
	u.Status = StatusActive
	if !u.IsActive() {
		t.Error("IsActive() = false for active status, want true")
	}

	u.Status = StatusSuspended
	if u.IsActive() {
		t.Error("IsActive() = true for suspended status, want false")
	}
}

func TestUser_CanLogin(t *testing.T) {
	u := validUser()
	u.Status = StatusActive
	u.EmailVerified = true
	if !u.CanLogin() {
		t.Error("CanLogin() = false for active+verified user, want true")
	}

	u.EmailVerified = false
	if u.CanLogin() {
		t.Error("CanLogin() = true for unverified user, want false")
	}

	u.EmailVerified = true
	u.Status = StatusSuspended
	if u.CanLogin() {
		t.Error("CanLogin() = true for suspended user, want false")
	}
}

func TestUser_UpdateLastLogin(t *testing.T) {
	u := validUser()
	before := time.Now()
	u.UpdateLastLogin()
	after := time.Now()

	if !u.LastLogin.Valid {
		t.Error("UpdateLastLogin() did not set LastLogin.Valid = true")
	}
	if u.LastLogin.Time.Before(before) || u.LastLogin.Time.After(after) {
		t.Errorf("UpdateLastLogin() set time %v outside expected range [%v, %v]",
			u.LastLogin.Time, before, after)
	}
}

func TestUser_SanitizeForJSON(t *testing.T) {
	u := validUser()
	u.PasswordHash = "secret-hash"
	u.TwoFactorSecret = "secret-2fa"
	u.PasswordResetToken = "reset-token"

	safe := u.SanitizeForJSON()
	if safe == nil {
		t.Fatal("SanitizeForJSON() returned nil")
	}
	if safe.PasswordHash != "" {
		t.Error("SanitizeForJSON() exposed PasswordHash")
	}
	if safe.TwoFactorSecret != "" {
		t.Error("SanitizeForJSON() exposed TwoFactorSecret")
	}
	if safe.PasswordResetToken != "" {
		t.Error("SanitizeForJSON() exposed PasswordResetToken")
	}
	if safe.Username != u.Username {
		t.Errorf("SanitizeForJSON().Username = %q, want %q", safe.Username, u.Username)
	}
}

// ---- Profile tests ----

func validProfile() *Profile {
	return &Profile{
		Slug:        "alice",
		DisplayName: "Alice",
		Bio:         "Hello world",
	}
}

func TestProfile_Validate_Valid(t *testing.T) {
	p := validProfile()
	if err := p.Validate(); err != nil {
		t.Errorf("valid profile Validate() = %v, want nil", err)
	}
}

func TestProfile_Validate_InvalidSlug(t *testing.T) {
	p := validProfile()
	p.Slug = "invalid slug!"
	if err := p.Validate(); err != ErrInvalidSlug {
		t.Errorf("invalid slug Validate() = %v, want ErrInvalidSlug", err)
	}
}

func TestProfile_Validate_EmptySlug(t *testing.T) {
	p := validProfile()
	p.Slug = ""
	if err := p.Validate(); err != ErrInvalidSlug {
		t.Errorf("empty slug Validate() = %v, want ErrInvalidSlug", err)
	}
}

func TestProfile_Validate_ValidSlugChars(t *testing.T) {
	for _, slug := range []string{"alice", "alice123", "alice-bob", "alice_bob", "ALICE"} {
		p := validProfile()
		p.Slug = slug
		if err := p.Validate(); err != nil {
			t.Errorf("Validate() with slug %q = %v, want nil", slug, err)
		}
	}
}

func TestProfile_Validate_DisplayNameTooLong(t *testing.T) {
	p := validProfile()
	p.DisplayName = strings.Repeat("a", 101)
	if err := p.Validate(); err != ErrDisplayNameTooLong {
		t.Errorf("long display name Validate() = %v, want ErrDisplayNameTooLong", err)
	}
}

func TestProfile_Validate_BioTooLong(t *testing.T) {
	p := validProfile()
	p.Bio = strings.Repeat("x", 501)
	if err := p.Validate(); err != ErrBioTooLong {
		t.Errorf("long bio Validate() = %v, want ErrBioTooLong", err)
	}
}

func TestProfile_Validate_MetaTitleTooLong(t *testing.T) {
	p := validProfile()
	p.MetaTitle = strings.Repeat("t", 61)
	if err := p.Validate(); err != ErrMetaTitleTooLong {
		t.Errorf("long meta title Validate() = %v, want ErrMetaTitleTooLong", err)
	}
}

func TestProfile_Validate_MetaDescriptionTooLong(t *testing.T) {
	p := validProfile()
	p.MetaDescription = strings.Repeat("d", 161)
	if err := p.Validate(); err != ErrMetaDescriptionTooLong {
		t.Errorf("long meta description Validate() = %v, want ErrMetaDescriptionTooLong", err)
	}
}

func TestProfile_GetPublicURL_NoCustomDomain(t *testing.T) {
	p := validProfile()
	p.Slug = "alice"
	url := p.GetPublicURL("https://example.com")
	if url != "https://example.com/alice" {
		t.Errorf("GetPublicURL() = %q, want https://example.com/alice", url)
	}
}

func TestProfile_GetPublicURL_CustomDomain_Unverified(t *testing.T) {
	p := validProfile()
	p.Slug = "alice"
	p.CustomDomain = "alice.io"
	p.DomainVerified = false
	url := p.GetPublicURL("https://example.com")
	if url != "https://example.com/alice" {
		t.Errorf("GetPublicURL() with unverified custom domain = %q, want https://example.com/alice", url)
	}
}

func TestProfile_GetPublicURL_CustomDomain_Verified(t *testing.T) {
	p := validProfile()
	p.CustomDomain = "alice.io"
	p.DomainVerified = true
	url := p.GetPublicURL("https://example.com")
	if url != "https://alice.io" {
		t.Errorf("GetPublicURL() with verified custom domain = %q, want https://alice.io", url)
	}
}

func TestProfile_IsAccessible(t *testing.T) {
	p := validProfile()
	p.IsPublic = true
	p.PasswordProtected = false
	if !p.IsAccessible() {
		t.Error("IsAccessible() = false for public unprotected profile, want true")
	}

	p.PasswordProtected = true
	if p.IsAccessible() {
		t.Error("IsAccessible() = true for password-protected profile, want false")
	}

	p.IsPublic = false
	p.PasswordProtected = false
	if p.IsAccessible() {
		t.Error("IsAccessible() = true for private profile, want false")
	}
}

func TestProfile_IncrementViewCount(t *testing.T) {
	p := validProfile()
	p.ViewCount = 0
	p.IncrementViewCount()
	if p.ViewCount != 1 {
		t.Errorf("IncrementViewCount() ViewCount = %d, want 1", p.ViewCount)
	}
	p.IncrementViewCount()
	if p.ViewCount != 2 {
		t.Errorf("IncrementViewCount() ViewCount = %d, want 2", p.ViewCount)
	}
}

// ---- ProfileTheme tests ----

func validProfileTheme() *ProfileTheme {
	return &ProfileTheme{
		BackgroundType: "color",
		ButtonStyle:    "rounded",
	}
}

func TestProfileTheme_Validate_Valid(t *testing.T) {
	pt := validProfileTheme()
	if err := pt.Validate(); err != nil {
		t.Errorf("valid theme Validate() = %v, want nil", err)
	}
}

func TestProfileTheme_Validate_AllBackgroundTypes(t *testing.T) {
	for _, bt := range []string{"color", "gradient", "image"} {
		pt := validProfileTheme()
		pt.BackgroundType = bt
		if err := pt.Validate(); err != nil {
			t.Errorf("Validate() with BackgroundType %q = %v, want nil", bt, err)
		}
	}
}

func TestProfileTheme_Validate_InvalidBackgroundType(t *testing.T) {
	pt := validProfileTheme()
	pt.BackgroundType = "video"
	if err := pt.Validate(); err != ErrInvalidBackgroundType {
		t.Errorf("invalid BackgroundType Validate() = %v, want ErrInvalidBackgroundType", err)
	}
}

func TestProfileTheme_Validate_AllButtonStyles(t *testing.T) {
	for _, bs := range []string{"rounded", "square", "pill"} {
		pt := validProfileTheme()
		pt.ButtonStyle = bs
		if err := pt.Validate(); err != nil {
			t.Errorf("Validate() with ButtonStyle %q = %v, want nil", bs, err)
		}
	}
}

func TestProfileTheme_Validate_InvalidButtonStyle(t *testing.T) {
	pt := validProfileTheme()
	pt.ButtonStyle = "circle"
	if err := pt.Validate(); err != ErrInvalidButtonStyle {
		t.Errorf("invalid ButtonStyle Validate() = %v, want ErrInvalidButtonStyle", err)
	}
}

// ---- Link tests ----

func TestLink_Validate_Valid(t *testing.T) {
	l := &Link{Title: "GitHub", URL: "https://github.com/alice"}
	if err := l.Validate(); err != nil {
		t.Errorf("valid link Validate() = %v, want nil", err)
	}
}

func TestLink_Validate_TitleTooLong(t *testing.T) {
	l := &Link{Title: strings.Repeat("x", 101), URL: "https://example.com"}
	if err := l.Validate(); err != ErrTitleTooLong {
		t.Errorf("long title Validate() = %v, want ErrTitleTooLong", err)
	}
}

func TestLink_Validate_InvalidURL(t *testing.T) {
	l := &Link{Title: "Bad", URL: "not-a-url"}
	if err := l.Validate(); err != ErrInvalidURL {
		t.Errorf("invalid URL Validate() = %v, want ErrInvalidURL", err)
	}
}

func TestLink_Validate_EmptyURL(t *testing.T) {
	l := &Link{Title: "Empty", URL: ""}
	if err := l.Validate(); err != ErrInvalidURL {
		t.Errorf("empty URL Validate() = %v, want ErrInvalidURL", err)
	}
}

func TestLink_IncrementClickCount(t *testing.T) {
	l := &Link{ClickCount: 0}
	l.IncrementClickCount()
	if l.ClickCount != 1 {
		t.Errorf("IncrementClickCount() = %d, want 1", l.ClickCount)
	}
}

func TestLink_Toggle(t *testing.T) {
	l := &Link{IsActive: true}
	l.Toggle()
	if l.IsActive {
		t.Error("Toggle() did not deactivate link")
	}
	l.Toggle()
	if !l.IsActive {
		t.Error("Toggle() did not activate link")
	}
}

func TestLink_GetDisplayText(t *testing.T) {
	l := &Link{Username: "alice"}
	if got := l.GetDisplayText("GitHub", true); got != "alice@GitHub" {
		t.Errorf("GetDisplayText(show=true) = %q, want alice@GitHub", got)
	}
	if got := l.GetDisplayText("GitHub", false); got != "GitHub" {
		t.Errorf("GetDisplayText(show=false) = %q, want GitHub", got)
	}
	l.Username = ""
	if got := l.GetDisplayText("GitHub", true); got != "GitHub" {
		t.Errorf("GetDisplayText(empty username) = %q, want GitHub", got)
	}
}

// ---- FooterItem tests ----

func TestFooterItem_Validate_ValidTypes(t *testing.T) {
	for _, ft := range []string{FooterTypeText, FooterTypeLink, FooterTypeSocialRow, FooterTypeBadge, FooterTypeHTML} {
		f := &FooterItem{ItemType: ft}
		if err := f.Validate(); err != nil {
			t.Errorf("Validate() with FooterType %q = %v, want nil", ft, err)
		}
	}
}

func TestFooterItem_Validate_InvalidType(t *testing.T) {
	f := &FooterItem{ItemType: "video"}
	if err := f.Validate(); err != ErrInvalidFooterType {
		t.Errorf("invalid footer type Validate() = %v, want ErrInvalidFooterType", err)
	}
}

// ---- Shortlink tests ----

func TestShortlink_Validate_Valid(t *testing.T) {
	s := &Shortlink{TargetURL: "https://example.com"}
	if err := s.Validate(); err != nil {
		t.Errorf("valid shortlink Validate() = %v, want nil", err)
	}
}

func TestShortlink_Validate_InvalidURL(t *testing.T) {
	s := &Shortlink{TargetURL: "not-a-url"}
	if err := s.Validate(); err != ErrInvalidURL {
		t.Errorf("invalid URL shortlink Validate() = %v, want ErrInvalidURL", err)
	}
}

func TestShortlink_IncrementClickCount(t *testing.T) {
	s := &Shortlink{ClickCount: 5}
	s.IncrementClickCount()
	if s.ClickCount != 6 {
		t.Errorf("IncrementClickCount() = %d, want 6", s.ClickCount)
	}
}

func TestShortlink_IsExpired_NoExpiry(t *testing.T) {
	s := &Shortlink{}
	if s.IsExpired() {
		t.Error("IsExpired() = true for shortlink with no expiry, want false")
	}
}

func TestShortlink_IsExpired_Future(t *testing.T) {
	future := time.Now().Add(time.Hour)
	s := &Shortlink{ExpiresAt: &future}
	if s.IsExpired() {
		t.Error("IsExpired() = true for future expiry, want false")
	}
}

func TestShortlink_IsExpired_Past(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	s := &Shortlink{ExpiresAt: &past}
	if !s.IsExpired() {
		t.Error("IsExpired() = false for past expiry, want true")
	}
}

// ---- QRCodeSettings tests ----

// TestIsValidURL_ParseError exercises the url.Parse error branch inside
// isValidURL. A URL with an invalid IPv6 bracket causes url.Parse to return an
// error on some Go versions; the fallback uses invalid percent-encoding which
// reliably triggers a parse error.
func TestIsValidURL_ParseError(t *testing.T) {
	// "http://bad%zz.com" has invalid percent-encoding — url.Parse returns error.
	if isValidURL("http://bad%zz.com") {
		t.Error("isValidURL(bad percent-encoding) = true, want false")
	}
}

func TestQRCodeSettings_Validate_ValidSizes(t *testing.T) {
	for _, size := range []int{128, 256, 512, 1024} {
		q := &QRCodeSettings{Size: size}
		if err := q.Validate(); err != nil {
			t.Errorf("Validate() with size %d = %v, want nil", size, err)
		}
	}
}

func TestQRCodeSettings_Validate_InvalidSize(t *testing.T) {
	q := &QRCodeSettings{Size: 200}
	if err := q.Validate(); err != ErrInvalidQRCodeSize {
		t.Errorf("invalid size Validate() = %v, want ErrInvalidQRCodeSize", err)
	}
}
