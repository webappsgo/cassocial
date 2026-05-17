package model

import (
	"testing"
)

func validService() *Service {
	return &Service{
		Name:     "GitHub",
		Category: CategoryDevelopment,
		IsActive: true,
	}
}

func TestService_Validate_Valid(t *testing.T) {
	s := validService()
	if err := s.Validate(); err != nil {
		t.Errorf("valid service Validate() = %v, want nil", err)
	}
}

func TestService_Validate_EmptyName(t *testing.T) {
	s := validService()
	s.Name = ""
	if err := s.Validate(); err != ErrServiceNameEmpty {
		t.Errorf("empty name Validate() = %v, want ErrServiceNameEmpty", err)
	}
}

func TestService_Validate_InvalidCategory(t *testing.T) {
	s := validService()
	s.Category = "invalid-cat"
	if err := s.Validate(); err != ErrInvalidCategory {
		t.Errorf("invalid category Validate() = %v, want ErrInvalidCategory", err)
	}
}

func TestService_Validate_AllCategories(t *testing.T) {
	for _, cat := range AllCategories() {
		s := &Service{Name: "Test", Category: cat}
		if err := s.Validate(); err != nil {
			t.Errorf("category %q Validate() = %v, want nil", cat, err)
		}
	}
}

func TestAllCategories_Count(t *testing.T) {
	cats := AllCategories()
	if len(cats) == 0 {
		t.Error("AllCategories() returned empty slice")
	}
}

func TestService_IncrementPopularity(t *testing.T) {
	s := validService()
	s.Popularity = 5
	s.IncrementPopularity()
	if s.Popularity != 6 {
		t.Errorf("IncrementPopularity() = %d, want 6", s.Popularity)
	}
}

func TestService_BuildURL_WithUsername(t *testing.T) {
	s := &Service{
		URLPattern:       "https://github.com/{username}",
		RequiresUsername: true,
	}
	got := s.BuildURL("alice")
	if got != "https://github.com/alice" {
		t.Errorf("BuildURL(alice) = %q, want https://github.com/alice", got)
	}
}

func TestService_BuildURL_EmptyPattern(t *testing.T) {
	s := &Service{URLPattern: ""}
	got := s.BuildURL("alice")
	if got != "" {
		t.Errorf("BuildURL with empty pattern = %q, want empty", got)
	}
}

func TestService_BuildURL_NoUsername(t *testing.T) {
	s := &Service{
		URLPattern:       "https://github.com/{username}",
		RequiresUsername: false,
	}
	// When RequiresUsername is false, no replacement happens
	got := s.BuildURL("alice")
	// Should return the pattern as-is (no replacement because RequiresUsername=false)
	if got == "" {
		t.Error("BuildURL should return the pattern even without username replacement")
	}
}

func TestService_GetPlaceholder_Custom(t *testing.T) {
	s := &Service{PlaceholderText: "Enter handle"}
	if got := s.GetPlaceholder(); got != "Enter handle" {
		t.Errorf("GetPlaceholder() = %q, want 'Enter handle'", got)
	}
}

func TestService_GetPlaceholder_RequiresUsername(t *testing.T) {
	s := &Service{Name: "Twitter", RequiresUsername: true}
	got := s.GetPlaceholder()
	if got == "" {
		t.Error("GetPlaceholder() with RequiresUsername returned empty")
	}
}

func TestService_GetPlaceholder_Default(t *testing.T) {
	s := &Service{Name: "Custom", RequiresUsername: false}
	got := s.GetPlaceholder()
	if got != "Enter URL" {
		t.Errorf("GetPlaceholder() = %q, want 'Enter URL'", got)
	}
}
