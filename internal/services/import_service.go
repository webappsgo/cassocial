package services

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/casapps/cassocial/internal/database"
	"github.com/casapps/cassocial/internal/models"
)

var (
	ErrUnsupportedImportSource = errors.New("unsupported import source")
	ErrInvalidImportData       = errors.New("invalid import data format")
	ErrImportJobNotFound       = errors.New("import job not found")
)

// ImportService handles importing data from various sources
type ImportService struct {
	db             *database.DB
	profileService *ProfileService
	linkService    *LinkService
}

// NewImportService creates a new import service
func NewImportService(db *database.DB, profileService *ProfileService, linkService *LinkService) *ImportService {
	return &ImportService{
		db:             db,
		profileService: profileService,
		linkService:    linkService,
	}
}

// ImportData imports data from various sources
func (s *ImportService) ImportData(userID, source string, data []byte) (string, error) {
	// Create import job
	jobID, err := s.createImportJob(userID, source)
	if err != nil {
		return "", fmt.Errorf("failed to create import job: %w", err)
	}

	// Update job status to processing
	if err := s.updateJobStatus(jobID, "processing", nil); err != nil {
		return "", fmt.Errorf("failed to update job status: %w", err)
	}

	// Import based on source
	var result map[string]interface{}
	var importErr error

	switch source {
	case "linktree":
		result, importErr = s.importFromLinktree(userID, data)
	case "linkstack":
		result, importErr = s.importFromLinkstack(userID, data)
	case "carrd":
		result, importErr = s.importFromCarrd(userID, data)
	case "aboutme":
		result, importErr = s.importFromAboutMe(userID, data)
	case "csv":
		result, importErr = s.importFromCSV(userID, data)
	case "json":
		result, importErr = s.importFromJSON(userID, data)
	default:
		importErr = ErrUnsupportedImportSource
	}

	// Update job with result
	if importErr != nil {
		s.updateJobStatus(jobID, "failed", map[string]interface{}{"error": importErr.Error()})
		return jobID, importErr
	}

	if err := s.updateJobStatus(jobID, "completed", result); err != nil {
		return jobID, fmt.Errorf("failed to update job result: %w", err)
	}

	return jobID, nil
}

// GetImportJob retrieves an import job by ID
func (s *ImportService) GetImportJob(jobID string) (map[string]interface{}, error) {
	query := `SELECT id, user_id, source, status, result, created_at, completed_at FROM import_jobs WHERE id = ?`

	var job struct {
		ID          string
		UserID      string
		Source      string
		Status      string
		Result      string
		CreatedAt   time.Time
		CompletedAt *time.Time
	}

	err := s.db.QueryRow(query, jobID).Scan(
		&job.ID, &job.UserID, &job.Source, &job.Status,
		&job.Result, &job.CreatedAt, &job.CompletedAt,
	)

	if err != nil {
		return nil, ErrImportJobNotFound
	}

	// Parse result JSON
	var result map[string]interface{}
	if job.Result != "" {
		json.Unmarshal([]byte(job.Result), &result)
	}

	return map[string]interface{}{
		"id":           job.ID,
		"user_id":      job.UserID,
		"source":       job.Source,
		"status":       job.Status,
		"result":       result,
		"created_at":   job.CreatedAt,
		"completed_at": job.CompletedAt,
	}, nil
}

// Import implementations for different sources

func (s *ImportService) importFromLinktree(userID string, data []byte) (map[string]interface{}, error) {
	// Parse Linktree JSON export
	var linktreeData struct {
		AccountData struct {
			Username    string `json:"username"`
			DisplayName string `json:"displayName"`
			Bio         string `json:"bio"`
			AvatarURL   string `json:"profilePictureUrl"`
		} `json:"accountData"`
		Links []struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		} `json:"links"`
	}

	if err := json.Unmarshal(data, &linktreeData); err != nil {
		return nil, fmt.Errorf("failed to parse Linktree data: %w", err)
	}

	// Create profile
	profile := &models.Profile{
		Slug:        linktreeData.AccountData.Username,
		DisplayName: linktreeData.AccountData.DisplayName,
		Bio:         linktreeData.AccountData.Bio,
		AvatarURL:   linktreeData.AccountData.AvatarURL,
	}

	if err := s.profileService.Create(userID, profile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Import links
	imported := 0
	for i, linkData := range linktreeData.Links {
		link := &models.Link{
			ProfileID: profile.ID,
			Title:     linkData.Title,
			URL:       linkData.URL,
			Position:  i + 1,
		}

		if err := s.linkService.Create(link); err != nil {
			continue // Skip invalid links
		}
		imported++
	}

	return map[string]interface{}{
		"profile_id":    profile.ID,
		"links_imported": imported,
		"total_links":   len(linktreeData.Links),
	}, nil
}

func (s *ImportService) importFromLinkstack(userID string, data []byte) (map[string]interface{}, error) {
	// Parse Linkstack JSON export
	var linkstackData struct {
		Profile struct {
			Username string `json:"username"`
			Name     string `json:"name"`
			Bio      string `json:"bio"`
			Avatar   string `json:"avatar"`
		} `json:"profile"`
		Links []struct {
			Title string `json:"title"`
			URL   string `json:"url"`
			Order int    `json:"order"`
		} `json:"links"`
	}

	if err := json.Unmarshal(data, &linkstackData); err != nil {
		return nil, fmt.Errorf("failed to parse Linkstack data: %w", err)
	}

	// Create profile
	profile := &models.Profile{
		Slug:        linkstackData.Profile.Username,
		DisplayName: linkstackData.Profile.Name,
		Bio:         linkstackData.Profile.Bio,
		AvatarURL:   linkstackData.Profile.Avatar,
	}

	if err := s.profileService.Create(userID, profile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Import links
	imported := 0
	for _, linkData := range linkstackData.Links {
		link := &models.Link{
			ProfileID: profile.ID,
			Title:     linkData.Title,
			URL:       linkData.URL,
			Position:  linkData.Order,
		}

		if err := s.linkService.Create(link); err != nil {
			continue // Skip invalid links
		}
		imported++
	}

	return map[string]interface{}{
		"profile_id":    profile.ID,
		"links_imported": imported,
		"total_links":   len(linkstackData.Links),
	}, nil
}

func (s *ImportService) importFromCarrd(userID string, data []byte) (map[string]interface{}, error) {
	// Parse Carrd HTML export (simplified - actual implementation would parse HTML)
	// This is a placeholder implementation
	var carrdData struct {
		Title string   `json:"title"`
		Bio   string   `json:"bio"`
		Links []string `json:"links"`
	}

	if err := json.Unmarshal(data, &carrdData); err != nil {
		return nil, fmt.Errorf("failed to parse Carrd data: %w", err)
	}

	// Create profile
	slug := strings.ToLower(strings.ReplaceAll(carrdData.Title, " ", "-"))
	profile := &models.Profile{
		Slug:        slug,
		DisplayName: carrdData.Title,
		Bio:         carrdData.Bio,
	}

	if err := s.profileService.Create(userID, profile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Import links
	imported := 0
	for i, url := range carrdData.Links {
		link := &models.Link{
			ProfileID: profile.ID,
			Title:     fmt.Sprintf("Link %d", i+1),
			URL:       url,
			Position:  i + 1,
		}

		if err := s.linkService.Create(link); err != nil {
			continue
		}
		imported++
	}

	return map[string]interface{}{
		"profile_id":    profile.ID,
		"links_imported": imported,
		"total_links":   len(carrdData.Links),
	}, nil
}

func (s *ImportService) importFromAboutMe(userID string, data []byte) (map[string]interface{}, error) {
	// Parse About.me JSON export
	var aboutMeData struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Bio      string `json:"headline"`
		Avatar   string `json:"avatar"`
		Links    []struct {
			Label string `json:"label"`
			URL   string `json:"url"`
		} `json:"links"`
	}

	if err := json.Unmarshal(data, &aboutMeData); err != nil {
		return nil, fmt.Errorf("failed to parse About.me data: %w", err)
	}

	// Create profile
	profile := &models.Profile{
		Slug:        aboutMeData.Username,
		DisplayName: aboutMeData.Name,
		Bio:         aboutMeData.Bio,
		AvatarURL:   aboutMeData.Avatar,
	}

	if err := s.profileService.Create(userID, profile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Import links
	imported := 0
	for i, linkData := range aboutMeData.Links {
		link := &models.Link{
			ProfileID: profile.ID,
			Title:     linkData.Label,
			URL:       linkData.URL,
			Position:  i + 1,
		}

		if err := s.linkService.Create(link); err != nil {
			continue
		}
		imported++
	}

	return map[string]interface{}{
		"profile_id":    profile.ID,
		"links_imported": imported,
		"total_links":   len(aboutMeData.Links),
	}, nil
}

func (s *ImportService) importFromCSV(userID string, data []byte) (map[string]interface{}, error) {
	// Parse CSV format: title,url,username,service
	reader := csv.NewReader(strings.NewReader(string(data)))

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Validate header
	if len(header) < 2 || header[0] != "title" || header[1] != "url" {
		return nil, errors.New("invalid CSV format: expected columns 'title,url,username,service'")
	}

	// Create a default profile
	profile := &models.Profile{
		Slug:        fmt.Sprintf("imported-%d", time.Now().Unix()),
		DisplayName: "Imported Profile",
	}

	if err := s.profileService.Create(userID, profile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Import links
	imported := 0
	position := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // Skip invalid rows
		}

		if len(record) < 2 {
			continue
		}

		link := &models.Link{
			ProfileID: profile.ID,
			Title:     record[0],
			URL:       record[1],
			Position:  position,
		}

		if len(record) > 2 {
			link.Username = record[2]
		}

		if err := s.linkService.Create(link); err != nil {
			continue
		}

		imported++
		position++
	}

	return map[string]interface{}{
		"profile_id":    profile.ID,
		"links_imported": imported,
	}, nil
}

func (s *ImportService) importFromJSON(userID string, data []byte) (map[string]interface{}, error) {
	// Parse generic JSON format
	var jsonData struct {
		Profile struct {
			Slug        string `json:"slug"`
			DisplayName string `json:"display_name"`
			Bio         string `json:"bio"`
			AvatarURL   string `json:"avatar_url"`
		} `json:"profile"`
		Links []struct {
			Title    string `json:"title"`
			URL      string `json:"url"`
			Username string `json:"username,omitempty"`
		} `json:"links"`
	}

	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON data: %w", err)
	}

	// Create profile
	profile := &models.Profile{
		Slug:        jsonData.Profile.Slug,
		DisplayName: jsonData.Profile.DisplayName,
		Bio:         jsonData.Profile.Bio,
		AvatarURL:   jsonData.Profile.AvatarURL,
	}

	if profile.Slug == "" {
		profile.Slug = fmt.Sprintf("imported-%d", time.Now().Unix())
	}

	if err := s.profileService.Create(userID, profile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Import links
	imported := 0
	for i, linkData := range jsonData.Links {
		link := &models.Link{
			ProfileID: profile.ID,
			Title:     linkData.Title,
			URL:       linkData.URL,
			Username:  linkData.Username,
			Position:  i + 1,
		}

		if err := s.linkService.Create(link); err != nil {
			continue
		}
		imported++
	}

	return map[string]interface{}{
		"profile_id":    profile.ID,
		"links_imported": imported,
		"total_links":   len(jsonData.Links),
	}, nil
}

// Helper functions

func (s *ImportService) createImportJob(userID, source string) (string, error) {
	jobID := s.generateID()

	query := `
		INSERT INTO import_jobs (id, user_id, source, status, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, jobID, userID, source, "pending", time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to create import job: %w", err)
	}

	return jobID, nil
}

func (s *ImportService) updateJobStatus(jobID, status string, result map[string]interface{}) error {
	resultJSON := ""
	if result != nil {
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		resultJSON = string(data)
	}

	query := `
		UPDATE import_jobs
		SET status = ?, result = ?, completed_at = ?
		WHERE id = ?
	`

	completedAt := time.Now()
	if status == "processing" {
		completedAt = time.Time{} // NULL
	}

	_, err := s.db.Exec(query, status, resultJSON, completedAt, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

func (s *ImportService) generateID() string {
	return fmt.Sprintf("import-%d", time.Now().UnixNano())
}
