package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

func newTestOrganizationHandler(t *testing.T) (*OrganizationHandler, *store.DB) {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	return NewOrganizationHandler(&config.Config{}, db), db
}

// TestNewOrganizationHandler verifies the constructor.
func TestNewOrganizationHandler(t *testing.T) {
	h, _ := newTestOrganizationHandler(t)
	if h == nil {
		t.Fatal("NewOrganizationHandler returned nil")
	}
}

// TestHandleCreateOrganization covers: wrong method, missing auth, valid request.
func TestHandleCreateOrganization(t *testing.T) {
	h, db := newTestOrganizationHandler(t)
	userID := createTestUser(t, db, "orgcreator", "orgcreator@example.com")

	t.Run("wrong method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orgs", nil)
		rr := httptest.NewRecorder()
		h.HandleCreateOrganization(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("no auth returns 401", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"name": "Acme", "slug": "acme"})
		req := httptest.NewRequest(http.MethodPost, "/api/orgs", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		h.HandleCreateOrganization(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("bad JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/orgs", bytes.NewBufferString("{bad"))
		req = withUserID(req, userID)
		rr := httptest.NewRecorder()
		h.HandleCreateOrganization(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid create returns 201", func(t *testing.T) {
		body, _ := json.Marshal(CreateOrganizationRequest{
			Name:        "Acme Inc",
			Slug:        "acme-inc",
			Description: "A test org",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/orgs", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = withUserID(req, userID)
		rr := httptest.NewRecorder()
		h.HandleCreateOrganization(rr, req)
		if rr.Code != http.StatusCreated {
			t.Errorf("got %d, want %d", rr.Code, http.StatusCreated)
		}
	})
}

// TestHandleGetOrganization covers: missing ID, valid ID.
func TestHandleGetOrganization(t *testing.T) {
	h, _ := newTestOrganizationHandler(t)

	t.Run("missing id returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orgs/", nil)
		rr := httptest.NewRecorder()
		h.HandleGetOrganization(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid id returns 200 with org data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orgs/?id=org-123", nil)
		rr := httptest.NewRecorder()
		h.HandleGetOrganization(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		var resp map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("response not valid JSON: %v", err)
		}
		if resp["id"] != "org-123" {
			t.Errorf("id = %q, want org-123", resp["id"])
		}
	})
}

// TestHandleListOrganizations covers: missing auth, authenticated request.
func TestHandleListOrganizations(t *testing.T) {
	h, db := newTestOrganizationHandler(t)
	userID := createTestUser(t, db, "orglistuser", "orglistuser@example.com")

	t.Run("no auth returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orgs", nil)
		rr := httptest.NewRecorder()
		h.HandleListOrganizations(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("authenticated returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orgs", nil)
		req = withUserID(req, userID)
		rr := httptest.NewRecorder()
		h.HandleListOrganizations(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
		var resp map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("response not valid JSON: %v", err)
		}
		if _, ok := resp["organizations"]; !ok {
			t.Error("response missing 'organizations' field")
		}
	})
}

// TestHandleAddMember covers: wrong method, bad JSON, valid request.
func TestHandleAddMember(t *testing.T) {
	h, _ := newTestOrganizationHandler(t)

	t.Run("wrong method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orgs/members", nil)
		rr := httptest.NewRecorder()
		h.HandleAddMember(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("bad JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/orgs/members", bytes.NewBufferString("{bad"))
		rr := httptest.NewRecorder()
		h.HandleAddMember(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid add returns 200", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"org_id": "org-1",
			"email":  "member@example.com",
			"role":   "member",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/orgs/members", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.HandleAddMember(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
		}
	})
}

// TestHandleRemoveMember covers: wrong method, bad JSON, DELETE and POST methods.
func TestHandleRemoveMember(t *testing.T) {
	h, _ := newTestOrganizationHandler(t)

	t.Run("wrong method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orgs/members", nil)
		rr := httptest.NewRecorder()
		h.HandleRemoveMember(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("bad JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/orgs/members", bytes.NewBufferString("{bad"))
		rr := httptest.NewRecorder()
		h.HandleRemoveMember(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	for _, method := range []string{http.MethodPost, http.MethodDelete} {
		method := method
		t.Run("valid remove via "+method+" returns 200", func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{
				"org_id":    "org-1",
				"member_id": "user-2",
			})
			req := httptest.NewRequest(method, "/api/orgs/members", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h.HandleRemoveMember(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("method=%s got %d, want %d", method, rr.Code, http.StatusOK)
			}
		})
	}
}

// TestHandleUpdateMemberRole covers: wrong method, bad JSON, POST and PUT.
func TestHandleUpdateMemberRole(t *testing.T) {
	h, _ := newTestOrganizationHandler(t)

	t.Run("wrong method returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orgs/members/role", nil)
		rr := httptest.NewRecorder()
		h.HandleUpdateMemberRole(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("bad JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/orgs/members/role", bytes.NewBufferString("{bad"))
		rr := httptest.NewRecorder()
		h.HandleUpdateMemberRole(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	for _, method := range []string{http.MethodPost, http.MethodPut} {
		method := method
		t.Run("valid role update via "+method+" returns 200", func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{
				"org_id":    "org-1",
				"member_id": "user-2",
				"role":      "admin",
			})
			req := httptest.NewRequest(method, "/api/orgs/members/role", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h.HandleUpdateMemberRole(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("method=%s got %d, want %d", method, rr.Code, http.StatusOK)
			}
			var resp map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("response not valid JSON: %v", err)
			}
			if resp["status"] != "success" {
				t.Errorf("status = %q, want success", resp["status"])
			}
		})
	}
}
