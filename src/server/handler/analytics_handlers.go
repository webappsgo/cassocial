package handler

import (
	"net/http"
	"time"

	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
)

// AnalyticsHandlers handles analytics-related HTTP requests
type AnalyticsHandlers struct {
	db *store.DB
}

// NewAnalyticsHandlers creates a new AnalyticsHandlers instance
func NewAnalyticsHandlers(db *store.DB) *AnalyticsHandlers {
	return &AnalyticsHandlers{db: db}
}

// GetProfileAnalytics retrieves analytics for a profile
// GET /api/analytics/profile/{id}
func (h *AnalyticsHandlers) GetProfileAnalytics(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profileID := r.PathValue("id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "profile ID required")
		return
	}

	// Verify profile ownership
	if !h.userOwnsProfile(userID, profileID) {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get query parameters
	period := r.URL.Query().Get("period") // day, week, month, year, all
	if period == "" {
		period = "week"
	}

	// Calculate date range
	endDate := time.Now()
	var startDate time.Time
	switch period {
	case "day":
		startDate = endDate.AddDate(0, 0, -1)
	case "week":
		startDate = endDate.AddDate(0, 0, -7)
	case "month":
		startDate = endDate.AddDate(0, -1, 0)
	case "year":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		startDate = time.Time{} // All time
	}

	// Get total views
	totalViews := h.getTotalViews(profileID, startDate, endDate)

	// Get unique visitors
	uniqueVisitors := h.getUniqueVisitors(profileID, startDate, endDate)

	// Get total clicks
	totalClicks := h.getTotalClicks(profileID, startDate, endDate)

	// Get views by day
	viewsByDay := h.getViewsByDay(profileID, startDate, endDate)

	// Get top referrers
	topReferrers := h.getTopReferrers(profileID, startDate, endDate, 10)

	// Get device breakdown
	deviceBreakdown := h.getDeviceBreakdown(profileID, startDate, endDate)

	// Get country breakdown
	countryBreakdown := h.getCountryBreakdown(profileID, startDate, endDate, 10)

	analytics := map[string]interface{}{
		"period":            period,
		"start_date":        startDate,
		"end_date":          endDate,
		"total_views":       totalViews,
		"unique_visitors":   uniqueVisitors,
		"total_clicks":      totalClicks,
		"views_by_day":      viewsByDay,
		"top_referrers":     topReferrers,
		"device_breakdown":  deviceBreakdown,
		"country_breakdown": countryBreakdown,
	}

	respondJSON(w, http.StatusOK, analytics)
}

// GetLinkAnalytics retrieves analytics for links in a profile
// GET /api/analytics/links/{profile_id}
func (h *AnalyticsHandlers) GetLinkAnalytics(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profileID := r.PathValue("profile_id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "profile ID required")
		return
	}

	// Verify profile ownership
	if !h.userOwnsProfile(userID, profileID) {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get link click statistics
	query := `SELECT l.id, COALESCE(l.title,''), COALESCE(l.url,''), l.click_count,
			  COALESCE(COUNT(a.id), 0) as recent_clicks
			  FROM links l
			  LEFT JOIN analytics a ON a.link_id = l.id AND a.event_type = 'click'
			    AND a.created_at >= ?
			  WHERE l.profile_id = ?
			  GROUP BY l.id, l.title, l.url, l.click_count
			  ORDER BY l.click_count DESC`

	if h.db.Driver == "postgres" {
		query = `SELECT l.id, COALESCE(l.title,''), COALESCE(l.url,''), l.click_count,
				 COALESCE(COUNT(a.id), 0) as recent_clicks
				 FROM links l
				 LEFT JOIN analytics a ON a.link_id = l.id AND a.event_type = 'click'
				   AND a.created_at >= $1
				 WHERE l.profile_id = $2
				 GROUP BY l.id, l.title, l.url, l.click_count
				 ORDER BY l.click_count DESC`
	}

	// Last 30 days
	startDate := time.Now().AddDate(0, 0, -30)

	rows, err := h.db.Query(query, h.db.BindTime(startDate), profileID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch link analytics")
		return
	}
	defer rows.Close()

	links := []map[string]interface{}{}
	for rows.Next() {
		var id, title, url string
		var totalClicks, recentClicks int
		err := rows.Scan(&id, &title, &url, &totalClicks, &recentClicks)
		if err != nil {
			continue
		}
		links = append(links, map[string]interface{}{
			"id":            id,
			"title":         title,
			"url":           url,
			"total_clicks":  totalClicks,
			"recent_clicks": recentClicks,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"links":      links,
		"start_date": startDate,
		"end_date":   time.Now(),
	})
}

// ExportAnalytics exports analytics data
// GET /api/analytics/export/{profile_id}?format={csv|json}
func (h *AnalyticsHandlers) ExportAnalytics(w http.ResponseWriter, r *http.Request) {
	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profileID := r.PathValue("profile_id")
	if profileID == "" {
		respondError(w, http.StatusBadRequest, "profile ID required")
		return
	}

	// Verify profile ownership
	if !h.userOwnsProfile(userID, profileID) {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "export not yet implemented",
		"format":  format,
	})
}

// Helper functions

func (h *AnalyticsHandlers) getTotalViews(profileID string, startDate, endDate time.Time) int {
	var count int
	query := `SELECT COUNT(*) FROM analytics
			  WHERE profile_id = ? AND event_type = 'view'
			  AND created_at BETWEEN ? AND ?`

	if h.db.Driver == "postgres" {
		query = `SELECT COUNT(*) FROM analytics
				 WHERE profile_id = $1 AND event_type = 'view'
				 AND created_at BETWEEN $2 AND $3`
	}

	h.db.QueryRow(query, profileID, h.db.BindTime(startDate), h.db.BindTime(endDate)).Scan(&count)
	return count
}

func (h *AnalyticsHandlers) getUniqueVisitors(profileID string, startDate, endDate time.Time) int {
	var count int
	query := `SELECT COUNT(DISTINCT ip_hash) FROM analytics
			  WHERE profile_id = ? AND event_type = 'view'
			  AND created_at BETWEEN ? AND ?`

	if h.db.Driver == "postgres" {
		query = `SELECT COUNT(DISTINCT ip_hash) FROM analytics
				 WHERE profile_id = $1 AND event_type = 'view'
				 AND created_at BETWEEN $2 AND $3`
	}

	h.db.QueryRow(query, profileID, h.db.BindTime(startDate), h.db.BindTime(endDate)).Scan(&count)
	return count
}

func (h *AnalyticsHandlers) getTotalClicks(profileID string, startDate, endDate time.Time) int {
	var count int
	query := `SELECT COUNT(*) FROM analytics
			  WHERE profile_id = ? AND event_type = 'click'
			  AND created_at BETWEEN ? AND ?`

	if h.db.Driver == "postgres" {
		query = `SELECT COUNT(*) FROM analytics
				 WHERE profile_id = $1 AND event_type = 'click'
				 AND created_at BETWEEN $2 AND $3`
	}

	h.db.QueryRow(query, profileID, h.db.BindTime(startDate), h.db.BindTime(endDate)).Scan(&count)
	return count
}

func (h *AnalyticsHandlers) getViewsByDay(profileID string, startDate, endDate time.Time) []map[string]interface{} {
	query := `SELECT DATE(created_at) as date, COUNT(*) as views
			  FROM analytics
			  WHERE profile_id = ? AND event_type = 'view'
			  AND created_at BETWEEN ? AND ?
			  GROUP BY DATE(created_at)
			  ORDER BY date ASC`

	if h.db.Driver == "postgres" {
		query = `SELECT DATE(created_at) as date, COUNT(*) as views
				 FROM analytics
				 WHERE profile_id = $1 AND event_type = 'view'
				 AND created_at BETWEEN $2 AND $3
				 GROUP BY DATE(created_at)
				 ORDER BY date ASC`
	}

	rows, err := h.db.Query(query, profileID, h.db.BindTime(startDate), h.db.BindTime(endDate))
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	result := []map[string]interface{}{}
	for rows.Next() {
		var date time.Time
		var views int
		rows.Scan(&date, &views)
		result = append(result, map[string]interface{}{
			"date":  date.Format("2006-01-02"),
			"views": views,
		})
	}

	return result
}

func (h *AnalyticsHandlers) getTopReferrers(profileID string, startDate, endDate time.Time, limit int) []map[string]interface{} {
	query := `SELECT referrer, COUNT(*) as count
			  FROM analytics
			  WHERE profile_id = ? AND event_type = 'view'
			  AND referrer IS NOT NULL AND referrer != ''
			  AND created_at BETWEEN ? AND ?
			  GROUP BY referrer
			  ORDER BY count DESC LIMIT ?`

	if h.db.Driver == "postgres" {
		query = `SELECT referrer, COUNT(*) as count
				 FROM analytics
				 WHERE profile_id = $1 AND event_type = 'view'
				 AND referrer IS NOT NULL AND referrer != ''
				 AND created_at BETWEEN $2 AND $3
				 GROUP BY referrer
				 ORDER BY count DESC LIMIT $4`
	}

	rows, err := h.db.Query(query, profileID, h.db.BindTime(startDate), h.db.BindTime(endDate), limit)
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	result := []map[string]interface{}{}
	for rows.Next() {
		var referrer string
		var count int
		rows.Scan(&referrer, &count)
		result = append(result, map[string]interface{}{
			"referrer": referrer,
			"count":    count,
		})
	}

	return result
}

func (h *AnalyticsHandlers) getDeviceBreakdown(profileID string, startDate, endDate time.Time) map[string]int {
	query := `SELECT device_type, COUNT(*) as count
			  FROM analytics
			  WHERE profile_id = ? AND event_type = 'view'
			  AND created_at BETWEEN ? AND ?
			  GROUP BY device_type`

	if h.db.Driver == "postgres" {
		query = `SELECT device_type, COUNT(*) as count
				 FROM analytics
				 WHERE profile_id = $1 AND event_type = 'view'
				 AND created_at BETWEEN $2 AND $3
				 GROUP BY device_type`
	}

	rows, err := h.db.Query(query, profileID, h.db.BindTime(startDate), h.db.BindTime(endDate))
	if err != nil {
		return map[string]int{}
	}
	defer rows.Close()

	result := map[string]int{
		"mobile":  0,
		"tablet":  0,
		"desktop": 0,
	}

	for rows.Next() {
		var deviceType string
		var count int
		rows.Scan(&deviceType, &count)
		result[deviceType] = count
	}

	return result
}

func (h *AnalyticsHandlers) getCountryBreakdown(profileID string, startDate, endDate time.Time, limit int) []map[string]interface{} {
	query := `SELECT country, COUNT(*) as count
			  FROM analytics
			  WHERE profile_id = ? AND event_type = 'view'
			  AND country IS NOT NULL AND country != ''
			  AND created_at BETWEEN ? AND ?
			  GROUP BY country
			  ORDER BY count DESC LIMIT ?`

	if h.db.Driver == "postgres" {
		query = `SELECT country, COUNT(*) as count
				 FROM analytics
				 WHERE profile_id = $1 AND event_type = 'view'
				 AND country IS NOT NULL AND country != ''
				 AND created_at BETWEEN $2 AND $3
				 GROUP BY country
				 ORDER BY count DESC LIMIT $4`
	}

	rows, err := h.db.Query(query, profileID, h.db.BindTime(startDate), h.db.BindTime(endDate), limit)
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	result := []map[string]interface{}{}
	for rows.Next() {
		var country string
		var count int
		rows.Scan(&country, &count)
		result = append(result, map[string]interface{}{
			"country": country,
			"count":   count,
		})
	}

	return result
}

func (h *AnalyticsHandlers) userOwnsProfile(userID, profileID string) bool {
	var count int
	query := "SELECT COUNT(*) FROM profiles WHERE id = ? AND user_id = ?"
	if h.db.Driver == "postgres" {
		query = "SELECT COUNT(*) FROM profiles WHERE id = $1 AND user_id = $2"
	}
	h.db.QueryRow(query, profileID, userID).Scan(&count)
	return count > 0
}
