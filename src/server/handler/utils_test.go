package handler

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Tests for replaceQuestionMarks — PostgreSQL placeholder conversion.
// The function replaces ? with $N by iterating count down to 1, substituting
// the first occurrence each time.  With count=2 on "a = ? AND b = ?" the
// loop first replaces the first ? with $2, then the first remaining ? with $1,
// yielding "a = $2 AND b = $1" — this is the documented backward-pass behaviour.
func TestReplaceQuestionMarks(t *testing.T) {
	tests := []struct {
		query string
		count int
		want  string
	}{
		// backward pass: first ? becomes $2, second ? becomes $1
		{"SELECT * FROM t WHERE a = ? AND b = ?", 2, "SELECT * FROM t WHERE a = $2 AND b = $1"},
		{"SELECT 1", 0, "SELECT 1"},
		{"WHERE x = ?", 1, "WHERE x = $1"},
	}
	for _, tt := range tests {
		got := replaceQuestionMarks(tt.query, tt.count)
		if got != tt.want {
			t.Errorf("replaceQuestionMarks(%q, %d) = %q, want %q", tt.query, tt.count, got, tt.want)
		}
	}
}

// Tests for replaceQuestionMarksWithArgs — sequential forward replacement.
func TestReplaceQuestionMarksWithArgs(t *testing.T) {
	tests := []struct {
		query    string
		argCount int
		want     string
	}{
		{"SELECT * FROM t WHERE a = ? AND b = ?", 2, "SELECT * FROM t WHERE a = $1 AND b = $2"},
		{"no placeholders", 0, "no placeholders"},
		{"WHERE x = ?", 1, "WHERE x = $1"},
	}
	for _, tt := range tests {
		got := replaceQuestionMarksWithArgs(tt.query, tt.argCount)
		if got != tt.want {
			t.Errorf("replaceQuestionMarksWithArgs(%q, %d) = %q, want %q", tt.query, tt.argCount, got, tt.want)
		}
	}
}

// Tests for validateEmail.
func TestValidateEmail(t *testing.T) {
	valid := []string{
		"user@example.com",
		"a@b.io",
		"user+tag@sub.domain.org",
	}
	for _, e := range valid {
		if !validateEmail(e) {
			t.Errorf("validateEmail(%q) = false, want true", e)
		}
	}

	invalid := []string{
		"",
		"notanemail",
		"missing-at.com",
		"@nodomain",
	}
	for _, e := range invalid {
		if validateEmail(e) {
			t.Errorf("validateEmail(%q) = true, want false", e)
		}
	}
}

// Tests for validateURL.
func TestValidateURL(t *testing.T) {
	valid := []string{
		"http://example.com",
		"https://example.com/path?q=1",
		"https://localhost:8080",
	}
	for _, u := range valid {
		if !validateURL(u) {
			t.Errorf("validateURL(%q) = false, want true", u)
		}
	}

	invalid := []string{
		"",
		"ftp://files.example.com",
		"example.com",
		"//example.com",
	}
	for _, u := range invalid {
		if validateURL(u) {
			t.Errorf("validateURL(%q) = true, want false", u)
		}
	}
}

// Tests for sanitizeString.
func TestSanitizeString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  hello  ", "hello"},
		{"hello\x00world", "helloworld"},
		{"\x00\x00", ""},
		{"normal string", "normal string"},
		{"", ""},
	}
	for _, tt := range tests {
		got := sanitizeString(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// Tests for truncateString — boundary conditions at exact length, over, and empty.
func TestTruncateString(t *testing.T) {
	tests := []struct {
		s         string
		maxLength int
		want      string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 3, "hel"},
		{"", 5, ""},
		{"a", 0, ""},
	}
	for _, tt := range tests {
		got := truncateString(tt.s, tt.maxLength)
		if got != tt.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.s, tt.maxLength, got, tt.want)
		}
	}
}

// Tests for pointer helpers.
func TestBoolPtr(t *testing.T) {
	b := boolPtr(true)
	if b == nil || !*b {
		t.Errorf("boolPtr(true) = %v, want ptr to true", b)
	}
	b2 := boolPtr(false)
	if b2 == nil || *b2 {
		t.Errorf("boolPtr(false) = %v, want ptr to false", b2)
	}
}

func TestStringPtr(t *testing.T) {
	s := stringPtr("hello")
	if s == nil || *s != "hello" {
		t.Errorf("stringPtr(\"hello\") = %v, want ptr to \"hello\"", s)
	}
	s2 := stringPtr("")
	if s2 == nil || *s2 != "" {
		t.Errorf("stringPtr(\"\") = %v, want ptr to empty string", s2)
	}
}

func TestIntPtr(t *testing.T) {
	p := intPtr(42)
	if p == nil || *p != 42 {
		t.Errorf("intPtr(42) = %v, want ptr to 42", p)
	}
	p2 := intPtr(0)
	if p2 == nil || *p2 != 0 {
		t.Errorf("intPtr(0) = %v, want ptr to 0", p2)
	}
	p3 := intPtr(-1)
	if p3 == nil || *p3 != -1 {
		t.Errorf("intPtr(-1) = %v, want ptr to -1", p3)
	}
}

func TestTimePtr(t *testing.T) {
	now := time.Now()
	p := timePtr(now)
	if p == nil || !p.Equal(now) {
		t.Errorf("timePtr did not return correct time")
	}
}

// Tests for contains.
func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	if !contains(slice, "a") {
		t.Error("contains([a,b,c], \"a\") = false, want true")
	}
	if !contains(slice, "c") {
		t.Error("contains([a,b,c], \"c\") = false, want true")
	}
	if contains(slice, "z") {
		t.Error("contains([a,b,c], \"z\") = true, want false")
	}
	if contains([]string{}, "a") {
		t.Error("contains([], \"a\") = true, want false")
	}
}

// Tests for unique.
func TestUnique(t *testing.T) {
	tests := []struct {
		input []string
		wantN int
	}{
		{[]string{"a", "b", "a", "c", "b"}, 3},
		{[]string{}, 0},
		{[]string{"x"}, 1},
		{[]string{"x", "x", "x"}, 1},
	}
	for _, tt := range tests {
		got := unique(tt.input)
		if len(got) != tt.wantN {
			t.Errorf("unique(%v) = %v (len %d), want len %d", tt.input, got, len(got), tt.wantN)
		}
		seen := map[string]bool{}
		for _, v := range got {
			if seen[v] {
				t.Errorf("unique(%v) returned duplicate %q", tt.input, v)
			}
			seen[v] = true
		}
	}
}

// Tests for parseBool — covers all recognised truthy values and some falsy ones.
func TestParseBool(t *testing.T) {
	truthy := []string{"true", "True", "TRUE", "1", "yes", "Yes", "on", "On", " true ", " 1 "}
	for _, s := range truthy {
		if !parseBool(s) {
			t.Errorf("parseBool(%q) = false, want true", s)
		}
	}
	falsy := []string{"false", "0", "no", "off", "", "maybe", "2"}
	for _, s := range falsy {
		if parseBool(s) {
			t.Errorf("parseBool(%q) = true, want false", s)
		}
	}
}

// Tests for parseInt.
func TestParseInt(t *testing.T) {
	tests := []struct {
		input        string
		defaultValue int
		want         int
	}{
		{"42", 0, 42},
		{"-7", 0, -7},
		{"0", 99, 0},
		{"abc", 5, 5},
		{"", 3, 3},
	}
	for _, tt := range tests {
		got := parseInt(tt.input, tt.defaultValue)
		if got != tt.want {
			t.Errorf("parseInt(%q, %d) = %d, want %d", tt.input, tt.defaultValue, got, tt.want)
		}
	}
}

// Tests for parseFloat.
func TestParseFloat(t *testing.T) {
	tests := []struct {
		input        string
		defaultValue float64
		want         float64
	}{
		{"3.14", 0.0, 3.14},
		{"-1.5", 0.0, -1.5},
		{"0", 9.9, 0.0},
		{"abc", 2.5, 2.5},
		{"", 1.1, 1.1},
	}
	for _, tt := range tests {
		got := parseFloat(tt.input, tt.defaultValue)
		if got != tt.want {
			t.Errorf("parseFloat(%q, %v) = %v, want %v", tt.input, tt.defaultValue, got, tt.want)
		}
	}
}

// Tests for NewSuccessResponse.
func TestNewSuccessResponse(t *testing.T) {
	resp := NewSuccessResponse("ok", map[string]int{"count": 1})
	if !resp.Success {
		t.Error("NewSuccessResponse.Success = false, want true")
	}
	if resp.Message != "ok" {
		t.Errorf("NewSuccessResponse.Message = %q, want \"ok\"", resp.Message)
	}
	if resp.Error != "" {
		t.Errorf("NewSuccessResponse.Error = %q, want empty", resp.Error)
	}
	if resp.Data == nil {
		t.Error("NewSuccessResponse.Data is nil")
	}
}

// Tests for NewErrorResponse.
func TestNewErrorResponse(t *testing.T) {
	resp := NewErrorResponse("something went wrong")
	if resp.Success {
		t.Error("NewErrorResponse.Success = true, want false")
	}
	if resp.Error != "something went wrong" {
		t.Errorf("NewErrorResponse.Error = %q, want \"something went wrong\"", resp.Error)
	}
	if resp.Message != "" {
		t.Errorf("NewErrorResponse.Message = %q, want empty", resp.Message)
	}
}

// Tests for ParsePagination.
func TestParsePagination(t *testing.T) {
	t.Run("defaults when no query params", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		p := ParsePagination(req)
		if p.Page != 1 {
			t.Errorf("Page = %d, want 1", p.Page)
		}
		if p.PageSize != 20 {
			t.Errorf("PageSize = %d, want 20", p.PageSize)
		}
		if p.Offset != 0 {
			t.Errorf("Offset = %d, want 0", p.Offset)
		}
	})

	t.Run("explicit page and page_size", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?page=3&page_size=10", nil)
		p := ParsePagination(req)
		if p.Page != 3 {
			t.Errorf("Page = %d, want 3", p.Page)
		}
		if p.PageSize != 10 {
			t.Errorf("PageSize = %d, want 10", p.PageSize)
		}
		if p.Offset != 20 {
			t.Errorf("Offset = %d, want 20", p.Offset)
		}
	})

	t.Run("page < 1 clamped to 1", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?page=-5", nil)
		p := ParsePagination(req)
		if p.Page != 1 {
			t.Errorf("Page = %d, want 1 for negative input", p.Page)
		}
	})

	t.Run("page_size > 100 clamped to 100", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?page_size=200", nil)
		p := ParsePagination(req)
		if p.PageSize != 100 {
			t.Errorf("PageSize = %d, want 100 for oversized input", p.PageSize)
		}
	})

	t.Run("page_size 0 clamped to 20", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?page_size=0", nil)
		p := ParsePagination(req)
		if p.PageSize != 20 {
			t.Errorf("PageSize = %d, want 20 for zero input", p.PageSize)
		}
	})
}

// Tests for NewPaginatedResponse.
func TestNewPaginatedResponse(t *testing.T) {
	t.Run("correct total_pages calculation", func(t *testing.T) {
		resp := NewPaginatedResponse([]int{1, 2, 3}, 1, 10, 25)
		if resp.TotalPages != 3 {
			t.Errorf("TotalPages = %d, want 3", resp.TotalPages)
		}
		if resp.TotalItems != 25 {
			t.Errorf("TotalItems = %d, want 25", resp.TotalItems)
		}
	})

	t.Run("exact multiple", func(t *testing.T) {
		resp := NewPaginatedResponse(nil, 2, 10, 20)
		if resp.TotalPages != 2 {
			t.Errorf("TotalPages = %d, want 2", resp.TotalPages)
		}
	})

	t.Run("zero total items", func(t *testing.T) {
		resp := NewPaginatedResponse(nil, 1, 10, 0)
		if resp.TotalPages != 0 {
			t.Errorf("TotalPages = %d, want 0", resp.TotalPages)
		}
	})

	t.Run("fields set correctly", func(t *testing.T) {
		data := []string{"a", "b"}
		resp := NewPaginatedResponse(data, 2, 5, 7)
		if resp.Page != 2 {
			t.Errorf("Page = %d, want 2", resp.Page)
		}
		if resp.PageSize != 5 {
			t.Errorf("PageSize = %d, want 5", resp.PageSize)
		}
	})
}

// Tests for getIPAddress.
func TestGetIPAddress(t *testing.T) {
	t.Run("X-Forwarded-For single IP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		if got := getIPAddress(req); got != "1.2.3.4" {
			t.Errorf("getIPAddress = %q, want \"1.2.3.4\"", got)
		}
	})

	t.Run("X-Forwarded-For list returns first", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2, 10.0.0.3")
		if got := getIPAddress(req); got != "10.0.0.1" {
			t.Errorf("getIPAddress = %q, want \"10.0.0.1\"", got)
		}
	})

	t.Run("X-Real-IP fallback", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Real-IP", "5.6.7.8")
		if got := getIPAddress(req); got != "5.6.7.8" {
			t.Errorf("getIPAddress = %q, want \"5.6.7.8\"", got)
		}
	})

	t.Run("RemoteAddr fallback strips port", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.100:54321"
		if got := getIPAddress(req); got != "192.168.1.100" {
			t.Errorf("getIPAddress = %q, want \"192.168.1.100\"", got)
		}
	})
}

// Tests for detectDeviceType.
func TestDetectDeviceType(t *testing.T) {
	mobile := []string{
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
		"Mozilla/5.0 (Android 10; Mobile)",
		"BlackBerry9900",
	}
	for _, ua := range mobile {
		if got := detectDeviceType(ua); got != "mobile" {
			t.Errorf("detectDeviceType(%q) = %q, want \"mobile\"", ua, got)
		}
	}

	tablets := []string{
		"Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X)",
		"Kindle/3.0",
	}
	for _, ua := range tablets {
		if got := detectDeviceType(ua); got != "tablet" {
			t.Errorf("detectDeviceType(%q) = %q, want \"tablet\"", ua, got)
		}
	}

	desktops := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		"",
	}
	for _, ua := range desktops {
		if got := detectDeviceType(ua); got != "desktop" {
			t.Errorf("detectDeviceType(%q) = %q, want \"desktop\"", ua, got)
		}
	}
}

// Tests for hashIP — ensures output is a hex SHA-256 and is stable.
func TestHashIP(t *testing.T) {
	h1 := hashIP("192.168.1.1")
	h2 := hashIP("192.168.1.1")
	if h1 != h2 {
		t.Error("hashIP is not deterministic")
	}
	if len(h1) != 64 {
		t.Errorf("hashIP length = %d, want 64 (hex SHA-256)", len(h1))
	}
	if !isHex(h1) {
		t.Errorf("hashIP = %q, want hex string", h1)
	}
	// Different IPs must produce different hashes.
	h3 := hashIP("10.0.0.1")
	if h1 == h3 {
		t.Error("hashIP collision between different IPs")
	}
}

func isHex(s string) bool {
	for _, c := range s {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}
