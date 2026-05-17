package service

import (
	"strings"
	"testing"

	models "github.com/casapps/cassocial/src/server/model"
)

// ---- NewClient ----

func TestNewClient_NilConfig(t *testing.T) {
	_, err := NewClient(nil)
	if err == nil {
		t.Error("NewClient(nil) should return error")
	}
}

func TestNewClient_InvalidConfig_EmptyHost(t *testing.T) {
	cfg := &models.SMTPConfig{
		Host:        "",
		Port:        587,
		FromAddress: "from@example.com",
	}
	_, err := NewClient(cfg)
	if err == nil {
		t.Error("NewClient(empty host) should return error")
	}
}

func TestNewClient_ValidConfig(t *testing.T) {
	cfg := &models.SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		FromAddress: "from@example.com",
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient(valid) returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient(valid) returned nil client")
	}
}

// ---- GetProviderHost ----

func TestGetProviderHost_Custom(t *testing.T) {
	host, err := GetProviderHost(ProviderCustom)
	if err != nil {
		t.Errorf("GetProviderHost(Custom) returned error: %v", err)
	}
	if host != "" {
		t.Errorf("GetProviderHost(Custom) = %q, want empty string", host)
	}
}

func TestGetProviderHost_Gmail(t *testing.T) {
	host, err := GetProviderHost(ProviderGmail)
	if err != nil {
		t.Errorf("GetProviderHost(Gmail) returned error: %v", err)
	}
	if host != "smtp.gmail.com" {
		t.Errorf("GetProviderHost(Gmail) = %q, want smtp.gmail.com", host)
	}
}

func TestGetProviderHost_Yahoo(t *testing.T) {
	host, err := GetProviderHost(ProviderYahoo)
	if err != nil {
		t.Errorf("GetProviderHost(Yahoo) returned error: %v", err)
	}
	if host != "smtp.mail.yahoo.com" {
		t.Errorf("GetProviderHost(Yahoo) = %q, want smtp.mail.yahoo.com", host)
	}
}

func TestGetProviderHost_Outlook(t *testing.T) {
	host, err := GetProviderHost(ProviderOutlook)
	if err != nil {
		t.Errorf("GetProviderHost(Outlook) returned error: %v", err)
	}
	if host != "smtp-mail.outlook.com" {
		t.Errorf("GetProviderHost(Outlook) = %q, want smtp-mail.outlook.com", host)
	}
}

func TestGetProviderHost_Unknown(t *testing.T) {
	_, err := GetProviderHost("UnknownProvider")
	if err == nil {
		t.Error("GetProviderHost(unknown) should return error")
	}
}

// ---- buildMessage ----

func TestBuildMessage_PlainText(t *testing.T) {
	msg := buildMessage("Sender Name", "sender@example.com", []string{"recipient@example.com"}, "Test Subject", "Hello World", false)
	msgStr := string(msg)

	if !strings.Contains(msgStr, "From: Sender Name <sender@example.com>") {
		t.Errorf("buildMessage() missing From header, got:\n%s", msgStr)
	}
	if !strings.Contains(msgStr, "To: recipient@example.com") {
		t.Errorf("buildMessage() missing To header, got:\n%s", msgStr)
	}
	if !strings.Contains(msgStr, "Subject: Test Subject") {
		t.Errorf("buildMessage() missing Subject header, got:\n%s", msgStr)
	}
	if !strings.Contains(msgStr, "Content-Type: text/plain") {
		t.Errorf("buildMessage() missing text/plain Content-Type, got:\n%s", msgStr)
	}
	if !strings.Contains(msgStr, "Hello World") {
		t.Errorf("buildMessage() missing body, got:\n%s", msgStr)
	}
}

func TestBuildMessage_HTML(t *testing.T) {
	msg := buildMessage("", "sender@example.com", []string{"recipient@example.com"}, "HTML Test", "<h1>Hi</h1>", true)
	msgStr := string(msg)

	if !strings.Contains(msgStr, "Content-Type: text/html") {
		t.Errorf("buildMessage(isHTML=true) missing text/html Content-Type, got:\n%s", msgStr)
	}
	if !strings.Contains(msgStr, "From: sender@example.com") {
		t.Errorf("buildMessage(noFromName) missing bare From header, got:\n%s", msgStr)
	}
}

func TestBuildMessage_MultipleRecipients(t *testing.T) {
	msg := buildMessage("Sender", "s@example.com", []string{"a@example.com", "b@example.com", "c@example.com"}, "Multi", "body", false)
	msgStr := string(msg)

	// First recipient in To:
	if !strings.Contains(msgStr, "To: a@example.com") {
		t.Errorf("buildMessage() missing primary recipient, got:\n%s", msgStr)
	}
	// Additional recipients as Cc:
	if !strings.Contains(msgStr, "Cc: b@example.com") {
		t.Errorf("buildMessage() missing Cc for b@example.com, got:\n%s", msgStr)
	}
	if !strings.Contains(msgStr, "Cc: c@example.com") {
		t.Errorf("buildMessage() missing Cc for c@example.com, got:\n%s", msgStr)
	}
}

func TestBuildMessage_SingleRecipient_NoCc(t *testing.T) {
	msg := buildMessage("S", "s@example.com", []string{"a@example.com"}, "Single", "body", false)
	msgStr := string(msg)

	if strings.Contains(msgStr, "Cc:") {
		t.Errorf("buildMessage() with single recipient should not have Cc header, got:\n%s", msgStr)
	}
}

func TestBuildMessage_MIMEVersion(t *testing.T) {
	msg := buildMessage("S", "s@example.com", []string{"r@example.com"}, "Sub", "body", false)
	if !strings.Contains(string(msg), "MIME-Version: 1.0") {
		t.Error("buildMessage() missing MIME-Version header")
	}
}

// ---- PortConfigs and constants ----

func TestPortConfigs_Count(t *testing.T) {
	if len(PortConfigs) != 4 {
		t.Errorf("PortConfigs has %d entries, want 4", len(PortConfigs))
	}
}

func TestPortConfigs_Values(t *testing.T) {
	ports := map[int]SecurityType{
		25:   SecurityNone,
		587:  SecuritySTARTTLS,
		465:  SecuritySSLTLS,
		2525: SecuritySTARTTLS,
	}
	for _, pc := range PortConfigs {
		expected, ok := ports[pc.Port]
		if !ok {
			t.Errorf("unexpected port %d in PortConfigs", pc.Port)
			continue
		}
		if pc.Security != expected {
			t.Errorf("PortConfigs[%d].Security = %q, want %q", pc.Port, pc.Security, expected)
		}
		if pc.Label == "" {
			t.Errorf("PortConfigs[%d].Label is empty", pc.Port)
		}
	}
}
