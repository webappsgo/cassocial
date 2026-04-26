package graphql

import (
	"encoding/json"
	"net/http"
)

// Handler serves GraphQL API and GraphiQL playground
type Handler struct {
	schema *Schema
}

// NewHandler creates a new GraphQL handler
func NewHandler() *Handler {
	return &Handler{
		schema: buildSchema(),
	}
}

// ServeGraphiQL serves the GraphiQL playground UI
func (h *Handler) ServeGraphiQL(w http.ResponseWriter, r *http.Request) {
	// Get theme preference (dark, light, auto)
	theme := getThemePreference(r)

	html := h.generateGraphiQLHTML(theme)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// ServeGraphQL handles GraphQL queries
func (h *Handler) ServeGraphQL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// TODO: Execute GraphQL query
	// For now, return a mock response
	response := GraphQLResponse{
		Data: map[string]interface{}{
			"profiles": []interface{}{},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	OperationName string              `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}   `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
	Path    []string `json:"path,omitempty"`
}

// Schema represents the GraphQL schema
type Schema struct {
	Query    *ObjectType
	Mutation *ObjectType
}

// ObjectType represents a GraphQL object type
type ObjectType struct {
	Name   string
	Fields map[string]*Field
}

// Field represents a GraphQL field
type Field struct {
	Type        string
	Description string
	Args        map[string]*Argument
}

// Argument represents a field argument
type Argument struct {
	Type        string
	Description string
}

// buildSchema builds the GraphQL schema
func buildSchema() *Schema {
	return &Schema{
		Query: &ObjectType{
			Name: "Query",
			Fields: map[string]*Field{
				"profiles": {
					Type:        "[Profile]",
					Description: "List all public profiles",
				},
				"profile": {
					Type:        "Profile",
					Description: "Get profile by slug",
					Args: map[string]*Argument{
						"slug": {
							Type:        "String!",
							Description: "Profile slug",
						},
					},
				},
				"links": {
					Type:        "[Link]",
					Description: "Get links for a profile",
					Args: map[string]*Argument{
						"profileId": {
							Type:        "ID!",
							Description: "Profile ID",
						},
					},
				},
				"services": {
					Type:        "[Service]",
					Description: "List all supported services",
				},
			},
		},
		Mutation: &ObjectType{
			Name: "Mutation",
			Fields: map[string]*Field{
				"createProfile": {
					Type:        "Profile",
					Description: "Create a new profile",
				},
				"updateProfile": {
					Type:        "Profile",
					Description: "Update a profile",
				},
				"deleteProfile": {
					Type:        "Boolean",
					Description: "Delete a profile",
				},
				"createLink": {
					Type:        "Link",
					Description: "Create a new link",
				},
				"updateLink": {
					Type:        "Link",
					Description: "Update a link",
				},
				"deleteLink": {
					Type:        "Boolean",
					Description: "Delete a link",
				},
			},
		},
	}
}

// generateGraphiQLHTML generates GraphiQL playground HTML
func (h *Handler) generateGraphiQLHTML(theme string) string {
	themeClass := "graphiql-dark"
	if theme == "light" {
		themeClass = "graphiql-light"
	}

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
<body class="` + themeClass + `">
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
                defaultTheme: '` + theme + `'
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
