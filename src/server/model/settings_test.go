package model

import (
	"testing"
)

func TestSetting_GetString(t *testing.T) {
	s := &Setting{Key: "key", Value: "hello"}
	if got := s.GetString(); got != "hello" {
		t.Errorf("GetString() = %q, want hello", got)
	}
}

func TestSetting_GetInt_Valid(t *testing.T) {
	s := &Setting{Value: "42"}
	got, err := s.GetInt()
	if err != nil {
		t.Errorf("GetInt() error = %v, want nil", err)
	}
	if got != 42 {
		t.Errorf("GetInt() = %d, want 42", got)
	}
}

func TestSetting_GetInt_Invalid(t *testing.T) {
	s := &Setting{Value: "not-a-number"}
	_, err := s.GetInt()
	if err == nil {
		t.Error("GetInt(invalid) should return error")
	}
}

func TestSetting_GetBool(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
		{"yes", true},
		{"no", false},
	}
	for _, tt := range tests {
		s := &Setting{Value: tt.value}
		got, err := s.GetBool()
		if err != nil {
			t.Errorf("GetBool(%q) error = %v", tt.value, err)
		}
		if got != tt.want {
			t.Errorf("GetBool(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestSetting_GetFloat_Valid(t *testing.T) {
	s := &Setting{Value: "3.14"}
	got, err := s.GetFloat()
	if err != nil {
		t.Errorf("GetFloat() error = %v, want nil", err)
	}
	if got != 3.14 {
		t.Errorf("GetFloat() = %f, want 3.14", got)
	}
}

func TestSetting_GetFloat_Invalid(t *testing.T) {
	s := &Setting{Value: "abc"}
	_, err := s.GetFloat()
	if err == nil {
		t.Error("GetFloat(invalid) should return error")
	}
}

func TestSetting_GetJSON(t *testing.T) {
	s := &Setting{Value: `{"foo":"bar"}`}
	var m map[string]string
	if err := s.GetJSON(&m); err != nil {
		t.Errorf("GetJSON() error = %v, want nil", err)
	}
	if m["foo"] != "bar" {
		t.Errorf("GetJSON() foo = %q, want bar", m["foo"])
	}
}

func TestSetting_GetJSON_Invalid(t *testing.T) {
	s := &Setting{Value: "not-json"}
	var m map[string]string
	if err := s.GetJSON(&m); err == nil {
		t.Error("GetJSON(invalid) should return error")
	}
}

func TestSetting_SetString(t *testing.T) {
	s := &Setting{}
	s.SetString("newval")
	if s.Value != "newval" {
		t.Errorf("SetString() value = %q, want newval", s.Value)
	}
	if s.UpdatedAt.IsZero() {
		t.Error("SetString() should update UpdatedAt")
	}
}

func TestSetting_SetInt(t *testing.T) {
	s := &Setting{}
	s.SetInt(99)
	if s.Value != "99" {
		t.Errorf("SetInt() value = %q, want 99", s.Value)
	}
}

func TestSetting_SetBool(t *testing.T) {
	s := &Setting{}
	s.SetBool(true)
	if s.Value != "true" {
		t.Errorf("SetBool(true) value = %q, want true", s.Value)
	}
	s.SetBool(false)
	if s.Value != "false" {
		t.Errorf("SetBool(false) value = %q, want false", s.Value)
	}
}

func TestSetting_SetFloat(t *testing.T) {
	s := &Setting{}
	s.SetFloat(1.5)
	if s.Value != "1.5" {
		t.Errorf("SetFloat(1.5) value = %q, want 1.5", s.Value)
	}
}

func TestSetting_SetJSON(t *testing.T) {
	s := &Setting{}
	m := map[string]int{"x": 1}
	if err := s.SetJSON(m); err != nil {
		t.Errorf("SetJSON() error = %v, want nil", err)
	}
	if s.Value == "" {
		t.Error("SetJSON() should set Value")
	}
}

func TestSetting_SetJSON_UnmarshalableValue(t *testing.T) {
	s := &Setting{}
	// A channel cannot be marshaled to JSON; json.Marshal must return an error.
	err := s.SetJSON(make(chan int))
	if err == nil {
		t.Error("SetJSON(chan) should return an error for an unmarshalable value")
	}
}

func TestSMTPConfig_Validate_Valid(t *testing.T) {
	sc := &SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		FromAddress: "no-reply@example.com",
	}
	if err := sc.Validate(); err != nil {
		t.Errorf("valid SMTP Validate() = %v, want nil", err)
	}
}

func TestSMTPConfig_Validate_EmptyHost(t *testing.T) {
	sc := &SMTPConfig{Port: 587, FromAddress: "a@b.com"}
	if err := sc.Validate(); err == nil {
		t.Error("empty host should return error")
	}
}

func TestSMTPConfig_Validate_InvalidPort_Zero(t *testing.T) {
	sc := &SMTPConfig{Host: "smtp.example.com", Port: 0, FromAddress: "a@b.com"}
	if err := sc.Validate(); err == nil {
		t.Error("port 0 should return error")
	}
}

func TestSMTPConfig_Validate_InvalidPort_High(t *testing.T) {
	sc := &SMTPConfig{Host: "smtp.example.com", Port: 99999, FromAddress: "a@b.com"}
	if err := sc.Validate(); err == nil {
		t.Error("port >65535 should return error")
	}
}

func TestSMTPConfig_Validate_EmptyFromAddress(t *testing.T) {
	sc := &SMTPConfig{Host: "smtp.example.com", Port: 25}
	if err := sc.Validate(); err == nil {
		t.Error("empty from address should return error")
	}
}

func TestSMTPConfig_Validate_UserWithoutPassword(t *testing.T) {
	sc := &SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		FromAddress: "a@b.com",
		User:        "user",
		Password:    "",
	}
	if err := sc.Validate(); err == nil {
		t.Error("user without password should return error")
	}
}

func TestSMTPConfig_Validate_UserWithPassword(t *testing.T) {
	sc := &SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		FromAddress: "a@b.com",
		User:        "user",
		Password:    "secret",
	}
	if err := sc.Validate(); err != nil {
		t.Errorf("user with password Validate() = %v, want nil", err)
	}
}
