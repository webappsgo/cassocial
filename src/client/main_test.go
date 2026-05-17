package main

import (
	"bytes"
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
