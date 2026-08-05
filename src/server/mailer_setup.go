package server

import (
	"log"
	"strconv"

	"github.com/casapps/cassocial/src/server/model"
	"github.com/casapps/cassocial/src/server/store"
	"github.com/casapps/cassocial/src/service"
)

// LoadSMTPConfig reads SMTP settings from the DB settings table — the same
// keys the admin panel's SMTP config endpoints read/write (see
// AdminHandlers.GetSMTPConfig / UpdateSMTPConfig / TestSMTPConnection).
func LoadSMTPConfig(db *store.DB) *model.SMTPConfig {
	host, _ := db.GetSetting("smtp_host")
	portStr, _ := db.GetSetting("smtp_port")
	security, _ := db.GetSetting("smtp_security")
	user, _ := db.GetSetting("smtp_user")
	password, _ := db.GetSetting("smtp_password")
	fromName, _ := db.GetSetting("smtp_from_name")
	fromAddress, _ := db.GetSetting("smtp_from_address")
	adminEmail, _ := db.GetSetting("admin_email")
	provider, _ := db.GetSetting("smtp_provider")

	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 {
		port = 587
	}

	return &model.SMTPConfig{
		Provider:    provider,
		Host:        host,
		Port:        port,
		Security:    security,
		User:        user,
		Password:    password,
		FromName:    fromName,
		FromAddress: fromAddress,
		AdminEmail:  adminEmail,
	}
}

// NewMailerFromSettings builds a Mailer from the DB-backed SMTP settings.
//
// Per AI.md PART 18: "ALL emails require a valid and working SMTP server.
// No SMTP = No emails." There is no manual enable/disable toggle — email
// sending is enabled only when a live SMTP connection test succeeds at
// construction time (checked once at startup). Missing config or a failed
// connection test disables email silently (logged as a warning, never as
// an error) and never fails startup; it is retried the next time the
// process starts.
func NewMailerFromSettings(db *store.DB, siteName, siteURL string) *service.Mailer {
	cfg := LoadSMTPConfig(db)

	switch {
	case cfg.Host == "" || cfg.FromAddress == "":
		log.Println("SMTP not configured — account emails disabled")
		cfg.Enabled = false
	default:
		if err := testSMTPConnection(cfg); err != nil {
			log.Printf("SMTP connection test failed, account emails disabled: %v", err)
			cfg.Enabled = false
		} else {
			cfg.Enabled = true
		}
	}

	mailer, err := service.NewMailer(cfg, siteName, siteURL)
	if err != nil {
		log.Printf("Failed to create mailer: %v", err)
	}
	return mailer
}

// testSMTPConnection validates the SMTP config and performs a live
// connection test, matching AdminHandlers.TestSMTPConnection's behavior.
func testSMTPConnection(cfg *model.SMTPConfig) error {
	testCfg := *cfg
	testCfg.Enabled = true

	client, err := service.NewClient(&testCfg)
	if err != nil {
		return err
	}

	return client.TestConnection()
}
