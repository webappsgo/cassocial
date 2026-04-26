package store

import (
	"context"
	"database/sql"
	"time"
)

// Store defines the interface for data access operations
// Per PART 24: Store interface for all data models from PART 36
type Store interface {
	// Connection management
	Ping() error
	Close() error
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	// Settings operations (PART 5: Configuration)
	GetSetting(key string) (string, error)
	SetSetting(key, value string) error
	GetAllSettings() (map[string]string, error)

	// User operations (Regular Users - PART 23)
	GetUserByID(id string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	CreateUser(user *User) error
	UpdateUser(user *User) error
	DeleteUser(id string) error
	ListUsers(limit, offset int) ([]*User, error)
	CountUsers() (int, error)

	// Server Admin operations (PART 23: separate from Regular Users)
	GetServerAdminByID(id string) (*ServerAdmin, error)
	GetServerAdminByEmail(email string) (*ServerAdmin, error)
	GetServerAdminByUsername(username string) (*ServerAdmin, error)
	CreateServerAdmin(admin *ServerAdmin) error
	UpdateServerAdmin(admin *ServerAdmin) error
	DeleteServerAdmin(id string) error
	GetPrimaryAdmin() (*ServerAdmin, error)
	ListServerAdmins() ([]*ServerAdmin, error)

	// Profile operations (PART 36: Profile model)
	GetProfileByID(id string) (*Profile, error)
	GetProfileBySlug(slug string) (*Profile, error)
	GetProfileByCustomDomain(domain string) (*Profile, error)
	GetProfilesByUserID(userID string) ([]*Profile, error)
	CreateProfile(profile *Profile) error
	UpdateProfile(profile *Profile) error
	DeleteProfile(id string) error
	CountProfilesByUserID(userID string) (int, error)
	IncrementProfileViewCount(profileID string) error

	// Profile Theme operations (PART 36: ProfileTheme model)
	GetProfileTheme(profileID string) (*ProfileTheme, error)
	UpdateProfileTheme(theme *ProfileTheme) error
	DeleteProfileTheme(profileID string) error

	// QR Code Settings operations (PART 36: QRCodeSettings model)
	GetQRCodeSettings(profileID string) (*QRCodeSettings, error)
	UpdateQRCodeSettings(settings *QRCodeSettings) error
	DeleteQRCodeSettings(profileID string) error

	// Service operations (PART 36: Service model - 5000+ services)
	GetServiceByID(id string) (*Service, error)
	GetServiceByName(name string) (*Service, error)
	ListServices(category string, limit, offset int) ([]*Service, error)
	SearchServices(query string, limit int) ([]*Service, error)
	CreateService(service *Service) error
	UpdateService(service *Service) error
	DeleteService(id string) error
	CountServices() (int, error)

	// Link operations (PART 36: Link model)
	GetLinkByID(id string) (*Link, error)
	GetLinksByProfileID(profileID string) ([]*Link, error)
	CreateLink(link *Link) error
	UpdateLink(link *Link) error
	DeleteLink(id string) error
	ReorderLinks(profileID string, linkIDs []string) error
	CountLinksByProfileID(profileID string) (int, error)
	IncrementLinkClickCount(linkID string) error

	// Footer Item operations (PART 36: FooterItem model)
	GetFooterItemsByProfileID(profileID string) ([]*FooterItem, error)
	CreateFooterItem(item *FooterItem) error
	UpdateFooterItem(item *FooterItem) error
	DeleteFooterItem(id string) error

	// Shortlink operations (PART 36: Shortlink model)
	GetShortlinkByID(id string) (*Shortlink, error)
	GetShortlinkByCode(code string) (*Shortlink, error)
	GetShortlinksByProfileID(profileID string) ([]*Shortlink, error)
	CreateShortlink(shortlink *Shortlink) error
	UpdateShortlink(shortlink *Shortlink) error
	DeleteShortlink(id string) error
	IncrementShortlinkClickCount(id string) error
	DeleteExpiredShortlinks() error

	// Session operations (PART 23: Authentication)
	CreateSession(session *Session) error
	GetSession(sessionID string) (*Session, error)
	DeleteSession(sessionID string) error
	DeleteSessionsByUserID(userID string) error
	CleanupExpiredSessions() error

	// Email verification operations (PART 23)
	CreateEmailVerificationToken(token *EmailVerificationToken) error
	GetEmailVerificationToken(token string) (*EmailVerificationToken, error)
	DeleteEmailVerificationToken(token string) error
	DeleteExpiredEmailVerificationTokens() error

	// Password reset operations (PART 23)
	CreatePasswordResetToken(token *PasswordResetToken) error
	GetPasswordResetToken(token string) (*PasswordResetToken, error)
	DeletePasswordResetToken(token string) error
	DeleteExpiredPasswordResetTokens() error

	// Analytics operations (PART 36: ProfileView and LinkClick models)
	// Per PART 36: IP addresses are hashed for GDPR compliance
	RecordProfileView(view *ProfileView) error
	RecordLinkClick(click *LinkClick) error
	GetProfileAnalytics(profileID string, days int) (*ProfileAnalytics, error)
	GetLinkAnalytics(linkID string, days int) (*LinkAnalytics, error)
	GetTopLinks(profileID string, limit int) ([]*LinkStat, error)
	GetTopReferrers(profileID string, limit int) ([]*ReferrerStat, error)

	// Cluster operations (PART 24: Cluster support)
	CreateClusterNode(node *ClusterNode) error
	UpdateClusterNode(node *ClusterNode) error
	GetClusterNode(id string) (*ClusterNode, error)
	ListClusterNodes() ([]*ClusterNode, error)
	UpdateNodeHeartbeat(id string) error
	DeleteClusterNode(id string) error
	GetPrimaryNode() (*ClusterNode, error)
	MarkNodeOffline(id string) error

	// Profile Tags operations (PART 36: search/categorization)
	AddProfileTag(profileID, tag string) error
	RemoveProfileTag(profileID, tag string) error
	GetProfileTags(profileID string) ([]string, error)
	SearchProfilesByTag(tag string, limit, offset int) ([]*Profile, error)

	// Migrations (PART 24: Self-creating schema)
	RunMigrations() error
}

// User represents a user in the system (Regular User)
// Per PART 23: Regular Users are separate from Server Admins
type User struct {
	ID               string
	Username         string
	Email            string
	PasswordHash     string
	Role             string
	Status           string
	EmailVerified    bool
	TwoFactorEnabled bool
	TwoFactorSecret  string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	LastLogin        *time.Time
}

// ServerAdmin represents an application administrator
// Per PART 23: Server Admins stored in separate table from Regular Users
type ServerAdmin struct {
	ID               string
	Username         string
	Email            string
	PasswordHash     string
	IsPrimary        bool
	TwoFactorEnabled bool
	TwoFactorSecret  string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	LastLogin        *time.Time
}

// Profile represents a user profile (landing page)
// Per PART 36: Profile is a user's public landing page
type Profile struct {
	ID                  string
	UserID              string
	Slug                string
	DisplayName         string
	Bio                 string
	AvatarURL           string
	HeaderImageURL      string
	ThemeID             string
	CustomCSS           string
	ShowUsernames       bool
	IsPublic            bool
	PasswordProtected   bool
	ProtectionPassword  string
	CustomDomain        string
	DomainVerified      bool
	AnalyticsEnabled    bool
	MetaTitle           string
	MetaDescription     string
	OgImageURL          string
	ViewCount           int
	QRCodeEnabled       bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// ProfileTheme represents custom theme settings for a profile
// Per PART 36: Theme customization model
type ProfileTheme struct {
	ProfileID              string
	BackgroundType         string
	BackgroundValue        string
	ButtonStyle            string
	ButtonAnimation        string
	ButtonShadow           string
	FontOverride           string
	CustomCSS              string
	LinkThumbnailPosition  string
	UpdatedAt              time.Time
}

// QRCodeSettings represents QR code generation settings
// Per PART 36: QR code customization model
type QRCodeSettings struct {
	ProfileID       string
	Size            int
	ErrorCorrection string
	Style           string
	DarkColor       string
	LightColor      string
	LogoEnabled     bool
	LogoSize        int
	Format          string
	UpdatedAt       time.Time
}

// Service represents a predefined service (social media, etc.)
// Per PART 36: 5000+ predefined services
type Service struct {
	ID                string
	Name              string
	Category          string
	IconURL           string
	IconSVG           string
	URLPattern        string
	BackgroundColor   string
	TextColor         string
	Popularity        int
	IsActive          bool
	RequiresUsername  bool
	PlaceholderText   string
	ValidationPattern string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Link represents a link in a profile
// Per PART 36: Link on a user's profile
type Link struct {
	ID              string
	ProfileID       string
	ServiceID       string
	Title           string
	Username        string
	URL             string
	IconURL         string
	BackgroundColor string
	TextColor       string
	Position        int
	IsActive        bool
	ClickCount      int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// FooterItem represents a footer element on a profile
// Per PART 36: Footer customization
type FooterItem struct {
	ID        string
	ProfileID string
	ItemType  string
	Content   string
	Position  int
	IsActive  bool
	CreatedAt time.Time
}

// Shortlink represents a shortened URL
// Per PART 36: URL shortener functionality
type Shortlink struct {
	ID         string
	ShortCode  string
	TargetURL  string
	ProfileID  string
	Title      string
	ClickCount int
	ExpiresAt  *time.Time
	CreatedAt  time.Time
}

// Session represents a user session
// Per PART 23: Session management for authentication
type Session struct {
	ID        string
	UserID    string
	UserType  string
	Username  string
	Role      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// EmailVerificationToken represents an email verification token
// Per PART 23: Email verification flow
type EmailVerificationToken struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// PasswordResetToken represents a password reset token
// Per PART 23: Password reset flow
type PasswordResetToken struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// ProfileView represents a profile page view (analytics)
// Per PART 36: Profile view tracking with hashed IPs for GDPR
type ProfileView struct {
	ProfileID string
	ViewerIP  string
	Referrer  string
	UserAgent string
	Country   string
	Timestamp time.Time
}

// LinkClick represents a link click (analytics)
// Per PART 36: Link click tracking with hashed IPs for GDPR
type LinkClick struct {
	LinkID    string
	ClickerIP string
	Referrer  string
	UserAgent string
	Country   string
	Timestamp time.Time
}

// ProfileAnalytics represents aggregated analytics for a profile
// Per PART 36: Analytics data structure
type ProfileAnalytics struct {
	ProfileID    string
	Views        int
	UniqueIPs    int
	Clicks       int
	TopLinks     []*LinkStat
	TopReferrers []*ReferrerStat
}

// LinkAnalytics represents aggregated analytics for a link
type LinkAnalytics struct {
	LinkID       string
	Clicks       int
	UniqueIPs    int
	TopReferrers []*ReferrerStat
}

// LinkStat represents click statistics for a link
type LinkStat struct {
	LinkID string
	Title  string
	Clicks int
}

// ReferrerStat represents referer statistics
type ReferrerStat struct {
	Referrer string
	Count    int
}

// ClusterNode represents a node in a cluster
// Per PART 24: Cluster support with heartbeat monitoring
type ClusterNode struct {
	ID            string
	Hostname      string
	Address       string
	Port          int
	Status        string
	IsPrimary     bool
	LastHeartbeat time.Time
	CreatedAt     time.Time
}
