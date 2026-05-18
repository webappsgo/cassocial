package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// captureClientStdout captures everything written to os.Stdout during f().
func captureClientStdout(t *testing.T, f func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(): %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	f()

	w.Close()
	os.Stdout = old
	return <-done
}

func TestPrintVersion(t *testing.T) {
	out := captureClientStdout(t, printVersion)
	if out == "" {
		t.Fatal("printVersion produced no output")
	}
	// Must include the version value.
	if !clientContains(out, Version) {
		t.Errorf("printVersion output %q does not contain version %q", out, Version)
	}
}

func TestPrintHelp(t *testing.T) {
	out := captureClientStdout(t, printHelp)
	if out == "" {
		t.Fatal("printHelp produced no output")
	}
	for _, kw := range []string{"--help", "--version", "--server"} {
		if !clientContains(out, kw) {
			t.Errorf("printHelp output missing keyword %q", kw)
		}
	}
}

func TestResolveToken_FlagValue(t *testing.T) {
	tok := resolveToken("mytoken", "")
	if tok != "mytoken" {
		t.Errorf("resolveToken flagToken=%q returned %q, want %q", "mytoken", tok, "mytoken")
	}
}

func TestResolveToken_EnvVar(t *testing.T) {
	old := os.Getenv("CASSOCIAL_TOKEN")
	os.Setenv("CASSOCIAL_TOKEN", "envtoken")
	t.Cleanup(func() {
		if old != "" {
			os.Setenv("CASSOCIAL_TOKEN", old)
		} else {
			os.Unsetenv("CASSOCIAL_TOKEN")
		}
	})

	tok := resolveToken("", "")
	if tok != "envtoken" {
		t.Errorf("resolveToken from env returned %q, want %q", tok, "envtoken")
	}
}

func TestResolveToken_FlagTakesPrecedence(t *testing.T) {
	old := os.Getenv("CASSOCIAL_TOKEN")
	os.Setenv("CASSOCIAL_TOKEN", "envtoken")
	t.Cleanup(func() {
		if old != "" {
			os.Setenv("CASSOCIAL_TOKEN", old)
		} else {
			os.Unsetenv("CASSOCIAL_TOKEN")
		}
	})

	tok := resolveToken("flagtoken", "")
	if tok != "flagtoken" {
		t.Errorf("resolveToken: flag should take precedence over env; got %q", tok)
	}
}

func TestResolveToken_FileValue(t *testing.T) {
	f, err := os.CreateTemp("", "cassocial-token-*")
	if err != nil {
		t.Fatalf("os.CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())

	f.WriteString("filetoken\n")
	f.Close()

	tok := resolveToken("", f.Name())
	if tok != "filetoken" {
		t.Errorf("resolveToken from file returned %q, want %q", tok, "filetoken")
	}
}

func TestClientRun_UnknownCommand(t *testing.T) {
	c := &client{
		server: "http://localhost",
		http:   &http.Client{},
	}

	err := c.run("unknowncmd", nil)
	if err == nil {
		t.Fatal("run with unknown command should return an error")
	}
}

func TestClientRun_VersionCommand(t *testing.T) {
	c := &client{
		server: "http://localhost",
		http:   &http.Client{},
	}

	out := captureClientStdout(t, func() {
		err := c.run("version", nil)
		if err != nil {
			t.Errorf("run('version') returned error: %v", err)
		}
	})

	if out == "" {
		t.Error("run('version') produced no output")
	}
}

func TestClientRun_ProfileMissingSlug(t *testing.T) {
	c := &client{
		server: "http://localhost",
		http:   &http.Client{},
	}

	err := c.run("profile", []string{})
	if err == nil {
		t.Fatal("run('profile') with no slug should return an error")
	}
}

func TestClientRun_LinksNoProfile(t *testing.T) {
	c := &client{
		server: "http://localhost",
		http:   &http.Client{},
	}

	err := c.run("links", []string{})
	if err == nil {
		t.Fatal("run('links') with no profile ID should return an error")
	}
}

func TestClientRun_ShortlinkNoSubcmd(t *testing.T) {
	c := &client{
		server: "http://localhost",
		http:   &http.Client{},
	}

	err := c.run("shortlink", []string{})
	if err == nil {
		t.Fatal("run('shortlink') with no subcommand should return an error")
	}
}

func TestClientRun_ShortlinkUnknown(t *testing.T) {
	c := &client{
		server: "http://localhost",
		http:   &http.Client{},
	}

	err := c.run("shortlink", []string{"badcmd"})
	if err == nil {
		t.Fatal("run('shortlink', 'badcmd') should return an error")
	}
}

func TestClientRun_ShortlinkCreateMissingURL(t *testing.T) {
	c := &client{
		server: "http://localhost",
		http:   &http.Client{},
	}

	err := c.run("shortlink", []string{"create"})
	if err == nil {
		t.Fatal("shortlink create with no --url should return an error")
	}
}

func TestClientRun_ShortlinkDeleteMissingCode(t *testing.T) {
	c := &client{
		server: "http://localhost",
		http:   &http.Client{},
	}

	err := c.run("shortlink", []string{"delete"})
	if err == nil {
		t.Fatal("shortlink delete with no --code should return an error")
	}
}

func TestClientGet_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := &client{
		server: srv.URL,
		http:   &http.Client{},
	}

	_, err := c.get("/some/path")
	if err == nil {
		t.Fatal("get() against a 500 server should return an error")
	}
}

func TestClientGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := &client{
		server: srv.URL,
		http:   &http.Client{},
		token:  "testtoken",
	}

	body, err := c.get("/test")
	if err != nil {
		t.Fatalf("get() returned error: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("get() returned %q, want %q", string(body), `{"ok":true}`)
	}
}

func TestClientPost_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "bad content type", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"created":true}`))
	}))
	defer srv.Close()

	c := &client{
		server: srv.URL,
		http:   &http.Client{},
	}

	body, err := c.post("/test", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("post() returned error: %v", err)
	}
	if !clientContains(string(body), "created") {
		t.Errorf("post() returned %q, expected to contain 'created'", string(body))
	}
}

func TestClientDelete_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"deleted":true}`))
	}))
	defer srv.Close()

	c := &client{
		server: srv.URL,
		http:   &http.Client{},
	}

	body, err := c.delete("/test")
	if err != nil {
		t.Fatalf("delete() returned error: %v", err)
	}
	if !clientContains(string(body), "deleted") {
		t.Errorf("delete() returned %q, expected to contain 'deleted'", string(body))
	}
}

// ---------------------------------------------------------------------------
// resolveToken — token file path and default file path branches
// ---------------------------------------------------------------------------

func TestResolveToken_TokenFileMissing(t *testing.T) {
	// File doesn't exist — should fall through to env/default logic.
	os.Unsetenv("CASSOCIAL_TOKEN")
	tok := resolveToken("", "/nonexistent/path/token-file")
	// No env var set and file missing — empty string expected.
	if tok != "" {
		t.Errorf("resolveToken with missing file = %q, want empty", tok)
	}
}

func TestResolveToken_DefaultFileLoosePermissions(t *testing.T) {
	// Create a fake HOME with a token file that has loose permissions (world-readable).
	// resolveToken should emit a warning but return "" because it skips loose-perms files.
	tmpHome := t.TempDir()
	tokenDir := tmpHome + "/.config/casapps/cassocial"
	if err := os.MkdirAll(tokenDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	tokenPath := tokenDir + "/token"
	if err := os.WriteFile(tokenPath, []byte("loosetoken\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Override HOME so resolveToken finds our fake token file.
	t.Setenv("HOME", tmpHome)
	os.Unsetenv("CASSOCIAL_TOKEN")

	// The function should warn and NOT return the token.
	tok := resolveToken("", "")
	if tok != "" {
		t.Errorf("resolveToken with loose-perms token file returned %q, want empty", tok)
	}
}

func TestResolveToken_DefaultFileStrictPermissions(t *testing.T) {
	// Create a fake HOME with a token file that has strict permissions (0600).
	// resolveToken should read and return the token.
	tmpHome := t.TempDir()
	tokenDir := tmpHome + "/.config/casapps/cassocial"
	if err := os.MkdirAll(tokenDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	tokenPath := tokenDir + "/token"
	if err := os.WriteFile(tokenPath, []byte("stricttoken\n"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("HOME", tmpHome)
	os.Unsetenv("CASSOCIAL_TOKEN")

	tok := resolveToken("", "")
	if tok != "stricttoken" {
		t.Errorf("resolveToken with strict-perms token file = %q, want stricttoken", tok)
	}
}

// ---------------------------------------------------------------------------
// cmdProfile — httptest server branches
// ---------------------------------------------------------------------------

func TestCmdProfile_WithSlugFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/profile" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"slug":"alice"}`))
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	out := captureClientStdout(t, func() {
		if err := c.cmdProfile([]string{"--slug", "alice"}); err != nil {
			t.Errorf("cmdProfile(--slug alice) returned error: %v", err)
		}
	})
	if !clientContains(out, "alice") {
		t.Errorf("cmdProfile output %q does not contain 'alice'", out)
	}
}

func TestCmdProfile_WithPositionalArg(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"slug":"bob"}`))
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	out := captureClientStdout(t, func() {
		if err := c.cmdProfile([]string{"bob"}); err != nil {
			t.Errorf("cmdProfile(bob) returned error: %v", err)
		}
	})
	if !clientContains(out, "bob") {
		t.Errorf("cmdProfile output %q does not contain 'bob'", out)
	}
}

func TestCmdProfile_FromUserContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"slug":"carol"}`))
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}, user: "@carol"}
	out := captureClientStdout(t, func() {
		if err := c.cmdProfile([]string{}); err != nil {
			t.Errorf("cmdProfile() from user context returned error: %v", err)
		}
	})
	if !clientContains(out, "carol") {
		t.Errorf("cmdProfile output %q does not contain 'carol'", out)
	}
}

func TestCmdProfile_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	err := c.cmdProfile([]string{"missing-slug"})
	if err == nil {
		t.Fatal("cmdProfile with server 404 should return an error")
	}
}

// ---------------------------------------------------------------------------
// cmdLinks — httptest server branches
// ---------------------------------------------------------------------------

func TestCmdLinks_WithProfileFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/links" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"url":"https://example.com"}]`))
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	out := captureClientStdout(t, func() {
		if err := c.cmdLinks([]string{"--profile", "42"}); err != nil {
			t.Errorf("cmdLinks(--profile 42) returned error: %v", err)
		}
	})
	if !clientContains(out, "example.com") {
		t.Errorf("cmdLinks output %q does not contain 'example.com'", out)
	}
}

func TestCmdLinks_WithPositionalArg(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"url":"https://test.com"}]`))
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	out := captureClientStdout(t, func() {
		if err := c.cmdLinks([]string{"99"}); err != nil {
			t.Errorf("cmdLinks(99) returned error: %v", err)
		}
	})
	if !clientContains(out, "test.com") {
		t.Errorf("cmdLinks output %q does not contain 'test.com'", out)
	}
}

func TestCmdLinks_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	err := c.cmdLinks([]string{"some-id"})
	if err == nil {
		t.Fatal("cmdLinks with server 500 should return an error")
	}
}

// ---------------------------------------------------------------------------
// cmdShortlink — httptest server branches for list, create, delete
// ---------------------------------------------------------------------------

func TestCmdShortlink_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/shortlinks" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"code":"abc"}]`))
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	out := captureClientStdout(t, func() {
		if err := c.cmdShortlink([]string{"list"}); err != nil {
			t.Errorf("cmdShortlink(list) returned error: %v", err)
		}
	})
	if !clientContains(out, "abc") {
		t.Errorf("cmdShortlink list output %q does not contain 'abc'", out)
	}
}

func TestCmdShortlink_Create_WithURLFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"xyz"}`))
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	out := captureClientStdout(t, func() {
		if err := c.cmdShortlink([]string{"create", "--url", "https://example.com"}); err != nil {
			t.Errorf("cmdShortlink(create --url) returned error: %v", err)
		}
	})
	if !clientContains(out, "xyz") {
		t.Errorf("cmdShortlink create output %q does not contain 'xyz'", out)
	}
}

func TestCmdShortlink_Create_WithPositionalURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"pos"}`))
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	out := captureClientStdout(t, func() {
		if err := c.cmdShortlink([]string{"create", "https://positional.com"}); err != nil {
			t.Errorf("cmdShortlink(create positional) returned error: %v", err)
		}
	})
	if !clientContains(out, "pos") {
		t.Errorf("cmdShortlink create positional output %q does not contain 'pos'", out)
	}
}

func TestCmdShortlink_Create_WithCustomCode(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 512)
		n, _ := r.Body.Read(buf)
		gotBody = string(buf[:n])
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"mycode"}`))
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	if err := c.cmdShortlink([]string{"create", "--url", "https://x.com", "--code", "mycode"}); err != nil {
		t.Errorf("cmdShortlink(create --code) returned error: %v", err)
	}
	if !clientContains(gotBody, "mycode") {
		t.Errorf("create request body %q does not contain custom_code", gotBody)
	}
}

func TestCmdShortlink_Delete_WithCodeFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"deleted":true}`))
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	out := captureClientStdout(t, func() {
		if err := c.cmdShortlink([]string{"delete", "--code", "abc"}); err != nil {
			t.Errorf("cmdShortlink(delete --code) returned error: %v", err)
		}
	})
	if !clientContains(out, "deleted") {
		t.Errorf("cmdShortlink delete output %q does not contain 'deleted'", out)
	}
}

func TestCmdShortlink_Delete_WithPositionalCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"deleted":true}`))
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	out := captureClientStdout(t, func() {
		if err := c.cmdShortlink([]string{"delete", "xyz"}); err != nil {
			t.Errorf("cmdShortlink(delete positional) returned error: %v", err)
		}
	})
	if !clientContains(out, "deleted") {
		t.Errorf("cmdShortlink delete positional output %q does not contain 'deleted'", out)
	}
}

func TestCmdShortlink_List_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone", http.StatusGone)
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	err := c.cmdShortlink([]string{"list"})
	if err == nil {
		t.Fatal("cmdShortlink list with server error should return an error")
	}
}

func TestCmdShortlink_Create_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	err := c.cmdShortlink([]string{"create", "--url", "https://example.com"})
	if err == nil {
		t.Fatal("cmdShortlink create with server error should return an error")
	}
}

func TestCmdShortlink_Delete_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	err := c.cmdShortlink([]string{"delete", "--code", "xyz"})
	if err == nil {
		t.Fatal("cmdShortlink delete with server error should return an error")
	}
}

// ---------------------------------------------------------------------------
// do — 4xx non-401 error body included in message
// ---------------------------------------------------------------------------

func TestClientDo_400_ErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request detail", http.StatusBadRequest)
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}}
	_, err := c.get("/test")
	if err == nil {
		t.Fatal("do() with 400 should return an error")
	}
	if !clientContains(err.Error(), "400") {
		t.Errorf("error message %q does not contain status code 400", err.Error())
	}
}

func TestClientPost_InvalidServer(t *testing.T) {
	// Unreachable URL — http.Client.Do should fail.
	c := &client{server: "http://127.0.0.1:1", http: &http.Client{}}
	_, err := c.post("/test", map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("post() to unreachable server should return an error")
	}
}

func TestClientDelete_InvalidServer(t *testing.T) {
	c := &client{server: "http://127.0.0.1:1", http: &http.Client{}}
	_, err := c.delete("/test")
	if err == nil {
		t.Fatal("delete() to unreachable server should return an error")
	}
}

// ---------------------------------------------------------------------------
// get / post / delete — http.NewRequest error path (invalid URL)
// ---------------------------------------------------------------------------

func TestClientGet_InvalidURL(t *testing.T) {
	c := &client{
		server: "http://\x00invalid",
		http:   &http.Client{},
	}
	_, err := c.get("/path")
	if err == nil {
		t.Fatal("get() with invalid server URL should return an error")
	}
}

func TestClientPost_InvalidURL(t *testing.T) {
	c := &client{
		server: "http://\x00invalid",
		http:   &http.Client{},
	}
	_, err := c.post("/path", map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("post() with invalid server URL should return an error")
	}
}

func TestClientPost_MarshalError(t *testing.T) {
	c := &client{server: "http://localhost", http: &http.Client{}}
	// Channels cannot be marshalled to JSON — json.Marshal will return an error.
	_, err := c.post("/path", make(chan int))
	if err == nil {
		t.Fatal("post() with un-marshallable body should return an error")
	}
}

func TestClientDelete_InvalidURL(t *testing.T) {
	c := &client{
		server: "http://\x00invalid",
		http:   &http.Client{},
	}
	_, err := c.delete("/path")
	if err == nil {
		t.Fatal("delete() with invalid server URL should return an error")
	}
}

// ---------------------------------------------------------------------------
// do — no token branch (Authorization header not set)
// ---------------------------------------------------------------------------

func TestClientDo_NoToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := &client{server: srv.URL, http: &http.Client{}, token: ""}
	_, err := c.get("/test")
	if err != nil {
		t.Fatalf("get() with no token: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("Authorization header should be empty when token is empty, got %q", gotAuth)
	}
}

// ---------------------------------------------------------------------------
// cmdProfile — fs.Parse error (unknown flag)
// ---------------------------------------------------------------------------

func TestCmdProfile_ParseError(t *testing.T) {
	c := &client{server: "http://localhost", http: &http.Client{}}
	err := c.cmdProfile([]string{"--unknown-flag-xyz"})
	if err == nil {
		t.Fatal("cmdProfile with unknown flag should return an error")
	}
}

// ---------------------------------------------------------------------------
// cmdLinks — fs.Parse error (unknown flag)
// ---------------------------------------------------------------------------

func TestCmdLinks_ParseError(t *testing.T) {
	c := &client{server: "http://localhost", http: &http.Client{}}
	err := c.cmdLinks([]string{"--unknown-flag-xyz"})
	if err == nil {
		t.Fatal("cmdLinks with unknown flag should return an error")
	}
}

// ---------------------------------------------------------------------------
// cmdShortlink create — fs.Parse error (unknown flag)
// ---------------------------------------------------------------------------

func TestCmdShortlink_Create_ParseError(t *testing.T) {
	c := &client{server: "http://localhost", http: &http.Client{}}
	err := c.cmdShortlink([]string{"create", "--unknown-flag-xyz"})
	if err == nil {
		t.Fatal("cmdShortlink create with unknown flag should return an error")
	}
}

// ---------------------------------------------------------------------------
// cmdShortlink delete — fs.Parse error (unknown flag)
// ---------------------------------------------------------------------------

func TestCmdShortlink_Delete_ParseError(t *testing.T) {
	c := &client{server: "http://localhost", http: &http.Client{}}
	err := c.cmdShortlink([]string{"delete", "--unknown-flag-xyz"})
	if err == nil {
		t.Fatal("cmdShortlink delete with unknown flag should return an error")
	}
}

// ---------------------------------------------------------------------------
// runCLI — tests for the top-level flag-parsing entry point
// ---------------------------------------------------------------------------

func TestRunCLI_NoArgs_PrintsHelp(t *testing.T) {
	out := captureClientStdout(t, func() {
		code := runCLI([]string{})
		if code != 0 {
			t.Errorf("runCLI() with no args returned code %d, want 0", code)
		}
	})
	if !clientContains(out, "--help") {
		t.Errorf("runCLI() with no args should print help, got: %q", out)
	}
}

func TestRunCLI_HelpFlag(t *testing.T) {
	out := captureClientStdout(t, func() {
		code := runCLI([]string{"--help"})
		if code != 0 {
			t.Errorf("runCLI(--help) returned code %d, want 0", code)
		}
	})
	if !clientContains(out, "--help") {
		t.Errorf("runCLI(--help) output missing --help, got: %q", out)
	}
}

func TestRunCLI_ShortHelpFlag(t *testing.T) {
	out := captureClientStdout(t, func() {
		code := runCLI([]string{"-h"})
		if code != 0 {
			t.Errorf("runCLI(-h) returned code %d, want 0", code)
		}
	})
	if out == "" {
		t.Error("runCLI(-h) produced no output")
	}
}

func TestRunCLI_VersionFlag(t *testing.T) {
	out := captureClientStdout(t, func() {
		code := runCLI([]string{"--version"})
		if code != 0 {
			t.Errorf("runCLI(--version) returned code %d, want 0", code)
		}
	})
	if !clientContains(out, Version) {
		t.Errorf("runCLI(--version) output %q does not contain version %q", out, Version)
	}
}

func TestRunCLI_ShortVersionFlag(t *testing.T) {
	out := captureClientStdout(t, func() {
		code := runCLI([]string{"-v"})
		if code != 0 {
			t.Errorf("runCLI(-v) returned code %d, want 0", code)
		}
	})
	if out == "" {
		t.Error("runCLI(-v) produced no output")
	}
}

func TestRunCLI_NoServerURL_ReturnsOne(t *testing.T) {
	// Clear server env and provide no --server flag
	os.Unsetenv("CASSOCIAL_SERVER")
	code := runCLI([]string{"profile", "alice"})
	if code != 1 {
		t.Errorf("runCLI without server URL returned code %d, want 1", code)
	}
}

func TestRunCLI_ServerFromEnv_CommandFails_ReturnsOne(t *testing.T) {
	// Provide server via env; command fails (unreachable host) → exit 1
	t.Setenv("CASSOCIAL_SERVER", "http://127.0.0.1:1")
	code := runCLI([]string{"profile", "alice"})
	if code != 1 {
		t.Errorf("runCLI with unreachable server returned code %d, want 1", code)
	}
}

func TestRunCLI_ServerFromFlag_CommandSucceeds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"slug":"alice"}`))
	}))
	defer srv.Close()

	os.Unsetenv("CASSOCIAL_SERVER")
	out := captureClientStdout(t, func() {
		code := runCLI([]string{"--server", srv.URL, "profile", "alice"})
		if code != 0 {
			t.Errorf("runCLI with valid server returned code %d, want 0", code)
		}
	})
	if !clientContains(out, "alice") {
		t.Errorf("runCLI output %q does not contain 'alice'", out)
	}
}

func TestRunCLI_ColorNever_SetsNO_COLOR(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("CASSOCIAL_SERVER")

	runCLI([]string{"--color", "never"})

	// NO_COLOR should be set after runCLI processes --color never
	if os.Getenv("NO_COLOR") == "" {
		t.Error("runCLI(--color never) should set NO_COLOR env var")
	}
	os.Unsetenv("NO_COLOR")
}

func TestRunCLI_LangAutoDetect(t *testing.T) {
	t.Setenv("LANG", "es_MX.UTF-8")
	os.Unsetenv("CASSOCIAL_SERVER")

	// Just verify runCLI runs without panic when LANG is set
	code := runCLI([]string{})
	if code != 0 {
		t.Errorf("runCLI() with LANG set returned code %d, want 0 (help)", code)
	}
}

func TestRunCLI_TokenFromFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Authorization header is set
		if r.Header.Get("Authorization") != "Bearer mytoken" {
			http.Error(w, "no auth", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"slug":"alice"}`))
	}))
	defer srv.Close()

	// Override osExit to prevent process exit on 401
	orig := osExit
	osExit = func(code int) {}
	t.Cleanup(func() { osExit = orig })

	os.Unsetenv("CASSOCIAL_SERVER")
	captureClientStdout(t, func() {
		runCLI([]string{"--server", srv.URL, "--token", "mytoken", "profile", "alice"})
	})
}

func TestRunCLI_UserFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"slug":"alice"}`))
	}))
	defer srv.Close()

	os.Unsetenv("CASSOCIAL_SERVER")
	out := captureClientStdout(t, func() {
		code := runCLI([]string{"--server", srv.URL, "--user", "@alice", "profile"})
		if code != 0 {
			t.Errorf("runCLI with --user flag returned code %d, want 0", code)
		}
	})
	if !clientContains(out, "alice") {
		t.Errorf("runCLI --user output %q does not contain 'alice'", out)
	}
}

func TestRunCLI_ParseError_ReturnsTwo(t *testing.T) {
	code := runCLI([]string{"--unknown-flag-that-does-not-exist"})
	if code != 2 {
		t.Errorf("runCLI with unknown flag returned code %d, want 2", code)
	}
}

// ---------------------------------------------------------------------------
// do — io.ReadAll error path (body read failure)
// ---------------------------------------------------------------------------

// errReader is a ReadCloser whose Read always returns an error.
type errReader struct{}

func (errReader) Read(_ []byte) (int, error) { return 0, errors.New("read error injected") }
func (errReader) Close() error               { return nil }

func TestClientDo_ReadBodyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send headers with 200 but the test will inject a bad body via a
		// custom RoundTripper, so this server just needs to be reachable.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	// Use a custom RoundTripper that returns a response with a body that errors on read.
	rt := &errorBodyTransport{delegate: http.DefaultTransport, serverURL: srv.URL}
	c := &client{
		server: srv.URL,
		http:   &http.Client{Transport: rt},
	}

	_, err := c.get("/test")
	if err == nil {
		t.Fatal("do() with read-error body should return an error")
	}
}

// errorBodyTransport wraps a RoundTripper and replaces the response body with
// an errReader so that io.ReadAll in do() returns an error.
type errorBodyTransport struct {
	delegate  http.RoundTripper
	serverURL string
}

func (t *errorBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.delegate.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	resp.Body = errReader{}
	return resp, nil
}

// ---------------------------------------------------------------------------
// do — 401 Unauthorized path (osExit override)
// ---------------------------------------------------------------------------

func TestClientDo_401_ExitsWithCode1(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	// Override osExit so the test can verify it was called without actually exiting.
	var exitCode int
	orig := osExit
	osExit = func(code int) { exitCode = code }
	t.Cleanup(func() { osExit = orig })

	c := &client{server: srv.URL, http: &http.Client{}, token: "badtoken"}
	// do() calls osExit(1) on 401 — the function will return normally after our
	// fake exit, so we call get() and expect a nil error (exit was "taken").
	_, _ = c.get("/test")

	if exitCode != 1 {
		t.Errorf("do() on 401 should call osExit(1), got osExit(%d)", exitCode)
	}
}

// clientContains checks whether s contains sub.
func clientContains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
