package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
)

// newTestAdminHandlers creates an AdminHandlers backed by an in-memory SQLite database.
func newTestAdminHandlers(t *testing.T) *AdminHandlers {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect returned error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	authSvc := server.NewAuth(db, "test-secret-for-admin")
	return NewAdminHandlers(db, authSvc)
}

func TestNewAdminHandlers(t *testing.T) {
	h := newTestAdminHandlers(t)
	if h == nil {
		t.Fatal("NewAdminHandlers returned nil")
	}
}

func TestAdminHandlers_ListUsers_Empty(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	rr := httptest.NewRecorder()
	h.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ListUsers returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestAdminHandlers_ListUsers_WithUsers(t *testing.T) {
	h := newTestAdminHandlers(t)

	// Register a user via auth service so it appears in the list.
	_, err := h.auth.Register("adminlistuser", "adminlist@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	rr := httptest.NewRecorder()
	h.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ListUsers returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var users []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&users); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(users) == 0 {
		t.Error("expected at least one user in list, got none")
	}
}

func TestAdminHandlers_GetUser_MissingID(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users/", nil)
	rr := httptest.NewRecorder()
	h.GetUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("GetUser with empty ID returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAdminHandlers_GetUser_NotFound(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users/nonexistent-id", nil)
	req.SetPathValue("id", "nonexistent-id")
	rr := httptest.NewRecorder()
	h.GetUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("GetUser with nonexistent ID returned status %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestAdminHandlers_GetUser_Valid(t *testing.T) {
	h := newTestAdminHandlers(t)

	user, err := h.auth.Register("getusertest", "getuser@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users/"+user.ID, nil)
	req.SetPathValue("id", user.ID)
	rr := httptest.NewRecorder()
	h.GetUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetUser returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAdminHandlers_UpdateUser_MissingID(t *testing.T) {
	h := newTestAdminHandlers(t)

	body := `{"role":"admin"}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("UpdateUser with empty ID returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAdminHandlers_UpdateUser_NotFound(t *testing.T) {
	h := newTestAdminHandlers(t)

	body := `{"role":"admin"}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/no-such-id", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "no-such-id")
	rr := httptest.NewRecorder()
	h.UpdateUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("UpdateUser with nonexistent user returned status %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestAdminHandlers_UpdateUser_InvalidBody(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/some-id", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "some-id")
	rr := httptest.NewRecorder()
	h.UpdateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("UpdateUser with invalid body returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAdminHandlers_UpdateUser_Valid(t *testing.T) {
	h := newTestAdminHandlers(t)

	user, err := h.auth.Register("updateusertest", "updateuser@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"status": "active"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/"+user.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", user.ID)
	rr := httptest.NewRecorder()
	h.UpdateUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("UpdateUser returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAdminHandlers_DeleteUser_MissingID(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/users/", nil)
	rr := httptest.NewRecorder()
	h.DeleteUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("DeleteUser with empty ID returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAdminHandlers_DeleteUser_Valid(t *testing.T) {
	h := newTestAdminHandlers(t)

	user, err := h.auth.Register("deleteusertest", "deleteuser@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+user.ID, nil)
	req.SetPathValue("id", user.ID)
	rr := httptest.NewRecorder()
	h.DeleteUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("DeleteUser returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAdminHandlers_DeleteUser_Self(t *testing.T) {
	h := newTestAdminHandlers(t)

	user, err := h.auth.Register("selfdeletetest", "selfdelete@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	// Set the user ID in context to simulate the user trying to delete themselves.
	ctx := context.WithValue(context.Background(), server.ContextKeyUserID, user.ID)
	req := httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+user.ID, nil)
	req = req.WithContext(ctx)
	req.SetPathValue("id", user.ID)
	rr := httptest.NewRecorder()
	h.DeleteUser(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("DeleteUser self-delete returned status %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestAdminHandlers_GetSystemStats(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats", nil)
	rr := httptest.NewRecorder()
	h.GetSystemStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetSystemStats returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	for _, field := range []string{"total_users", "active_users", "total_profiles"} {
		if _, ok := resp[field]; !ok {
			t.Errorf("GetSystemStats response missing field %q", field)
		}
	}
}

func TestAdminHandlers_GetSettings(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings", nil)
	rr := httptest.NewRecorder()
	h.GetSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetSettings returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAdminHandlers_UpdateSettings_Valid(t *testing.T) {
	h := newTestAdminHandlers(t)

	body, _ := json.Marshal(map[string]string{"site_name": "Test Site"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("UpdateSettings returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAdminHandlers_UpdateSettings_InvalidBody(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateSettings(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("UpdateSettings with invalid body returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAdminHandlers_GetSMTPConfig(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/smtp/config", nil)
	rr := httptest.NewRecorder()
	h.GetSMTPConfig(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetSMTPConfig returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	// smtp_password must never appear in the response.
	if _, found := resp["smtp_password"]; found {
		t.Error("GetSMTPConfig response must not include smtp_password")
	}
}

func TestAdminHandlers_UpdateSMTPConfig_Valid(t *testing.T) {
	h := newTestAdminHandlers(t)

	body, _ := json.Marshal(map[string]string{"smtp_host": "mail.example.com", "smtp_port": "587"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/smtp/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateSMTPConfig(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("UpdateSMTPConfig returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAdminHandlers_UpdateSMTPConfig_InvalidBody(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/smtp/config", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateSMTPConfig(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("UpdateSMTPConfig with invalid body returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAdminHandlers_GetNotificationPreferences(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/notifications/preferences", nil)
	rr := httptest.NewRecorder()
	h.GetNotificationPreferences(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetNotificationPreferences returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["notify_emergency"]; !ok {
		t.Error("GetNotificationPreferences response missing 'notify_emergency' field")
	}
}

func TestAdminHandlers_UpdateNotificationPreferences_Valid(t *testing.T) {
	h := newTestAdminHandlers(t)

	body, _ := json.Marshal(map[string]string{"notify_emergency": "true"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/notifications/preferences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateNotificationPreferences(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("UpdateNotificationPreferences returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAdminHandlers_UpdateNotificationPreferences_InvalidBody(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/notifications/preferences", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateNotificationPreferences(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("UpdateNotificationPreferences with invalid body returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAdminHandlers_ClearCache(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/cache/clear", nil)
	rr := httptest.NewRecorder()
	h.ClearCache(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ClearCache returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAdminHandlers_TriggerBackup(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/backup", nil)
	rr := httptest.NewRecorder()
	h.TriggerBackup(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("TriggerBackup returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestAdminHandlers_ImportServices(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/services/import", nil)
	rr := httptest.NewRecorder()
	h.ImportServices(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Errorf("ImportServices returned status %d, want %d", rr.Code, http.StatusNotImplemented)
	}
}

// ---- GetSettings — scan loop produces rows ----
// After migration there are default settings rows; this exercises the rows.Next body.
func TestAdminHandlers_GetSettings_WithRows(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings", nil)
	rr := httptest.NewRecorder()
	h.GetSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetSettings returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var settings []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&settings); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(settings) == 0 {
		t.Error("GetSettings expected non-empty settings list after migration")
	}
}

// ---- UpdateSettings — zero-key map exercises the empty-loop path ----
func TestAdminHandlers_UpdateSettings_EmptyBody(t *testing.T) {
	h := newTestAdminHandlers(t)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("UpdateSettings with empty map returned status %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

// ---- GetSystemStats — exercises all count queries ----
func TestAdminHandlers_GetSystemStats_WithData(t *testing.T) {
	h := newTestAdminHandlers(t)

	// Register a user and verify it appears in stats.
	_, err := h.auth.Register("statsuser", "statsuser@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats", nil)
	rr := httptest.NewRecorder()
	h.GetSystemStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetSystemStats with data returned status %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if totalUsers, _ := resp["total_users"].(float64); totalUsers == 0 {
		t.Errorf("GetSystemStats total_users expected > 0, got %v", resp["total_users"])
	}
}

// ---- TestSMTPConnection (was 0% covered) ----

func TestAdminHandlers_TestSMTPConnection(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/smtp/test", nil)
	rr := httptest.NewRecorder()
	h.TestSMTPConnection(rr, req)

	// The handler is not yet implemented and must return 501.
	if rr.Code != http.StatusNotImplemented {
		t.Errorf("TestSMTPConnection returned status %d, want %d; body: %s",
			rr.Code, http.StatusNotImplemented, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode TestSMTPConnection response: %v", err)
	}
	if _, ok := resp["message"]; !ok {
		t.Error("TestSMTPConnection response missing 'message' field")
	}
}

// ---- ListUsers — scan error branch ----
// We cannot easily force a scan error in SQLite, but we can verify the
// empty-users-table scan path (the rows.Next() body is exercised when
// the table is populated, which is already tested above).  This test
// ensures the non-empty path exercises the Scan loop successfully.
func TestAdminHandlers_ListUsers_ScanLoop(t *testing.T) {
	h := newTestAdminHandlers(t)

	// Insert multiple users to ensure the scan loop body executes more than once.
	for i, name := range []string{"scanloopA", "scanloopB", "scanloopC"} {
		_, err := h.auth.Register(name, name+"@example.com", "ValidPass1")
		if err != nil {
			t.Fatalf("Register[%d] returned error: %v", i, err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	rr := httptest.NewRecorder()
	h.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ListUsers returned status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var users []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&users); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(users) < 3 {
		t.Errorf("expected at least 3 users, got %d", len(users))
	}
}

// ---- UpdateUser — validation failure branch ----
// Sending an invalid role triggers user.Validate() to fail.
func TestAdminHandlers_UpdateUser_InvalidRole(t *testing.T) {
	h := newTestAdminHandlers(t)

	user, err := h.auth.Register("invalidroletest", "invalidrole@example.com", "ValidPass1")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"role": "superadmin-does-not-exist"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/"+user.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", user.ID)
	rr := httptest.NewRecorder()
	h.UpdateUser(rr, req)

	// Should fail validation: invalid role value.
	if rr.Code != http.StatusBadRequest {
		t.Errorf("UpdateUser with invalid role returned status %d, want non-200; body: %s",
			rr.Code, rr.Body.String())
	}
}

// ---- DeleteUser — non-existent user still returns 200 (DELETE is idempotent) ----
func TestAdminHandlers_DeleteUser_NonExistent(t *testing.T) {
	h := newTestAdminHandlers(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/users/no-such-user-id", nil)
	req.SetPathValue("id", "no-such-user-id")
	rr := httptest.NewRecorder()
	h.DeleteUser(rr, req)

	// DELETE of a non-existent row is still a success (no rows affected is not an error in SQL).
	if rr.Code != http.StatusOK {
		t.Errorf("DeleteUser non-existent returned status %d, want %d; body: %s",
			rr.Code, http.StatusOK, rr.Body.String())
	}
}

// ---- UpdateSettings / UpdateSMTPConfig / UpdateNotificationPreferences DB error paths ----
// Override db.Driver to "pgx" so that db.SetSetting uses $1/$2 placeholders
// which SQLite rejects, causing the error branch inside the handlers to execute.

func newClosedDBAdminHandlers(t *testing.T) *AdminHandlers {
	t.Helper()

	db, err := store.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("store.Connect: %v", err)
	}

	if err := db.RunMigrations(); err != nil {
		db.Close()
		t.Fatalf("RunMigrations: %v", err)
	}

	authSvc := server.NewAuth(db, "test-secret-for-admin")
	h := NewAdminHandlers(db, authSvc)
	// Close the DB so all subsequent operations fail.
	db.Close()
	return h
}

func TestAdminHandlers_UpdateSettings_DBError(t *testing.T) {
	h := newClosedDBAdminHandlers(t)

	body, _ := json.Marshal(map[string]string{"some_key": "some_value"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateSettings(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("UpdateSettings (DB error) returned %d, want %d; body: %s",
			rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

func TestAdminHandlers_UpdateSMTPConfig_DBError(t *testing.T) {
	h := newClosedDBAdminHandlers(t)

	body, _ := json.Marshal(map[string]string{"smtp_host": "mail.example.com"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/smtp/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateSMTPConfig(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("UpdateSMTPConfig (DB error) returned %d, want %d; body: %s",
			rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

func TestAdminHandlers_UpdateNotificationPreferences_DBError(t *testing.T) {
	h := newClosedDBAdminHandlers(t)

	body, _ := json.Marshal(map[string]string{"notify_emergency": "true"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/notifications/preferences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateNotificationPreferences(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("UpdateNotificationPreferences (DB error) returned %d, want %d; body: %s",
			rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}
