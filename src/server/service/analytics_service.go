package service

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/casapps/cassocial/src/server/store"
	"github.com/casapps/cassocial/src/server/model"
)

// AnalyticsService handles analytics tracking and aggregation
type AnalyticsService struct {
	db *store.DB
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *store.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// TrackView tracks a profile view event
func (s *AnalyticsService) TrackView(profileID, ipAddress, userAgent, referrer string) error {
	// Check if analytics is enabled
	enabled, err := s.isAnalyticsEnabled(profileID)
	if err != nil || !enabled {
		return nil // Silently ignore if disabled
	}

	// Hash IP address for privacy
	ipHash := s.hashIP(ipAddress)

	// Parse device type from user agent
	deviceType := s.parseDeviceType(userAgent)

	// Get country from IP (simplified - in production use GeoIP database)
	country := s.getCountryFromIP(ipAddress)

	// Create analytics event
	event := &model.Analytics{
		ID:         s.generateID(),
		ProfileID:  profileID,
		EventType:  model.EventTypeView,
		IPHash:     ipHash,
		UserAgent:  userAgent,
		Referrer:   referrer,
		Country:    country,
		DeviceType: deviceType,
		CreatedAt:  time.Now(),
	}

	// Validate event
	if err := event.Validate(); err != nil {
		return fmt.Errorf("analytics validation failed: %w", err)
	}

	// Insert into database
	query := `
		INSERT INTO analytics (
			id, profile_id, link_id, event_type, ip_hash, user_agent,
			referrer, country, device_type, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		event.ID, event.ProfileID, nil, event.EventType, event.IPHash,
		event.UserAgent, event.Referrer, event.Country, event.DeviceType,
		event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert analytics event: %w", err)
	}

	return nil
}

// TrackClick tracks a link click event
func (s *AnalyticsService) TrackClick(profileID, linkID, ipAddress, userAgent, referrer string) error {
	// Check if analytics is enabled
	enabled, err := s.isAnalyticsEnabled(profileID)
	if err != nil || !enabled {
		return nil // Silently ignore if disabled
	}

	// Hash IP address for privacy
	ipHash := s.hashIP(ipAddress)

	// Parse device type from user agent
	deviceType := s.parseDeviceType(userAgent)

	// Get country from IP
	country := s.getCountryFromIP(ipAddress)

	// Create analytics event
	event := &model.Analytics{
		ID:         s.generateID(),
		ProfileID:  profileID,
		LinkID:     linkID,
		EventType:  model.EventTypeClick,
		IPHash:     ipHash,
		UserAgent:  userAgent,
		Referrer:   referrer,
		Country:    country,
		DeviceType: deviceType,
		CreatedAt:  time.Now(),
	}

	// Validate event
	if err := event.Validate(); err != nil {
		return fmt.Errorf("analytics validation failed: %w", err)
	}

	// Insert into database
	query := `
		INSERT INTO analytics (
			id, profile_id, link_id, event_type, ip_hash, user_agent,
			referrer, country, device_type, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		event.ID, event.ProfileID, event.LinkID, event.EventType, event.IPHash,
		event.UserAgent, event.Referrer, event.Country, event.DeviceType,
		event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert analytics event: %w", err)
	}

	return nil
}

// TrackSession creates or updates an analytics session
func (s *AnalyticsService) TrackSession(session *model.AnalyticsSession) error {
	// Check if analytics is enabled
	enabled, err := s.isAnalyticsEnabled(session.ProfileID)
	if err != nil || !enabled {
		return nil
	}

	// Validate session
	if err := session.Validate(); err != nil {
		return fmt.Errorf("session validation failed: %w", err)
	}

	// Check if session exists
	exists, err := s.sessionExists(session.SessionID)
	if err != nil {
		return fmt.Errorf("failed to check session existence: %w", err)
	}

	if exists {
		// Update existing session
		query := `
			UPDATE analytics_sessions SET
				duration_seconds = ?, link_clicks = ?, updated_at = ?
			WHERE session_id = ?
		`

		_, err = s.db.Exec(query,
			session.DurationSeconds, session.LinkClicks, time.Now(), session.SessionID,
		)

		if err != nil {
			return fmt.Errorf("failed to update session: %w", err)
		}
	} else {
		// Create new session
		if session.ID == "" {
			session.ID = s.generateID()
		}
		if session.CreatedAt.IsZero() {
			session.CreatedAt = time.Now()
		}

		query := `
			INSERT INTO analytics_sessions (
				id, profile_id, session_id, ip_hash, country, region, city,
				device_type, browser, os, referrer_domain, referrer_path,
				utm_source, utm_medium, utm_campaign, landing_page,
				duration_seconds, link_clicks, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err = s.db.Exec(query,
			session.ID, session.ProfileID, session.SessionID, session.IPHash,
			session.Country, session.Region, session.City, session.DeviceType,
			session.Browser, session.OS, session.ReferrerDomain, session.ReferrerPath,
			session.UTMSource, session.UTMMedium, session.UTMCampaign,
			session.LandingPage, session.DurationSeconds, session.LinkClicks,
			session.CreatedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to insert session: %w", err)
		}
	}

	return nil
}

// AggregateHourly aggregates analytics data into hourly statistics
func (s *AnalyticsService) AggregateHourly(profileID string, hour time.Time) error {
	// Round to hour
	hour = hour.Truncate(time.Hour)

	// Calculate statistics for the hour
	query := `
		INSERT INTO analytics_hourly (
			profile_id, hour, views, unique_visitors, total_clicks,
			avg_duration_seconds, top_referrer, top_country
		)
		SELECT
			? as profile_id,
			? as hour,
			COUNT(CASE WHEN event_type = 'view' THEN 1 END) as views,
			COUNT(DISTINCT CASE WHEN event_type = 'view' THEN ip_hash END) as unique_visitors,
			COUNT(CASE WHEN event_type = 'click' THEN 1 END) as total_clicks,
			COALESCE(AVG(s.duration_seconds), 0) as avg_duration_seconds,
			(SELECT referrer FROM analytics WHERE profile_id = ? AND created_at >= ? AND created_at < ?
				AND referrer != '' GROUP BY referrer ORDER BY COUNT(*) DESC LIMIT 1) as top_referrer,
			(SELECT country FROM analytics WHERE profile_id = ? AND created_at >= ? AND created_at < ?
				AND country != '' GROUP BY country ORDER BY COUNT(*) DESC LIMIT 1) as top_country
		FROM analytics a
		LEFT JOIN analytics_sessions s ON a.profile_id = s.profile_id
			AND s.created_at >= ? AND s.created_at < ?
		WHERE a.profile_id = ? AND a.created_at >= ? AND a.created_at < ?
		ON CONFLICT(profile_id, hour) DO UPDATE SET
			views = EXCLUDED.views,
			unique_visitors = EXCLUDED.unique_visitors,
			total_clicks = EXCLUDED.total_clicks,
			avg_duration_seconds = EXCLUDED.avg_duration_seconds,
			top_referrer = EXCLUDED.top_referrer,
			top_country = EXCLUDED.top_country
	`

	hourEnd := hour.Add(time.Hour)

	_, err := s.db.Exec(query,
		profileID, hour,
		profileID, hour, hourEnd,
		profileID, hour, hourEnd,
		hour, hourEnd,
		profileID, hour, hourEnd,
	)

	if err != nil {
		return fmt.Errorf("failed to aggregate hourly analytics: %w", err)
	}

	return nil
}

// GetSummary retrieves aggregated analytics summary for a profile
func (s *AnalyticsService) GetSummary(profileID string, startDate, endDate time.Time) (*model.AnalyticsSummary, error) {
	summary := &model.AnalyticsSummary{
		DeviceBreakdown: make(map[string]int),
	}

	// Get total views and unique visitors
	query := `
		SELECT
			COUNT(CASE WHEN event_type = 'view' THEN 1 END) as total_views,
			COUNT(DISTINCT CASE WHEN event_type = 'view' THEN ip_hash END) as unique_visitors,
			COUNT(CASE WHEN event_type = 'click' THEN 1 END) as total_clicks
		FROM analytics
		WHERE profile_id = ? AND created_at >= ? AND created_at < ?
	`

	err := s.db.QueryRow(query, profileID, startDate, endDate).Scan(
		&summary.TotalViews, &summary.UniqueVisitors, &summary.TotalClicks,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get summary stats: %w", err)
	}

	// Get average duration
	durationQuery := `
		SELECT COALESCE(AVG(duration_seconds), 0)
		FROM analytics_sessions
		WHERE profile_id = ? AND created_at >= ? AND created_at < ?
	`

	err = s.db.QueryRow(durationQuery, profileID, startDate, endDate).Scan(&summary.AvgDuration)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get average duration: %w", err)
	}

	// Get top referrers
	referrerQuery := `
		SELECT referrer, COUNT(*) as count
		FROM analytics
		WHERE profile_id = ? AND created_at >= ? AND created_at < ? AND referrer != ''
		GROUP BY referrer
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err := s.db.Query(referrerQuery, profileID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get referrers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat model.ReferrerStat
		if err := rows.Scan(&stat.Referrer, &stat.Count); err != nil {
			return nil, fmt.Errorf("failed to scan referrer: %w", err)
		}
		summary.TopReferrers = append(summary.TopReferrers, stat)
	}

	// Get top countries
	countryQuery := `
		SELECT country, COUNT(*) as count
		FROM analytics
		WHERE profile_id = ? AND created_at >= ? AND created_at < ? AND country != ''
		GROUP BY country
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err = s.db.Query(countryQuery, profileID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get countries: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat model.CountryStat
		if err := rows.Scan(&stat.Country, &stat.Count); err != nil {
			return nil, fmt.Errorf("failed to scan country: %w", err)
		}
		summary.TopCountries = append(summary.TopCountries, stat)
	}

	// Get device breakdown
	deviceQuery := `
		SELECT device_type, COUNT(*) as count
		FROM analytics
		WHERE profile_id = ? AND created_at >= ? AND created_at < ? AND device_type != ''
		GROUP BY device_type
	`

	rows, err = s.db.Query(deviceQuery, profileID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get device breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var deviceType string
		var count int
		if err := rows.Scan(&deviceType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		summary.DeviceBreakdown[deviceType] = count
	}

	// Get link click stats
	linkQuery := `
		SELECT l.id, l.title, COUNT(*) as click_count
		FROM analytics a
		JOIN links l ON a.link_id = l.id
		WHERE a.profile_id = ? AND a.event_type = 'click'
			AND a.created_at >= ? AND a.created_at < ?
		GROUP BY l.id, l.title
		ORDER BY click_count DESC
		LIMIT 20
	`

	rows, err = s.db.Query(linkQuery, profileID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get link stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat model.LinkClickStat
		if err := rows.Scan(&stat.LinkID, &stat.LinkTitle, &stat.ClickCount); err != nil {
			return nil, fmt.Errorf("failed to scan link stat: %w", err)
		}
		summary.LinkClickStats = append(summary.LinkClickStats, stat)
	}

	// Get hourly stats
	hourlyQuery := `
		SELECT *
		FROM analytics_hourly
		WHERE profile_id = ? AND hour >= ? AND hour < ?
		ORDER BY hour ASC
	`

	rows, err = s.db.Query(hourlyQuery, profileID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat model.AnalyticsHourly
		if err := rows.Scan(
			&stat.ProfileID, &stat.Hour, &stat.Views, &stat.UniqueVisitors,
			&stat.TotalClicks, &stat.AvgDurationSecs, &stat.TopReferrer,
			&stat.TopCountry,
		); err != nil {
			return nil, fmt.Errorf("failed to scan hourly stat: %w", err)
		}
		summary.HourlyStats = append(summary.HourlyStats, stat)
	}

	return summary, nil
}

// CleanupOldData removes analytics data older than retention period
func (s *AnalyticsService) CleanupOldData() error {
	// Get retention period from settings
	retentionDays, err := s.getRetentionDays()
	if err != nil {
		return fmt.Errorf("failed to get retention days: %w", err)
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Delete old analytics events
	query := `DELETE FROM analytics WHERE created_at < ?`
	_, err = s.db.Exec(query, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to delete old analytics: %w", err)
	}

	// Delete old sessions
	query = `DELETE FROM analytics_sessions WHERE created_at < ?`
	_, err = s.db.Exec(query, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to delete old sessions: %w", err)
	}

	return nil
}

// Helper functions

func (s *AnalyticsService) hashIP(ip string) string {
	// Add salt from settings (in production, retrieve from database)
	salt := "cassocial-analytics-salt"
	hash := sha256.Sum256([]byte(ip + salt))
	return hex.EncodeToString(hash[:])
}

func (s *AnalyticsService) parseDeviceType(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	// Simple device detection - in production use a proper library
	if strings.Contains(userAgent, "mobile") || strings.Contains(userAgent, "android") ||
		strings.Contains(userAgent, "iphone") {
		return model.DeviceTypeMobile
	}

	if strings.Contains(userAgent, "tablet") || strings.Contains(userAgent, "ipad") {
		return model.DeviceTypeTablet
	}

	return model.DeviceTypeDesktop
}

func (s *AnalyticsService) getCountryFromIP(ipAddress string) string {
	// Simplified - in production use a GeoIP database like MaxMind
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return "Unknown"
	}

	// Check if local IP
	if ip.IsLoopback() || ip.IsPrivate() {
		return "Local"
	}

	// Default to unknown - implement proper GeoIP lookup
	return "Unknown"
}

func (s *AnalyticsService) isAnalyticsEnabled(profileID string) (bool, error) {
	query := `SELECT analytics_enabled FROM profiles WHERE id = ?`

	var enabled bool
	err := s.db.QueryRow(query, profileID).Scan(&enabled)
	if err != nil {
		return false, fmt.Errorf("failed to check analytics status: %w", err)
	}

	return enabled, nil
}

func (s *AnalyticsService) sessionExists(sessionID string) (bool, error) {
	query := `SELECT COUNT(*) FROM analytics_sessions WHERE session_id = ?`

	var count int
	err := s.db.QueryRow(query, sessionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return count > 0, nil
}

func (s *AnalyticsService) generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *AnalyticsService) getRetentionDays() (int, error) {
	retentionStr, err := s.db.GetSetting("analytics_retention_days")
	if err != nil {
		return 90, nil // Default value
	}

	var retention int
	_, err = fmt.Sscanf(retentionStr, "%d", &retention)
	if err != nil {
		return 90, nil // Default value
	}

	return retention, nil
}
