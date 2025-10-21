package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/casapps/cassocial/internal/database"
	"github.com/casapps/cassocial/internal/models"
)

var (
	ErrLinkNotFound       = errors.New("link not found")
	ErrMaxLinksReached    = errors.New("maximum number of links reached")
	ErrInvalidPosition    = errors.New("invalid position")
	ErrLinkLimitExceeded  = errors.New("link limit exceeded for profile")
)

// LinkService handles link management business logic
type LinkService struct {
	db *database.DB
}

// NewLinkService creates a new link service
func NewLinkService(db *database.DB) *LinkService {
	return &LinkService{db: db}
}

// Create creates a new link
func (s *LinkService) Create(link *models.Link) error {
	// Check link limit for profile
	maxLinks, err := s.getMaxLinksPerProfile()
	if err != nil {
		return fmt.Errorf("failed to get max links setting: %w", err)
	}

	count, err := s.CountByProfileID(link.ProfileID)
	if err != nil {
		return fmt.Errorf("failed to count profile links: %w", err)
	}

	if count >= maxLinks {
		return ErrMaxLinksReached
	}

	// Validate link
	if err := link.Validate(); err != nil {
		return fmt.Errorf("link validation failed: %w", err)
	}

	// Set defaults
	link.ID = s.generateID()
	link.CreatedAt = time.Now()
	link.UpdatedAt = time.Now()
	link.IsActive = true
	link.ClickCount = 0

	// Get next position if not set
	if link.Position == 0 {
		link.Position, err = s.getNextPosition(link.ProfileID)
		if err != nil {
			return fmt.Errorf("failed to get next position: %w", err)
		}
	}

	// Insert into database
	query := `
		INSERT INTO links (
			id, profile_id, service_id, title, username, url, icon_url,
			background_color, text_color, position, is_active, click_count,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		link.ID, link.ProfileID, link.ServiceID, link.Title, link.Username,
		link.URL, link.IconURL, link.BackgroundColor, link.TextColor,
		link.Position, link.IsActive, link.ClickCount,
		link.CreatedAt, link.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert link: %w", err)
	}

	return nil
}

// GetByID retrieves a link by ID
func (s *LinkService) GetByID(id string) (*models.Link, error) {
	query := `SELECT * FROM links WHERE id = ?`

	var link models.Link
	err := s.db.QueryRow(query, id).Scan(
		&link.ID, &link.ProfileID, &link.ServiceID, &link.Title,
		&link.Username, &link.URL, &link.IconURL, &link.BackgroundColor,
		&link.TextColor, &link.Position, &link.IsActive, &link.ClickCount,
		&link.CreatedAt, &link.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrLinkNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	return &link, nil
}

// GetByProfileID retrieves all links for a profile
func (s *LinkService) GetByProfileID(profileID string) ([]*models.Link, error) {
	query := `SELECT * FROM links WHERE profile_id = ? ORDER BY position ASC`

	rows, err := s.db.Query(query, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query links: %w", err)
	}
	defer rows.Close()

	var links []*models.Link
	for rows.Next() {
		var link models.Link
		err := rows.Scan(
			&link.ID, &link.ProfileID, &link.ServiceID, &link.Title,
			&link.Username, &link.URL, &link.IconURL, &link.BackgroundColor,
			&link.TextColor, &link.Position, &link.IsActive, &link.ClickCount,
			&link.CreatedAt, &link.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		links = append(links, &link)
	}

	return links, nil
}

// GetActiveByProfileID retrieves all active links for a profile
func (s *LinkService) GetActiveByProfileID(profileID string) ([]*models.Link, error) {
	query := `SELECT * FROM links WHERE profile_id = ? AND is_active = 1 ORDER BY position ASC`

	rows, err := s.db.Query(query, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query links: %w", err)
	}
	defer rows.Close()

	var links []*models.Link
	for rows.Next() {
		var link models.Link
		err := rows.Scan(
			&link.ID, &link.ProfileID, &link.ServiceID, &link.Title,
			&link.Username, &link.URL, &link.IconURL, &link.BackgroundColor,
			&link.TextColor, &link.Position, &link.IsActive, &link.ClickCount,
			&link.CreatedAt, &link.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		links = append(links, &link)
	}

	return links, nil
}

// Update updates a link
func (s *LinkService) Update(link *models.Link) error {
	// Validate link
	if err := link.Validate(); err != nil {
		return fmt.Errorf("link validation failed: %w", err)
	}

	link.UpdatedAt = time.Now()

	query := `
		UPDATE links SET
			service_id = ?, title = ?, username = ?, url = ?, icon_url = ?,
			background_color = ?, text_color = ?, position = ?, is_active = ?,
			updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(query,
		link.ServiceID, link.Title, link.Username, link.URL, link.IconURL,
		link.BackgroundColor, link.TextColor, link.Position, link.IsActive,
		link.UpdatedAt, link.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update link: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrLinkNotFound
	}

	return nil
}

// Delete deletes a link
func (s *LinkService) Delete(id string) error {
	// Get link to know its profile and position
	link, err := s.GetByID(id)
	if err != nil {
		return err
	}

	// Delete the link
	query := `DELETE FROM links WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrLinkNotFound
	}

	// Reorder remaining links
	if err := s.reorderAfterDelete(link.ProfileID, link.Position); err != nil {
		return fmt.Errorf("failed to reorder links: %w", err)
	}

	return nil
}

// Toggle toggles the active state of a link
func (s *LinkService) Toggle(id string) error {
	query := `UPDATE links SET is_active = NOT is_active, updated_at = ? WHERE id = ?`

	result, err := s.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to toggle link: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrLinkNotFound
	}

	return nil
}

// Reorder reorders links for a profile
func (s *LinkService) Reorder(profileID string, linkIDs []string) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Update positions
	query := `UPDATE links SET position = ?, updated_at = ? WHERE id = ? AND profile_id = ?`

	for i, linkID := range linkIDs {
		_, err := tx.Exec(query, i+1, time.Now(), linkID, profileID)
		if err != nil {
			return fmt.Errorf("failed to update link position: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// IncrementClickCount increments the click count for a link
func (s *LinkService) IncrementClickCount(id string) error {
	query := `UPDATE links SET click_count = click_count + 1 WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}

	return nil
}

// CountByProfileID counts the number of links for a profile
func (s *LinkService) CountByProfileID(profileID string) (int, error) {
	query := `SELECT COUNT(*) FROM links WHERE profile_id = ?`

	var count int
	err := s.db.QueryRow(query, profileID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count links: %w", err)
	}

	return count, nil
}

// GetTopClickedLinks retrieves the top clicked links for a profile
func (s *LinkService) GetTopClickedLinks(profileID string, limit int) ([]*models.Link, error) {
	query := `
		SELECT * FROM links
		WHERE profile_id = ? AND is_active = 1
		ORDER BY click_count DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, profileID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query links: %w", err)
	}
	defer rows.Close()

	var links []*models.Link
	for rows.Next() {
		var link models.Link
		err := rows.Scan(
			&link.ID, &link.ProfileID, &link.ServiceID, &link.Title,
			&link.Username, &link.URL, &link.IconURL, &link.BackgroundColor,
			&link.TextColor, &link.Position, &link.IsActive, &link.ClickCount,
			&link.CreatedAt, &link.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		links = append(links, &link)
	}

	return links, nil
}

// Helper functions

func (s *LinkService) getNextPosition(profileID string) (int, error) {
	query := `SELECT COALESCE(MAX(position), 0) + 1 FROM links WHERE profile_id = ?`

	var position int
	err := s.db.QueryRow(query, profileID).Scan(&position)
	if err != nil {
		return 0, fmt.Errorf("failed to get next position: %w", err)
	}

	return position, nil
}

func (s *LinkService) reorderAfterDelete(profileID string, deletedPosition int) error {
	query := `
		UPDATE links
		SET position = position - 1, updated_at = ?
		WHERE profile_id = ? AND position > ?
	`

	_, err := s.db.Exec(query, time.Now(), profileID, deletedPosition)
	return err
}

func (s *LinkService) generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *LinkService) getMaxLinksPerProfile() (int, error) {
	maxLinksStr, err := s.db.GetSetting("max_links_per_profile")
	if err != nil {
		return 100, nil // Default value
	}

	var maxLinks int
	_, err = fmt.Sscanf(maxLinksStr, "%d", &maxLinks)
	if err != nil {
		return 100, nil // Default value
	}

	return maxLinks, nil
}
