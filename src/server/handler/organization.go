package handler

import (
	"encoding/json"
	"net/http"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// OrganizationHandler handles organization/team features
type OrganizationHandler struct {
	config *config.Config
	db     *store.DB
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(cfg *config.Config, db *store.DB) *OrganizationHandler {
	return &OrganizationHandler{
		config: cfg,
		db:     db,
	}
}

// CreateOrganizationRequest represents an organization creation request
type CreateOrganizationRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// HandleCreateOrganization creates a new organization
func (h *OrganizationHandler) HandleCreateOrganization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get user ID from session
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// TODO: Create organization
	// TODO: Add creator as owner

	h.renderJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  "success",
		"message": "Organization created successfully",
		"org_id":  "temp-org-id",
	})
}

// HandleGetOrganization retrieves an organization
func (h *OrganizationHandler) HandleGetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("id")
	if orgID == "" {
		h.renderError(w, http.StatusBadRequest, "Organization ID required")
		return
	}

	// TODO: Get organization from database
	org := map[string]interface{}{
		"id":          orgID,
		"name":        "Organization",
		"slug":        "org-slug",
		"description": "Description",
		"members":     []interface{}{},
	}

	h.renderJSON(w, http.StatusOK, org)
}

// HandleListOrganizations lists user's organizations
func (h *OrganizationHandler) HandleListOrganizations(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user ID from session
	// TODO: Get organizations where user is a member

	orgs := []interface{}{}

	h.renderJSON(w, http.StatusOK, map[string]interface{}{
		"organizations": orgs,
		"total":         len(orgs),
	})
}

// HandleAddMember adds a member to an organization
func (h *OrganizationHandler) HandleAddMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OrgID  string `json:"org_id"`
		Email  string `json:"email"`
		Role   string `json:"role"` // owner, admin, member
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// TODO: Verify requester is owner/admin
	// TODO: Send invitation email
	// TODO: Add member to organization

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Invitation sent",
	})
}

// HandleRemoveMember removes a member from an organization
func (h *OrganizationHandler) HandleRemoveMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OrgID    string `json:"org_id"`
		MemberID string `json:"member_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// TODO: Verify requester is owner/admin
	// TODO: Remove member from organization

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Member removed",
	})
}

// HandleUpdateMemberRole updates a member's role
func (h *OrganizationHandler) HandleUpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OrgID    string `json:"org_id"`
		MemberID string `json:"member_id"`
		Role     string `json:"role"` // owner, admin, member
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// TODO: Verify requester is owner
	// TODO: Update member role

	h.renderJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Role updated",
	})
}

// renderJSON renders a JSON response
func (h *OrganizationHandler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// renderError renders an error response
func (h *OrganizationHandler) renderError(w http.ResponseWriter, status int, message string) {
	h.renderJSON(w, status, map[string]string{
		"error": message,
	})
}
