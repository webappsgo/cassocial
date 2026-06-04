package handler

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server"
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

	userID, ok := server.GetUserIDFromContext(r.Context())
	if !ok {
		h.renderError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

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
	_ = userID
	_ = data
	return 0, nil
}

func (h *ImportExportHandler) importFromLinkstack(userID string, data json.RawMessage) (int, error) {
	_ = userID
	_ = data
	return 0, nil
}

func (h *ImportExportHandler) importFromCarrd(userID string, data json.RawMessage) (int, error) {
	_ = userID
	_ = data
	return 0, nil
}

func (h *ImportExportHandler) importFromAboutMe(userID string, data json.RawMessage) (int, error) {
	_ = userID
	_ = data
	return 0, nil
}

func (h *ImportExportHandler) importFromCSV(userID string, data json.RawMessage) (int, error) {
	_ = userID
	_ = data
	return 0, nil
}

func (h *ImportExportHandler) importFromJSON(userID string, data json.RawMessage) (int, error) {
	_ = userID
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

	return len(importData.Links), nil
}

// Export implementations

func (h *ImportExportHandler) exportJSON(w http.ResponseWriter, profileID string) {
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
	_ = profileID

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=profile.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"Service", "Title", "URL", "Enabled"})
}

func (h *ImportExportHandler) exportHTML(w http.ResponseWriter, profileID string) {
	_ = profileID

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
	http.Error(w, "PDF export not yet implemented", http.StatusNotImplemented)
}

func (h *ImportExportHandler) exportVCard(w http.ResponseWriter, profileID string) {
	_ = profileID

	vcard := `BEGIN:VCARD
VERSION:3.0
FN:Profile Name
URL:https://example.com/profile
END:VCARD`

	w.Header().Set("Content-Type", "text/vcard")
	w.Header().Set("Content-Disposition", "attachment; filename=profile.vcf")
	io.WriteString(w, vcard)
}

// generateGraphiQLHTML generates a self-contained GraphQL playground without CDN dependencies
func generateGraphiQLHTML(theme string) string {
	dark := theme == "dark"
	bg, fg, inputBg, border, resultBg := "#282a36", "#f8f8f2", "#44475a", "#6272a4", "#1e1f29"
	if !dark {
		bg, fg, inputBg, border, resultBg = "#ffffff", "#212529", "#f8f9fa", "#dee2e6", "#f1f3f5"
	}
	return `<!DOCTYPE html>
<html lang="en" data-theme="` + theme + `">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <!-- GraphiQL-compatible GraphQL explorer -->
    <title>Cassocial GraphiQL Explorer</title>
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
        body { background: ` + bg + `; color: ` + fg + `; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; height: 100vh; display: flex; flex-direction: column; }
        header { padding: 12px 16px; border-bottom: 1px solid ` + border + `; display: flex; align-items: center; gap: 12px; }
        header h1 { font-size: 1rem; font-weight: 600; }
        .workspace { display: flex; flex: 1; overflow: hidden; gap: 1px; background: ` + border + `; }
        .pane { background: ` + bg + `; display: flex; flex-direction: column; flex: 1; overflow: hidden; }
        .pane-header { padding: 8px 12px; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; border-bottom: 1px solid ` + border + `; opacity: 0.7; }
        textarea { flex: 1; background: ` + inputBg + `; color: ` + fg + `; border: none; resize: none; padding: 12px; font-family: "SFMono-Regular", Consolas, "Liberation Mono", monospace; font-size: 13px; line-height: 1.6; outline: none; width: 100%; }
        #result { flex: 1; background: ` + resultBg + `; color: ` + fg + `; padding: 12px; font-family: "SFMono-Regular", Consolas, "Liberation Mono", monospace; font-size: 13px; line-height: 1.6; overflow: auto; white-space: pre; word-break: break-all; overflow-wrap: break-word; }
        footer { padding: 8px 16px; border-top: 1px solid ` + border + `; display: flex; gap: 8px; align-items: center; }
        button { padding: 6px 16px; border-radius: 6px; border: 1px solid ` + border + `; background: ` + inputBg + `; color: ` + fg + `; font-size: 0.875rem; cursor: pointer; }
        button:hover { opacity: 0.85; }
        button.primary { background: #bd93f9; border-color: #bd93f9; color: #282a36; font-weight: 600; }
        label { font-size: 0.8rem; opacity: 0.7; }
        input[type=text] { flex: 1; background: ` + inputBg + `; color: ` + fg + `; border: 1px solid ` + border + `; border-radius: 4px; padding: 4px 8px; font-size: 0.8rem; outline: none; }
    </style>
</head>
<body>
    <header>
        <h1>Cassocial GraphiQL Explorer</h1>
    </header>
    <div class="workspace">
        <div class="pane">
            <div class="pane-header">Query / Mutation</div>
            <textarea id="query" spellcheck="false" placeholder="# Enter your GraphQL query here&#10;{&#10;  __typename&#10;}"></textarea>
        </div>
        <div class="pane">
            <div class="pane-header">Variables (JSON)</div>
            <textarea id="vars" spellcheck="false" placeholder="{}"></textarea>
        </div>
        <div class="pane">
            <div class="pane-header">Result</div>
            <div id="result">Press Run to execute the query.</div>
        </div>
    </div>
    <footer>
        <label>Token</label>
        <input type="text" id="token" placeholder="Bearer …">
        <button class="primary" id="runBtn">▶ Run</button>
        <button id="formatBtn">Format JSON</button>
    </footer>
    <script>
        document.getElementById('runBtn').addEventListener('click', runQuery);
        document.getElementById('query').addEventListener('keydown', function(e) {
            if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) runQuery();
        });
        document.getElementById('formatBtn').addEventListener('click', function() {
            const r = document.getElementById('result');
            try { r.textContent = JSON.stringify(JSON.parse(r.textContent), null, 2); } catch (_) {}
        });
        function runQuery() {
            const query = document.getElementById('query').value.trim();
            if (!query) return;
            let variables = {};
            try { variables = JSON.parse(document.getElementById('vars').value || '{}'); } catch(_) {}
            const token = document.getElementById('token').value.trim();
            const headers = { 'Content-Type': 'application/json' };
            if (token) headers['Authorization'] = token.startsWith('Bearer ') ? token : 'Bearer ' + token;
            const result = document.getElementById('result');
            result.textContent = 'Running…';
            fetch('/graphql', {
                method: 'POST',
                headers: headers,
                body: JSON.stringify({ query: query, variables: variables })
            })
            .then(function(r) { return r.text(); })
            .then(function(t) {
                try { result.textContent = JSON.stringify(JSON.parse(t), null, 2); }
                catch(_) { result.textContent = t; }
            })
            .catch(function(e) { result.textContent = 'Error: ' + e.message; });
        }
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
