package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/casapps/cassocial/internal/database"
	"github.com/casapps/cassocial/internal/models"
)

var (
	ErrUnsupportedExportFormat = errors.New("unsupported export format")
	ErrProfileAccessDenied     = errors.New("profile access denied")
)

// ExportService handles exporting profile data to various formats
type ExportService struct {
	db             *database.DB
	profileService *ProfileService
	linkService    *LinkService
}

// NewExportService creates a new export service
func NewExportService(db *database.DB, profileService *ProfileService, linkService *LinkService) *ExportService {
	return &ExportService{
		db:             db,
		profileService: profileService,
		linkService:    linkService,
	}
}

// ExportProfile exports a profile in the specified format
func (s *ExportService) ExportProfile(profileID, userID, format string) ([]byte, string, error) {
	// Get profile
	profile, err := s.profileService.GetByID(profileID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get profile: %w", err)
	}

	// Check ownership
	if profile.UserID != userID {
		return nil, "", ErrProfileAccessDenied
	}

	// Get links
	links, err := s.linkService.GetByProfileID(profileID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get links: %w", err)
	}

	// Export based on format
	switch format {
	case "json":
		return s.exportToJSON(profile, links)
	case "csv":
		return s.exportToCSV(profile, links)
	case "html":
		return s.exportToHTML(profile, links)
	case "pdf":
		return s.exportToPDF(profile, links)
	case "vcard":
		return s.exportToVCard(profile, links)
	default:
		return nil, "", ErrUnsupportedExportFormat
	}
}

// ExportToJSON exports profile to JSON format
func (s *ExportService) exportToJSON(profile *models.Profile, links []*models.Link) ([]byte, string, error) {
	data := map[string]interface{}{
		"profile": map[string]interface{}{
			"slug":              profile.Slug,
			"display_name":      profile.DisplayName,
			"bio":               profile.Bio,
			"avatar_url":        profile.AvatarURL,
			"header_image_url":  profile.HeaderImageURL,
			"show_usernames":    profile.ShowUsernames,
			"is_public":         profile.IsPublic,
			"analytics_enabled": profile.AnalyticsEnabled,
			"meta_title":        profile.MetaTitle,
			"meta_description":  profile.MetaDescription,
			"created_at":        profile.CreatedAt,
			"updated_at":        profile.UpdatedAt,
		},
		"links": []map[string]interface{}{},
		"exported_at": time.Now(),
		"version":     "1.0.0",
	}

	// Add links
	linkList := make([]map[string]interface{}, 0, len(links))
	for _, link := range links {
		linkList = append(linkList, map[string]interface{}{
			"title":            link.Title,
			"username":         link.Username,
			"url":              link.URL,
			"icon_url":         link.IconURL,
			"background_color": link.BackgroundColor,
			"text_color":       link.TextColor,
			"position":         link.Position,
			"is_active":        link.IsActive,
			"click_count":      link.ClickCount,
		})
	}
	data["links"] = linkList

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	filename := fmt.Sprintf("%s-export.json", profile.Slug)
	return jsonData, filename, nil
}

// ExportToCSV exports profile links to CSV format
func (s *ExportService) exportToCSV(profile *models.Profile, links []*models.Link) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"Title", "URL", "Username", "Position", "Active", "Click Count"}
	if err := writer.Write(header); err != nil {
		return nil, "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write links
	for _, link := range links {
		record := []string{
			link.Title,
			link.URL,
			link.Username,
			fmt.Sprintf("%d", link.Position),
			fmt.Sprintf("%t", link.IsActive),
			fmt.Sprintf("%d", link.ClickCount),
		}

		if err := writer.Write(record); err != nil {
			return nil, "", fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", fmt.Errorf("CSV writer error: %w", err)
	}

	filename := fmt.Sprintf("%s-links.csv", profile.Slug)
	return buf.Bytes(), filename, nil
}

// ExportToHTML exports profile to standalone HTML page
func (s *ExportService) exportToHTML(profile *models.Profile, links []*models.Link) ([]byte, string, error) {
	var html strings.Builder

	// Build HTML
	html.WriteString("<!DOCTYPE html>\n")
	html.WriteString("<html lang=\"en\">\n")
	html.WriteString("<head>\n")
	html.WriteString(fmt.Sprintf("  <meta charset=\"UTF-8\">\n"))
	html.WriteString(fmt.Sprintf("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n"))
	html.WriteString(fmt.Sprintf("  <title>%s</title>\n", escapeHTML(profile.DisplayName)))

	if profile.MetaDescription != "" {
		html.WriteString(fmt.Sprintf("  <meta name=\"description\" content=\"%s\">\n", escapeHTML(profile.MetaDescription)))
	}

	// Add basic styles
	html.WriteString("  <style>\n")
	html.WriteString("    * { margin: 0; padding: 0; box-sizing: border-box; }\n")
	html.WriteString("    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #282a36; color: #f8f8f2; padding: 2rem; }\n")
	html.WriteString("    .container { max-width: 680px; margin: 0 auto; }\n")
	html.WriteString("    .profile { text-align: center; margin-bottom: 2rem; }\n")
	html.WriteString("    .avatar { width: 120px; height: 120px; border-radius: 50%; margin-bottom: 1rem; }\n")
	html.WriteString("    .display-name { font-size: 2rem; font-weight: bold; margin-bottom: 0.5rem; }\n")
	html.WriteString("    .bio { color: #6272a4; margin-bottom: 2rem; }\n")
	html.WriteString("    .links { display: flex; flex-direction: column; gap: 1rem; }\n")
	html.WriteString("    .link { background: #44475a; padding: 1rem; border-radius: 12px; text-decoration: none; color: inherit; display: block; transition: transform 0.2s; }\n")
	html.WriteString("    .link:hover { transform: translateY(-2px); background: #6272a4; }\n")
	html.WriteString("    .link-title { font-weight: 500; }\n")
	html.WriteString("    .footer { text-align: center; margin-top: 2rem; color: #6272a4; font-size: 0.875rem; }\n")
	html.WriteString("  </style>\n")
	html.WriteString("</head>\n")
	html.WriteString("<body>\n")
	html.WriteString("  <div class=\"container\">\n")

	// Profile section
	html.WriteString("    <div class=\"profile\">\n")
	if profile.AvatarURL != "" {
		html.WriteString(fmt.Sprintf("      <img src=\"%s\" alt=\"%s\" class=\"avatar\">\n", escapeHTML(profile.AvatarURL), escapeHTML(profile.DisplayName)))
	}
	html.WriteString(fmt.Sprintf("      <h1 class=\"display-name\">%s</h1>\n", escapeHTML(profile.DisplayName)))
	if profile.Bio != "" {
		html.WriteString(fmt.Sprintf("      <p class=\"bio\">%s</p>\n", escapeHTML(profile.Bio)))
	}
	html.WriteString("    </div>\n")

	// Links section
	html.WriteString("    <div class=\"links\">\n")
	for _, link := range links {
		if !link.IsActive {
			continue
		}

		html.WriteString(fmt.Sprintf("      <a href=\"%s\" class=\"link\" target=\"_blank\" rel=\"noopener noreferrer\">\n", escapeHTML(link.URL)))
		html.WriteString(fmt.Sprintf("        <div class=\"link-title\">%s</div>\n", escapeHTML(link.Title)))
		html.WriteString("      </a>\n")
	}
	html.WriteString("    </div>\n")

	// Footer
	html.WriteString("    <div class=\"footer\">\n")
	html.WriteString(fmt.Sprintf("      <p>Exported from Cassocial on %s</p>\n", time.Now().Format("January 2, 2006")))
	html.WriteString("    </div>\n")

	html.WriteString("  </div>\n")
	html.WriteString("</body>\n")
	html.WriteString("</html>\n")

	filename := fmt.Sprintf("%s.html", profile.Slug)
	return []byte(html.String()), filename, nil
}

// ExportToPDF exports profile to PDF format
func (s *ExportService) exportToPDF(profile *models.Profile, links []*models.Link) ([]byte, string, error) {
	// In production, use a library like gofpdf to generate actual PDF
	// This is a placeholder implementation

	var pdf strings.Builder
	pdf.WriteString("%PDF-1.4\n")
	pdf.WriteString("%Profile Export\n")
	pdf.WriteString(fmt.Sprintf("Profile: %s\n", profile.DisplayName))
	pdf.WriteString(fmt.Sprintf("Bio: %s\n\n", profile.Bio))
	pdf.WriteString("Links:\n")

	for _, link := range links {
		if link.IsActive {
			pdf.WriteString(fmt.Sprintf("- %s: %s\n", link.Title, link.URL))
		}
	}

	filename := fmt.Sprintf("%s.pdf", profile.Slug)
	return []byte(pdf.String()), filename, nil
}

// ExportToVCard exports profile to vCard format (VCF)
func (s *ExportService) exportToVCard(profile *models.Profile, links []*models.Link) ([]byte, string, error) {
	var vcard strings.Builder

	// vCard 3.0 format
	vcard.WriteString("BEGIN:VCARD\n")
	vcard.WriteString("VERSION:3.0\n")

	// Full name
	vcard.WriteString(fmt.Sprintf("FN:%s\n", profile.DisplayName))

	// Structured name (simplified)
	vcard.WriteString(fmt.Sprintf("N:%s;;;;\n", profile.DisplayName))

	// Nickname/username
	vcard.WriteString(fmt.Sprintf("NICKNAME:%s\n", profile.Slug))

	// Bio as note
	if profile.Bio != "" {
		vcard.WriteString(fmt.Sprintf("NOTE:%s\n", profile.Bio))
	}

	// Avatar photo
	if profile.AvatarURL != "" {
		vcard.WriteString(fmt.Sprintf("PHOTO;VALUE=URI:%s\n", profile.AvatarURL))
	}

	// Add URLs from links
	for _, link := range links {
		if link.IsActive {
			// Add as URL with label
			vcard.WriteString(fmt.Sprintf("URL;TYPE=%s:%s\n", link.Title, link.URL))
		}
	}

	// Profile URL
	vcard.WriteString(fmt.Sprintf("URL;TYPE=PROFILE:https://casjay.link/%s\n", profile.Slug))

	// Timestamp
	vcard.WriteString(fmt.Sprintf("REV:%s\n", time.Now().Format("2006-01-02T15:04:05Z")))

	vcard.WriteString("END:VCARD\n")

	filename := fmt.Sprintf("%s.vcf", profile.Slug)
	return []byte(vcard.String()), filename, nil
}

// ExportAnalytics exports analytics data for a profile
func (s *ExportService) ExportAnalytics(profileID, userID string, startDate, endDate time.Time, format string) ([]byte, string, error) {
	// Get profile to check ownership
	profile, err := s.profileService.GetByID(profileID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get profile: %w", err)
	}

	if profile.UserID != userID {
		return nil, "", ErrProfileAccessDenied
	}

	// Query analytics data
	query := `
		SELECT event_type, COUNT(*) as count, DATE(created_at) as date
		FROM analytics
		WHERE profile_id = ? AND created_at >= ? AND created_at < ?
		GROUP BY event_type, DATE(created_at)
		ORDER BY date DESC
	`

	rows, err := s.db.Query(query, profileID, startDate, endDate)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query analytics: %w", err)
	}
	defer rows.Close()

	type AnalyticsRecord struct {
		EventType string
		Count     int
		Date      string
	}

	var records []AnalyticsRecord
	for rows.Next() {
		var record AnalyticsRecord
		if err := rows.Scan(&record.EventType, &record.Count, &record.Date); err != nil {
			continue
		}
		records = append(records, record)
	}

	// Export based on format
	switch format {
	case "json":
		return s.exportAnalyticsJSON(profile, records)
	case "csv":
		return s.exportAnalyticsCSV(profile, records)
	default:
		return nil, "", ErrUnsupportedExportFormat
	}
}

func (s *ExportService) exportAnalyticsJSON(profile *models.Profile, records []interface{}) ([]byte, string, error) {
	data := map[string]interface{}{
		"profile_id":  profile.ID,
		"profile_slug": profile.Slug,
		"analytics":   records,
		"exported_at": time.Now(),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	filename := fmt.Sprintf("%s-analytics.json", profile.Slug)
	return jsonData, filename, nil
}

func (s *ExportService) exportAnalyticsCSV(profile *models.Profile, records []interface{}) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"Date", "Event Type", "Count"}
	if err := writer.Write(header); err != nil {
		return nil, "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write records (simplified - in production, properly type assert)
	for _, record := range records {
		// This would need proper implementation based on actual record type
		writer.Write([]string{"", "", ""})
	}

	writer.Flush()

	filename := fmt.Sprintf("%s-analytics.csv", profile.Slug)
	return buf.Bytes(), filename, nil
}

// Helper function to escape HTML entities
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
