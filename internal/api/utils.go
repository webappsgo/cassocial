package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError sends a JSON error response
func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	})
}

// generateUUID generates a simple UUID
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// replaceQuestionMarks replaces ? with $1, $2, etc. for PostgreSQL
func replaceQuestionMarks(query string, count int) string {
	result := query
	for i := count; i >= 1; i-- {
		result = strings.Replace(result, "?", fmt.Sprintf("$%d", i), 1)
	}
	return result
}

// replaceQuestionMarksWithArgs replaces ? with $1, $2, etc. based on argument count
func replaceQuestionMarksWithArgs(query string, argCount int) string {
	result := ""
	paramIndex := 1
	for _, char := range query {
		if char == '?' {
			result += fmt.Sprintf("$%d", paramIndex)
			paramIndex++
		} else {
			result += string(char)
		}
	}
	return result
}

// getIPAddress extracts the real IP address from the request
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Use RemoteAddr as fallback
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// hashIP creates a SHA-256 hash of an IP address for privacy
func hashIP(ip string) string {
	hash := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(hash[:])
}

// detectDeviceType detects the device type from user agent
func detectDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)

	// Mobile devices
	mobileKeywords := []string{
		"mobile", "android", "iphone", "ipod", "blackberry",
		"windows phone", "palm", "smartphone",
	}
	for _, keyword := range mobileKeywords {
		if strings.Contains(ua, keyword) {
			return "mobile"
		}
	}

	// Tablets
	tabletKeywords := []string{"ipad", "tablet", "kindle"}
	for _, keyword := range tabletKeywords {
		if strings.Contains(ua, keyword) {
			return "tablet"
		}
	}

	// Default to desktop
	return "desktop"
}

// getCurrentTimestamp returns the current timestamp
func getCurrentTimestamp() time.Time {
	return time.Now()
}

// validateEmail validates email format
func validateEmail(email string) bool {
	// Simple email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// validateURL validates URL format
func validateURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// sanitizeString removes dangerous characters from a string
func sanitizeString(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")
	// Trim whitespace
	return strings.TrimSpace(s)
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength]
}

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

// stringPtr returns a pointer to a string value
func stringPtr(s string) *string {
	return &s
}

// intPtr returns a pointer to an int value
func intPtr(i int) *int {
	return &i
}

// timePtr returns a pointer to a time.Time value
func timePtr(t time.Time) *time.Time {
	return &t
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// unique removes duplicates from a string slice
func unique(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// parseBool parses a string to a boolean
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes" || s == "on"
}

// parseInt parses a string to an integer with a default value
func parseInt(s string, defaultValue int) int {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return defaultValue
	}
	return i
}

// parseFloat parses a string to a float64 with a default value
func parseFloat(s string, defaultValue float64) float64 {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		return defaultValue
	}
	return f
}

// JSONResponse represents a standard JSON response structure
type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewSuccessResponse creates a successful JSON response
func NewSuccessResponse(message string, data interface{}) JSONResponse {
	return JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates an error JSON response
func NewErrorResponse(error string) JSONResponse {
	return JSONResponse{
		Success: false,
		Error:   error,
	}
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"offset"`
}

// ParsePagination parses pagination parameters from query string
func ParsePagination(r *http.Request) PaginationParams {
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("page_size"), 20)

	// Ensure page is at least 1
	if page < 1 {
		page = 1
	}

	// Ensure page size is reasonable
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
	}
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int         `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginatedResponse creates a paginated response
func NewPaginatedResponse(data interface{}, page, pageSize, totalItems int) PaginatedResponse {
	totalPages := (totalItems + pageSize - 1) / pageSize
	return PaginatedResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
