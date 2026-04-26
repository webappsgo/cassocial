package model

import (
	"errors"
	"time"
)

// Analytics represents an analytics event
type Analytics struct {
	ID         string    `json:"id" db:"id"`
	ProfileID  string    `json:"profile_id" db:"profile_id"`
	LinkID     string    `json:"link_id,omitempty" db:"link_id"`
	EventType  string    `json:"event_type" db:"event_type"`
	IPHash     string    `json:"ip_hash,omitempty" db:"ip_hash"`
	UserAgent  string    `json:"user_agent,omitempty" db:"user_agent"`
	Referrer   string    `json:"referrer,omitempty" db:"referrer"`
	Country    string    `json:"country,omitempty" db:"country"`
	DeviceType string    `json:"device_type,omitempty" db:"device_type"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// AnalyticsSession represents a user session
type AnalyticsSession struct {
	ID              string       `json:"id" db:"id"`
	ProfileID       string       `json:"profile_id" db:"profile_id"`
	SessionID       string       `json:"session_id" db:"session_id"`
	IPHash          string       `json:"ip_hash" db:"ip_hash"`
	Country         string       `json:"country,omitempty" db:"country"`
	Region          string       `json:"region,omitempty" db:"region"`
	City            string       `json:"city,omitempty" db:"city"`
	DeviceType      string       `json:"device_type,omitempty" db:"device_type"`
	Browser         string       `json:"browser,omitempty" db:"browser"`
	OS              string       `json:"os,omitempty" db:"os"`
	ReferrerDomain  string       `json:"referrer_domain,omitempty" db:"referrer_domain"`
	ReferrerPath    string       `json:"referrer_path,omitempty" db:"referrer_path"`
	UTMSource       string       `json:"utm_source,omitempty" db:"utm_source"`
	UTMMedium       string       `json:"utm_medium,omitempty" db:"utm_medium"`
	UTMCampaign     string       `json:"utm_campaign,omitempty" db:"utm_campaign"`
	LandingPage     string       `json:"landing_page,omitempty" db:"landing_page"`
	DurationSeconds int          `json:"duration_seconds" db:"duration_seconds"`
	LinkClicks      int          `json:"link_clicks" db:"link_clicks"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
}

// AnalyticsHourly represents aggregated hourly analytics
type AnalyticsHourly struct {
	ProfileID        string    `json:"profile_id" db:"profile_id"`
	Hour             time.Time `json:"hour" db:"hour"`
	Views            int       `json:"views" db:"views"`
	UniqueVisitors   int       `json:"unique_visitors" db:"unique_visitors"`
	TotalClicks      int       `json:"total_clicks" db:"total_clicks"`
	AvgDurationSecs  int       `json:"avg_duration_seconds" db:"avg_duration_seconds"`
	TopReferrer      string    `json:"top_referrer,omitempty" db:"top_referrer"`
	TopCountry       string    `json:"top_country,omitempty" db:"top_country"`
}

// Valid event types
const (
	EventTypeView  = "view"
	EventTypeClick = "click"
)

// Valid device types
const (
	DeviceTypeMobile  = "mobile"
	DeviceTypeTablet  = "tablet"
	DeviceTypeDesktop = "desktop"
)

var (
	ErrInvalidEventType  = errors.New("invalid event type")
	ErrInvalidDeviceType = errors.New("invalid device type")
)

// Validate validates the analytics model
func (a *Analytics) Validate() error {
	// Validate event type
	if a.EventType != EventTypeView && a.EventType != EventTypeClick {
		return ErrInvalidEventType
	}

	// Validate device type if provided
	if a.DeviceType != "" {
		if a.DeviceType != DeviceTypeMobile &&
		   a.DeviceType != DeviceTypeTablet &&
		   a.DeviceType != DeviceTypeDesktop {
			return ErrInvalidDeviceType
		}
	}

	return nil
}

// IsView checks if the event is a view event
func (a *Analytics) IsView() bool {
	return a.EventType == EventTypeView
}

// IsClick checks if the event is a click event
func (a *Analytics) IsClick() bool {
	return a.EventType == EventTypeClick
}

// Validate validates the analytics session model
func (as *AnalyticsSession) Validate() error {
	// Validate device type if provided
	if as.DeviceType != "" {
		if as.DeviceType != DeviceTypeMobile &&
		   as.DeviceType != DeviceTypeTablet &&
		   as.DeviceType != DeviceTypeDesktop {
			return ErrInvalidDeviceType
		}
	}

	return nil
}

// IncrementLinkClicks increments the link clicks counter
func (as *AnalyticsSession) IncrementLinkClicks() {
	as.LinkClicks++
}

// UpdateDuration updates the session duration
func (as *AnalyticsSession) UpdateDuration(seconds int) {
	as.DurationSeconds = seconds
}

// AnalyticsSummary represents aggregated analytics data
type AnalyticsSummary struct {
	TotalViews        int                    `json:"total_views"`
	UniqueVisitors    int                    `json:"unique_visitors"`
	TotalClicks       int                    `json:"total_clicks"`
	AvgDuration       int                    `json:"avg_duration_seconds"`
	TopReferrers      []ReferrerStat         `json:"top_referrers"`
	TopCountries      []CountryStat          `json:"top_countries"`
	DeviceBreakdown   map[string]int         `json:"device_breakdown"`
	LinkClickStats    []LinkClickStat        `json:"link_click_stats"`
	HourlyStats       []AnalyticsHourly      `json:"hourly_stats"`
}

// ReferrerStat represents referrer statistics
type ReferrerStat struct {
	Referrer string `json:"referrer"`
	Count    int    `json:"count"`
}

// CountryStat represents country statistics
type CountryStat struct {
	Country string `json:"country"`
	Count   int    `json:"count"`
}

// LinkClickStat represents link click statistics
type LinkClickStat struct {
	LinkID     string `json:"link_id"`
	LinkTitle  string `json:"link_title"`
	ClickCount int    `json:"click_count"`
}
