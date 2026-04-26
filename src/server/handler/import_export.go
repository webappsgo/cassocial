package handler

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// ImportExportHandler handles import/export operations
type ImportExportHandler struct {
	config *config.Config
	db     *store.DB
}

// NewImportExportHandler creates a new import/export handler
func NewImportExportHandler(cfg *config.Config, db *store.DB) *ImportExportHandler {
	return &ImportExportHandler{
		config: cfg,
		db:     db,
	}
}

// ImportRequest represents an import request
type ImportRequest struct {
	Source string          `json:"source"` // linktree, linkstack, carrd, aboutme, csv, json
	Data   json.RawMessage `json:"data"`
}

// HandleImport handles importing from various sources
func (h *ImportExportHandler) HandleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// TODO: Get user ID from session
	userID := "temp-user-id"

	var imported int
	var err error

	switch req.Source {
	case "linktree":
		imported, err = h.importFromLinktree(userID, req.Data)
	case "linkstack":
		imported, err = h.importFromLinkstack(userID, req.Data)
	case "carrd":
		imported, err = h.importFromCarrd(userID, req.Data)
	case "aboutme":
		imported, err = h.importFromAboutMe(userID, req.Data)
	case "csv":
		imported, err = h.importFromCSV(userID, req.Data)
	case "json":
		imported, err = h.importFromJSON(userID, req.Data)
	default:
		h.renderError(w, http.StatusBadRequest, "Unsupported import source")
		return
	}

	if err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "success",
		"message":  "Import completed successfully",
		"imported": imported,
		"source":   req.Source,
	})
}

// HandleExport handles exporting profile data
func (h *ImportExportHandler) HandleExport(w http.ResponseWriter, r *http.Request) {
	profileID := r.URL.Query().Get("profile_id")
	format := r.URL.Query().Get("format") // json, csv, html, pdf, vcard

	if profileID == "" {
		h.renderError(w, http.StatusBadRequest, "Profile ID required")
		return
	}

	// TODO: Get profile and links from database
	// TODO: Verify user owns profile

	switch format {
	case "json":
		h.exportJSON(w, profileID)
	case "csv":
		h.exportCSV(w, profileID)
	case "html":
		h.exportHTML(w, profileID)
	case "pdf":
		h.exportPDF(w, profileID)
	case "vcard":
		h.exportVCard(w, profileID)
	default:
		h.renderError(w, http.StatusBadRequest, "Invalid format. Use json, csv, html, pdf, or vcard")
	}
}

// Import implementations

func (h *ImportExportHandler) importFromLinktree(userID string, data json.RawMessage) (int, error) {
	// TODO: Parse Linktree export format
	// TODO: Create profile and links
	return 0, nil
}

func (h *ImportExportHandler) importFromLinkstack(userID string, data json.RawMessage) (int, error) {
	// TODO: Parse Linkstack export format
	return 0, nil
}

func (h *ImportExportHandler) importFromCarrd(userID string, data json.RawMessage) (int, error) {
	// TODO: Parse Carrd export format
	return 0, nil
}

func (h *ImportExportHandler) importFromAboutMe(userID string, data json.RawMessage) (int, error) {
	// TODO: Parse About.me export format
	return 0, nil
}

func (h *ImportExportHandler) importFromCSV(userID string, data json.RawMessage) (int, error) {
	// TODO: Parse CSV data
	return 0, nil
}

func (h *ImportExportHandler) importFromJSON(userID string, data json.RawMessage) (int, error) {
	// TODO: Parse JSON data
	var importData struct {
		Profile struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"profile"`
		Links []struct {
			Service string `json:"service"`
			URL     string `json:"url"`
			Title   string `json:"title"`
		} `json:"links"`
	}

	if err := json.Unmarshal(data, &importData); err != nil {
		return 0, err
	}

	// TODO: Create profile and links
	return len(importData.Links), nil
}

// Export implementations

func (h *ImportExportHandler) exportJSON(w http.ResponseWriter, profileID string) {
	// TODO: Get profile and links from database
	data := map[string]interface{}{
		"profile": map[string]interface{}{
			"id":          profileID,
			"title":       "Profile Title",
			"description": "Description",
		},
		"links": []interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=profile.json")
	json.NewEncoder(w).Encode(data)
}

func (h *ImportExportHandler) exportCSV(w http.ResponseWriter, profileID string) {
	// TODO: Get profile and links from database

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=profile.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Service", "Title", "URL", "Enabled"})

	// TODO: Write link data
	// writer.Write([]string{link.Service, link.Title, link.URL, "true"})
}

func (h *ImportExportHandler) exportHTML(w http.ResponseWriter, profileID string) {
	// TODO: Get profile and links
	// TODO: Generate standalone HTML page

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Profile Export</title>
    <meta charset="UTF-8">
</head>
<body>
    <h1>Profile</h1>
    <div class="links">
        <!-- Links here -->
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Disposition", "attachment; filename=profile.html")
	io.WriteString(w, html)
}

func (h *ImportExportHandler) exportPDF(w http.ResponseWriter, profileID string) {
	// TODO: Generate PDF
	http.Error(w, "PDF export not yet implemented", http.StatusNotImplemented)
}

func (h *ImportExportHandler) exportVCard(w http.ResponseWriter, profileID string) {
	// TODO: Get profile data
	// TODO: Generate vCard format

	vcard := `BEGIN:VCARD
VERSION:3.0
FN:Profile Name
URL:https://example.com/profile
END:VCARD`

	w.Header().Set("Content-Type", "text/vcard")
	w.Header().Set("Content-Disposition", "attachment; filename=profile.vcf")
	io.WriteString(w, vcard)
}

// generateGraphiQLHTML generates GraphiQL playground HTML
func (h *Handler) generateGraphiQLHTML(theme string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Cassocial GraphQL - GraphiQL</title>
    <style>
        body { margin: 0; height: 100vh; overflow: hidden; }
        #graphiql { height: 100vh; }
    </style>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphiql@3/graphiql.min.css">
</head>
<body>
    <div id="graphiql"></div>
    <script crossorigin src="https://unpkg.com/react@18/umd/react.production.min.js"></script>
    <script crossorigin src="https://unpkg.com/react-dom@18/umd/react-dom.production.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/graphiql@3/graphiql.min.js"></script>
    <script>
        const fetcher = GraphiQL.createFetcher({ url: '/graphql' });
        const root = ReactDOM.createRoot(document.getElementById('graphiql'));
        root.render(
            React.createElement(GraphiQL, {
                fetcher: fetcher,
                theme: '` + theme + `'
            })
        );
    </script>
</body>
</html>`
}

// getThemePreference gets theme preference from request
func getThemePreference(r *http.Request) string {
	// Check query parameter
	theme := r.URL.Query().Get("theme")
	if theme == "dark" || theme == "light" {
		return theme
	}

	// Check cookie
	if cookie, err := r.Cookie("theme"); err == nil {
		if cookie.Value == "dark" || cookie.Value == "light" {
			return cookie.Value
		}
	}

	// Default to dark
	return "dark"
}

// renderJSON renders a JSON response
func (h *ImportExportHandler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// renderError renders an error response
func (h *ImportExportHandler) renderError(w http.ResponseWriter, status int, message string) {
	h.renderJSON(w, status, map[string]string{
		"error": message,
	})
}
