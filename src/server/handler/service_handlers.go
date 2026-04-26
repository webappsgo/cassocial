package handler

import (
	"net/http"

	"github.com/casapps/cassocial/src/server/store"
	"github.com/casapps/cassocial/src/server/model"
)

// ServiceHandlers handles service-related HTTP requests
type ServiceHandlers struct {
	db *store.DB
}

// NewServiceHandlers creates a new ServiceHandlers instance
func NewServiceHandlers(db *store.DB) *ServiceHandlers {
	return &ServiceHandlers{db: db}
}

// ListServices lists all active services
// GET /api/services
func (h *ServiceHandlers) ListServices(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	category := r.URL.Query().Get("category")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	query := `SELECT id, name, category, icon_url, icon_svg, url_pattern,
			  background_color, text_color, popularity, is_active, requires_username,
			  placeholder_text, validation_pattern, created_at, updated_at
			  FROM services WHERE is_active = ?`

	args := []interface{}{true}

	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}

	query += " ORDER BY popularity DESC, name ASC"

	if limit != "" {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	if offset != "" {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	if h.db.Driver == "postgres" {
		query = replaceQuestionMarksWithArgs(query, len(args))
	}

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch services")
		return
	}
	defer rows.Close()

	services := []model.Service{}
	for rows.Next() {
		var s model.Service
		err := rows.Scan(&s.ID, &s.Name, &s.Category, &s.IconURL, &s.IconSVG, &s.URLPattern,
			&s.BackgroundColor, &s.TextColor, &s.Popularity, &s.IsActive, &s.RequiresUsername,
			&s.PlaceholderText, &s.ValidationPattern, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			continue
		}
		services = append(services, s)
	}

	respondJSON(w, http.StatusOK, services)
}

// SearchServices searches for services by name
// GET /api/services/search?q={query}
func (h *ServiceHandlers) SearchServices(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		respondError(w, http.StatusBadRequest, "search query required")
		return
	}

	query := `SELECT id, name, category, icon_url, icon_svg, url_pattern,
			  background_color, text_color, popularity, is_active, requires_username,
			  placeholder_text, validation_pattern, created_at, updated_at
			  FROM services WHERE is_active = ? AND name LIKE ?
			  ORDER BY popularity DESC, name ASC LIMIT 50`

	if h.db.Driver == "postgres" {
		query = `SELECT id, name, category, icon_url, icon_svg, url_pattern,
				 background_color, text_color, popularity, is_active, requires_username,
				 placeholder_text, validation_pattern, created_at, updated_at
				 FROM services WHERE is_active = $1 AND name ILIKE $2
				 ORDER BY popularity DESC, name ASC LIMIT 50`
	}

	searchPattern := "%" + searchQuery + "%"
	rows, err := h.db.Query(query, true, searchPattern)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to search services")
		return
	}
	defer rows.Close()

	services := []model.Service{}
	for rows.Next() {
		var s model.Service
		err := rows.Scan(&s.ID, &s.Name, &s.Category, &s.IconURL, &s.IconSVG, &s.URLPattern,
			&s.BackgroundColor, &s.TextColor, &s.Popularity, &s.IsActive, &s.RequiresUsername,
			&s.PlaceholderText, &s.ValidationPattern, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			continue
		}
		services = append(services, s)
	}

	respondJSON(w, http.StatusOK, services)
}

// ListCategories lists all service categories
// GET /api/services/categories
func (h *ServiceHandlers) ListCategories(w http.ResponseWriter, r *http.Request) {
	query := `SELECT DISTINCT category, COUNT(*) as count
			  FROM services WHERE is_active = ?
			  GROUP BY category ORDER BY category ASC`

	if h.db.Driver == "postgres" {
		query = `SELECT DISTINCT category, COUNT(*) as count
				 FROM services WHERE is_active = $1
				 GROUP BY category ORDER BY category ASC`
	}

	rows, err := h.db.Query(query, true)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}
	defer rows.Close()

	categories := []map[string]interface{}{}
	for rows.Next() {
		var category string
		var count int
		err := rows.Scan(&category, &count)
		if err != nil {
			continue
		}
		categories = append(categories, map[string]interface{}{
			"category": category,
			"count":    count,
		})
	}

	respondJSON(w, http.StatusOK, categories)
}

// ListPopularServices lists popular services
// GET /api/services/popular
func (h *ServiceHandlers) ListPopularServices(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "20"
	}

	query := `SELECT id, name, category, icon_url, icon_svg, url_pattern,
			  background_color, text_color, popularity, is_active, requires_username,
			  placeholder_text, validation_pattern, created_at, updated_at
			  FROM services WHERE is_active = ?
			  ORDER BY popularity DESC LIMIT ?`

	if h.db.Driver == "postgres" {
		query = `SELECT id, name, category, icon_url, icon_svg, url_pattern,
				 background_color, text_color, popularity, is_active, requires_username,
				 placeholder_text, validation_pattern, created_at, updated_at
				 FROM services WHERE is_active = $1
				 ORDER BY popularity DESC LIMIT $2`
	}

	rows, err := h.db.Query(query, true, limitStr)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch popular services")
		return
	}
	defer rows.Close()

	services := []model.Service{}
	for rows.Next() {
		var s model.Service
		err := rows.Scan(&s.ID, &s.Name, &s.Category, &s.IconURL, &s.IconSVG, &s.URLPattern,
			&s.BackgroundColor, &s.TextColor, &s.Popularity, &s.IsActive, &s.RequiresUsername,
			&s.PlaceholderText, &s.ValidationPattern, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			continue
		}
		services = append(services, s)
	}

	respondJSON(w, http.StatusOK, services)
}

// GetService retrieves a specific service by ID
// GET /api/services/{id}
func (h *ServiceHandlers) GetService(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("id")
	if serviceID == "" {
		respondError(w, http.StatusBadRequest, "service ID required")
		return
	}

	service := &model.Service{}
	query := `SELECT id, name, category, icon_url, icon_svg, url_pattern,
			  background_color, text_color, popularity, is_active, requires_username,
			  placeholder_text, validation_pattern, created_at, updated_at
			  FROM services WHERE id = ?`

	if h.db.Driver == "postgres" {
		query = replaceQuestionMarks(query, 1)
	}

	err := h.db.QueryRow(query, serviceID).Scan(&service.ID, &service.Name, &service.Category,
		&service.IconURL, &service.IconSVG, &service.URLPattern, &service.BackgroundColor,
		&service.TextColor, &service.Popularity, &service.IsActive, &service.RequiresUsername,
		&service.PlaceholderText, &service.ValidationPattern, &service.CreatedAt, &service.UpdatedAt)

	if err != nil {
		respondError(w, http.StatusNotFound, "service not found")
		return
	}

	respondJSON(w, http.StatusOK, service)
}
