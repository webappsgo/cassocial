package service

import (
	"log"
	"strconv"
	"sync"
	"time"

	models "github.com/casapps/cassocial/src/server/model"
)

// NotificationType represents different types of notifications
type NotificationType string

const (
	NotificationEmergency          NotificationType = "emergency"
	NotificationCertificate        NotificationType = "certificate"
	NotificationBugReport          NotificationType = "bug_report"
	NotificationUserRegistration   NotificationType = "user_registration"
	NotificationDomainVerification NotificationType = "domain_verification"
	NotificationBackupStatus       NotificationType = "backup_status"
	NotificationHighTraffic        NotificationType = "high_traffic"
)

// NotificationPriority defines retry counts for different priorities
type NotificationPriority int

const (
	PriorityNormal    NotificationPriority = 1  // 1 retry
	PriorityHigh      NotificationPriority = 3  // 3 retries
	PriorityEmergency NotificationPriority = 5  // 5 retries
)

// Notification represents a queued notification
type Notification struct {
	Type      NotificationType
	Recipient string
	Title     string
	Message   string
	Severity  string
	Priority  NotificationPriority
	CreatedAt time.Time
}

// NotificationManager handles notification queuing and batching
type NotificationManager struct {
	mailer      *Mailer
	preferences *models.NotificationPreferences
	queue       []*Notification
	mutex       sync.RWMutex
	batchDelay  time.Duration
	adminEmail  string
	ticker      *time.Ticker
	stopChan    chan bool
	running     bool
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(mailer *Mailer, preferences *models.NotificationPreferences, adminEmail string) *NotificationManager {
	if preferences == nil {
		preferences = &models.NotificationPreferences{
			Emergency:          true,
			Certificate:        true,
			BugReport:          true,
			UserRegistration:   true,
			DomainVerification: true,
			BackupStatus:       true,
			HighTraffic:        true,
			BatchDelay:         300, // 5 minutes default
		}
	}

	batchDelay := time.Duration(preferences.BatchDelay) * time.Second
	if batchDelay == 0 {
		batchDelay = 5 * time.Minute
	}

	return &NotificationManager{
		mailer:      mailer,
		preferences: preferences,
		queue:       make([]*Notification, 0),
		batchDelay:  batchDelay,
		adminEmail:  adminEmail,
		stopChan:    make(chan bool),
		running:     false,
	}
}

// Start starts the notification processing loop
func (nm *NotificationManager) Start() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if nm.running {
		return
	}

	nm.ticker = time.NewTicker(nm.batchDelay)
	nm.running = true

	go nm.processLoop()
}

// Stop stops the notification processing loop
func (nm *NotificationManager) Stop() {
	nm.mutex.Lock()

	if !nm.running {
		nm.mutex.Unlock()
		return
	}

	nm.stopChan <- true
	nm.ticker.Stop()
	nm.running = false
	nm.mutex.Unlock()

	// Process remaining notifications — called without holding the mutex so
	// processQueue() can acquire it without deadlocking.
	nm.processQueue()
}

// processLoop runs the background processing loop
func (nm *NotificationManager) processLoop() {
	for {
		select {
		case <-nm.ticker.C:
			nm.processQueue()
		case <-nm.stopChan:
			return
		}
	}
}

// Queue adds a notification to the queue
func (nm *NotificationManager) Queue(notification *Notification) {
	// Check if this notification type is enabled
	if !nm.isNotificationEnabled(notification.Type) {
		log.Printf("Notification type %s is disabled in preferences", notification.Type)
		return
	}

	// Emergency notifications bypass batching
	if notification.Priority == PriorityEmergency {
		nm.sendImmediately(notification)
		return
	}

	// Add to queue for batched sending
	nm.mutex.Lock()
	nm.queue = append(nm.queue, notification)
	nm.mutex.Unlock()
}

// processQueue processes all queued notifications
func (nm *NotificationManager) processQueue() {
	nm.mutex.Lock()
	notifications := nm.queue
	nm.queue = make([]*Notification, 0)
	nm.mutex.Unlock()

	if len(notifications) == 0 {
		return
	}

	log.Printf("Processing %d queued notifications", len(notifications))

	for _, notification := range notifications {
		nm.sendNotification(notification)
	}
}

// sendImmediately sends a notification immediately (bypassing batch)
func (nm *NotificationManager) sendImmediately(notification *Notification) {
	log.Printf("Sending emergency notification immediately: %s", notification.Title)
	nm.sendNotification(notification)
}

// sendNotification sends a single notification
func (nm *NotificationManager) sendNotification(notification *Notification) {
	if !nm.mailer.IsEnabled() {
		log.Printf("Mailer disabled: skipping notification: %s", notification.Title)
		return
	}

	recipient := notification.Recipient
	if recipient == "" {
		recipient = nm.adminEmail
	}

	retries := int(notification.Priority)

	err := nm.mailer.SendNotification(
		recipient,
		"Administrator",
		notification.Title,
		notification.Message,
		notification.Severity,
		retries,
	)

	if err != nil {
		log.Printf("Failed to send notification '%s' to %s: %v", notification.Title, recipient, err)
	} else {
		log.Printf("Sent notification '%s' to %s", notification.Title, recipient)
	}
}

// isNotificationEnabled checks if a notification type is enabled
func (nm *NotificationManager) isNotificationEnabled(notifType NotificationType) bool {
	switch notifType {
	case NotificationEmergency:
		return nm.preferences.Emergency
	case NotificationCertificate:
		return nm.preferences.Certificate
	case NotificationBugReport:
		return nm.preferences.BugReport
	case NotificationUserRegistration:
		return nm.preferences.UserRegistration
	case NotificationDomainVerification:
		return nm.preferences.DomainVerification
	case NotificationBackupStatus:
		return nm.preferences.BackupStatus
	case NotificationHighTraffic:
		return nm.preferences.HighTraffic
	default:
		return true
	}
}

// UpdatePreferences updates notification preferences
func (nm *NotificationManager) UpdatePreferences(preferences *models.NotificationPreferences) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	nm.preferences = preferences

	// Update batch delay if changed
	newBatchDelay := time.Duration(preferences.BatchDelay) * time.Second
	if newBatchDelay != nm.batchDelay && newBatchDelay > 0 {
		nm.batchDelay = newBatchDelay

		// Restart ticker with new delay
		if nm.running {
			nm.ticker.Stop()
			nm.ticker = time.NewTicker(nm.batchDelay)
		}
	}
}

// Helper methods for common notification scenarios

// NotifyEmergency sends an emergency notification
func (nm *NotificationManager) NotifyEmergency(title, message string) {
	nm.Queue(&Notification{
		Type:      NotificationEmergency,
		Recipient: nm.adminEmail,
		Title:     title,
		Message:   message,
		Severity:  "emergency",
		Priority:  PriorityEmergency,
		CreatedAt: time.Now(),
	})
}

// NotifyCertificateExpiring sends a certificate expiry notification
func (nm *NotificationManager) NotifyCertificateExpiring(domain string, daysUntilExpiry int) {
	var priority NotificationPriority
	if daysUntilExpiry <= 7 {
		priority = PriorityHigh
	} else {
		priority = PriorityNormal
	}

	nm.Queue(&Notification{
		Type:      NotificationCertificate,
		Recipient: nm.adminEmail,
		Title:     "SSL Certificate Expiring",
		Message:   formatCertificateMessage(domain, daysUntilExpiry),
		Severity:  "warning",
		Priority:  priority,
		CreatedAt: time.Now(),
	})
}

// NotifyBugReport sends a bug report notification
func (nm *NotificationManager) NotifyBugReport(reportedBy, description string) {
	nm.Queue(&Notification{
		Type:      NotificationBugReport,
		Recipient: nm.adminEmail,
		Title:     "New Bug Report",
		Message:   formatBugReportMessage(reportedBy, description),
		Severity:  "info",
		Priority:  PriorityNormal,
		CreatedAt: time.Now(),
	})
}

// NotifyUserRegistration sends a user registration notification
func (nm *NotificationManager) NotifyUserRegistration(username, email string) {
	nm.Queue(&Notification{
		Type:      NotificationUserRegistration,
		Recipient: nm.adminEmail,
		Title:     "New User Registration",
		Message:   formatUserRegistrationMessage(username, email),
		Severity:  "info",
		Priority:  PriorityNormal,
		CreatedAt: time.Now(),
	})
}

// NotifyDomainVerification sends a domain verification notification
func (nm *NotificationManager) NotifyDomainVerification(domain, status string) {
	severity := "info"
	if status != "verified" {
		severity = "warning"
	}

	nm.Queue(&Notification{
		Type:      NotificationDomainVerification,
		Recipient: nm.adminEmail,
		Title:     "Domain Verification Update",
		Message:   formatDomainVerificationMessage(domain, status),
		Severity:  severity,
		Priority:  PriorityNormal,
		CreatedAt: time.Now(),
	})
}

// NotifyBackupStatus sends a backup status notification
func (nm *NotificationManager) NotifyBackupStatus(status, details string) {
	severity := "info"
	priority := PriorityNormal

	if status != "success" {
		severity = "warning"
		priority = PriorityHigh
	}

	nm.Queue(&Notification{
		Type:      NotificationBackupStatus,
		Recipient: nm.adminEmail,
		Title:     "Backup Status",
		Message:   formatBackupStatusMessage(status, details),
		Severity:  severity,
		Priority:  priority,
		CreatedAt: time.Now(),
	})
}

// NotifyHighTraffic sends a high traffic notification
func (nm *NotificationManager) NotifyHighTraffic(currentLoad, threshold int) {
	nm.Queue(&Notification{
		Type:      NotificationHighTraffic,
		Recipient: nm.adminEmail,
		Title:     "High Traffic Alert",
		Message:   formatHighTrafficMessage(currentLoad, threshold),
		Severity:  "warning",
		Priority:  PriorityNormal,
		CreatedAt: time.Now(),
	})
}

// Message formatting helpers

func formatCertificateMessage(domain string, daysUntilExpiry int) string {
	return `<p>The SSL certificate for <strong>` + domain + `</strong> will expire in <strong>` +
		strconv.Itoa(daysUntilExpiry) + ` days</strong>.</p>
		<p>Automatic renewal will be attempted. If it fails, manual intervention may be required.</p>`
}

func formatBugReportMessage(reportedBy, description string) string {
	return `<p>A new bug report has been submitted:</p>
		<ul>
			<li><strong>Reported By:</strong> ` + reportedBy + `</li>
		</ul>
		<p><strong>Description:</strong></p>
		<p>` + description + `</p>`
}

func formatUserRegistrationMessage(username, email string) string {
	return `<p>A new user has registered:</p>
		<ul>
			<li><strong>Username:</strong> ` + username + `</li>
			<li><strong>Email:</strong> ` + email + `</li>
		</ul>
		<p>Please review and approve their account if required.</p>`
}

func formatDomainVerificationMessage(domain, status string) string {
	if status == "verified" {
		return `<p>The custom domain <strong>` + domain + `</strong> has been successfully verified.</p>`
	}
	return `<p>Domain verification failed for <strong>` + domain + `</strong>. Please check DNS settings.</p>`
}

func formatBackupStatusMessage(status, details string) string {
	if status == "success" {
		return `<p>Database backup completed successfully.</p><p>` + details + `</p>`
	}
	return `<p>Database backup failed.</p><p><strong>Error:</strong> ` + details + `</p>`
}

func formatHighTrafficMessage(currentLoad, threshold int) string {
	return `<p>Your instance is experiencing high traffic.</p>
		<ul>
			<li><strong>Current Load:</strong> ` + strconv.Itoa(currentLoad) + ` requests/minute</li>
			<li><strong>Threshold:</strong> ` + strconv.Itoa(threshold) + ` requests/minute</li>
		</ul>
		<p>System performance may be affected. Consider scaling resources if this persists.</p>`
}

// GetQueueLength returns the current queue length
func (nm *NotificationManager) GetQueueLength() int {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	return len(nm.queue)
}

// IsRunning returns whether the notification manager is running
func (nm *NotificationManager) IsRunning() bool {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	return nm.running
}
