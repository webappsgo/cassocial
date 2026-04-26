package swagger

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler serves Swagger UI and OpenAPI specification
type Handler struct {
	spec *OpenAPISpec
}

// NewHandler creates a new Swagger handler
func NewHandler() *Handler {
	return &Handler{
		spec: generateOpenAPISpec(),
	}
}

// ServeUI serves the Swagger UI
func (h *Handler) ServeUI(w http.ResponseWriter, r *http.Request) {
	// Get theme preference (dark, light, auto)
	theme := getThemePreference(r)

	// Serve Swagger UI HTML
	html := h.generateSwaggerHTML(theme)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// ServeSpec serves the OpenAPI JSON specification
func (h *Handler) ServeSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.spec)
}

// OpenAPISpec represents the OpenAPI specification
type OpenAPISpec struct {
	OpenAPI string                 `json:"openapi"`
	Info    Info                   `json:"info"`
	Servers []Server               `json:"servers"`
	Paths   map[string]PathItem    `json:"paths"`
	Components Components           `json:"components"`
}

// Info represents API information
type Info struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Version     string  `json:"version"`
	Contact     Contact `json:"contact"`
	License     License `json:"license"`
}

// Contact represents contact information
type Contact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email"`
}

// License represents license information
type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Server represents an API server
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

// PathItem represents an API path
type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

// Operation represents an API operation
type Operation struct {
	Summary     string              `json:"summary"`
	Description string              `json:"description,omitempty"`
	OperationID string              `json:"operationId"`
	Tags        []string            `json:"tags"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

// Parameter represents an operation parameter
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"` // query, path, header
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

// RequestBody represents a request body
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

// MediaType represents a media type
type MediaType struct {
	Schema Schema `json:"schema"`
}

// Response represents an API response
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// Schema represents a JSON schema
type Schema struct {
	Type       string            `json:"type,omitempty"`
	Format     string            `json:"format,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Ref        string            `json:"$ref,omitempty"`
}

// Components represents reusable components
type Components struct {
	Schemas         map[string]Schema `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme represents a security scheme
type SecurityScheme struct {
	Type   string `json:"type"`
	Scheme string `json:"scheme,omitempty"`
	In     string `json:"in,omitempty"`
	Name   string `json:"name,omitempty"`
}

// generateOpenAPISpec generates the OpenAPI specification
func generateOpenAPISpec() *OpenAPISpec {
	return &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       "Cassocial API",
			Description: "Self-hosted link aggregator and social profile API",
			Version:     "1.0.0",
			Contact: Contact{
				Name: "casapps",
				URL:  "https://github.com/casapps/cassocial",
			},
			License: License{
				Name: "MIT",
				URL:  "https://github.com/casapps/cassocial/blob/main/LICENSE.md",
			},
		},
		Servers: []Server{
			{
				URL:         "/api/v1",
				Description: "API v1",
			},
		},
		Paths: generatePaths(),
		Components: Components{
			SecuritySchemes: map[string]SecurityScheme{
				"bearerAuth": {
					Type:   "http",
					Scheme: "bearer",
				},
			},
		},
	}
}

// generatePaths generates API paths
func generatePaths() map[string]PathItem {
	return map[string]PathItem{
		"/healthz": {
			Get: &Operation{
				Summary:     "Health check",
				Description: "Check API health status",
				OperationID: "healthCheck",
				Tags:        []string{"health"},
				Responses: map[string]Response{
					"200": {
						Description: "Successful response",
						Content: map[string]MediaType{
							"application/json": {
								Schema: Schema{Type: "object"},
							},
						},
					},
				},
			},
		},
		"/profiles": {
			Get: &Operation{
				Summary:     "List profiles",
				Description: "Get list of public profiles",
				OperationID: "listProfiles",
				Tags:        []string{"profiles"},
				Responses: map[string]Response{
					"200": {
						Description: "Successful response",
					},
				},
			},
		},
		"/profiles/{slug}": {
			Get: &Operation{
				Summary:     "Get profile",
				Description: "Get profile by slug",
				OperationID: "getProfile",
				Tags:        []string{"profiles"},
				Parameters: []Parameter{
					{
						Name:        "slug",
						In:          "path",
						Description: "Profile slug",
						Required:    true,
						Schema:      Schema{Type: "string"},
					},
				},
				Responses: map[string]Response{
					"200": {Description: "Successful response"},
					"404": {Description: "Profile not found"},
				},
			},
		},
		"/links": {
			Post: &Operation{
				Summary:     "Create link",
				Description: "Create a new link",
				OperationID: "createLink",
				Tags:        []string{"links"},
				Security: []map[string][]string{
					{"bearerAuth": {}},
				},
				Responses: map[string]Response{
					"201": {Description: "Link created"},
					"401": {Description: "Unauthorized"},
				},
			},
		},
	}
}

// generateSwaggerHTML generates Swagger UI HTML
func (h *Handler) generateSwaggerHTML(theme string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Cassocial API - Swagger UI</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
    <style>
        body { margin: 0; padding: 0; }
        .swagger-ui .topbar { display: none; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/openapi.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                theme: "` + theme + `",
                layout: "BaseLayout"
            });
        };
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
func (h *Handler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// renderError renders an error response
func (h *Handler) renderError(w http.ResponseWriter, status int, message string) {
	h.renderJSON(w, status, map[string]string{
		"error": message,
	})
}
