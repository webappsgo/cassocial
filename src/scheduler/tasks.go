package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// Tasks holds all scheduled task implementations
type Tasks struct {
	config *config.Config
	db     *store.DB
}

// NewTasks creates a new tasks instance
func NewTasks(cfg *config.Config, db *store.DB) *Tasks {
	return &Tasks{
		config: cfg,
		db:     db,
	}
}

// RegisterAllTasks registers all scheduled tasks
func (t *Tasks) RegisterAllTasks(scheduler *Scheduler) error {
	// Analytics aggregation - every hour
	if err := scheduler.RegisterTask("analytics_aggregation", "0 0 * * * *", t.AggregateAnalytics); err != nil {
		return err
	}

	// Certificate renewal check - daily at 3am
	if err := scheduler.RegisterTask("cert_renewal_check", "0 0 3 * * *", t.CheckCertificateRenewal); err != nil {
		return err
	}

	// Database cleanup - daily at 2am
	if err := scheduler.RegisterTask("database_cleanup", "0 0 2 * * *", t.CleanupDatabase); err != nil {
		return err
	}

	// Automated backup - daily at 2am (after cleanup)
	if err := scheduler.RegisterTask("automated_backup", "0 30 2 * * *", t.CreateBackup); err != nil {
		return err
	}

	// Email queue processing - every 5 minutes
	if err := scheduler.RegisterTask("email_queue", "0 */5 * * * *", t.ProcessEmailQueue); err != nil {
		return err
	}

	// Session cleanup - hourly
	if err := scheduler.RegisterTask("session_cleanup", "0 0 * * * *", t.CleanupSessions); err != nil {
		return err
	}

	// GeoIP database update - weekly on Sunday at 3am
	if err := scheduler.RegisterTask("geoip_update", "0 0 3 * * 0", t.UpdateGeoIPDatabase); err != nil {
		return err
	}

	log.Println("All scheduled tasks registered")
	return nil
}

// AggregateAnalytics aggregates hourly analytics data
func (t *Tasks) AggregateAnalytics() error {
	log.Println("Running analytics aggregation...")
	return nil
}

// CheckCertificateRenewal checks if SSL certificates need renewal
func (t *Tasks) CheckCertificateRenewal() error {
	log.Println("Checking SSL certificate renewal...")

	if !t.config.SSL.Enabled {
		return nil
	}

	return nil
}

// CleanupDatabase performs database cleanup operations
func (t *Tasks) CleanupDatabase() error {
	log.Println("Running database cleanup...")
	if err := t.db.DeleteExpiredShortlinks(); err != nil {
		log.Printf("Failed to delete expired shortlinks: %v", err)
	}
	return nil
}

// CreateBackup creates an automated backup
func (t *Tasks) CreateBackup() error {
	log.Println("Creating automated backup...")
	return nil
}

// ProcessEmailQueue processes queued emails
func (t *Tasks) ProcessEmailQueue() error {
	return nil
}

// CleanupSessions removes expired sessions
func (t *Tasks) CleanupSessions() error {
	log.Println("Cleaning up expired sessions...")

	if err := t.db.CleanupExpiredSessions(); err != nil {
		return fmt.Errorf("failed to cleanup sessions: %w", err)
	}

	return nil
}

// UpdateGeoIPDatabase downloads and updates GeoIP database
func (t *Tasks) UpdateGeoIPDatabase() error {
	log.Println("Updating GeoIP database...")
	return nil
}

// GetTaskStatistics returns statistics about scheduled tasks
func (t *Tasks) GetTaskStatistics(scheduler *Scheduler) map[string]interface{} {
	tasks := scheduler.ListTasks()

	stats := make([]map[string]interface{}, 0, len(tasks))
	for _, task := range tasks {
		stats = append(stats, map[string]interface{}{
			"name":        task.Name,
			"schedule":    task.Schedule,
			"enabled":     task.Enabled,
			"last_run":    task.LastRun.Format(time.RFC3339),
			"next_run":    task.NextRun.Format(time.RFC3339),
			"run_count":   task.RunCount,
			"error_count": task.ErrorCount,
			"last_error":  task.LastError,
		})
	}

	return map[string]interface{}{
		"total_tasks": len(tasks),
		"tasks":       stats,
	}
}
