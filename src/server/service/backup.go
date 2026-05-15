package service

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// BackupService handles backup and restore operations
type BackupService struct {
	config *config.Config
	db     *store.DB
}

// NewBackupService creates a new backup service
func NewBackupService(cfg *config.Config, db *store.DB) *BackupService {
	return &BackupService{
		config: cfg,
		db:     db,
	}
}

// Backup represents a backup archive
type Backup struct {
	ID        string
	Filename  string
	Size      int64
	CreatedAt time.Time
	Type      string // auto, manual
}

// CreateBackup creates a full backup
func (s *BackupService) CreateBackup(backupType string) (*Backup, error) {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("cassocial-backup-%s-%s.tar.gz", backupType, timestamp)

	backupDir := filepath.Join(s.config.DataDir, "backup")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	backupPath := filepath.Join(backupDir, filename)

	log.Printf("Creating backup: %s", backupPath)

	// Create tar.gz file
	file, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Backup database files
	if err := s.addDatabaseToBackup(tw); err != nil {
		return nil, fmt.Errorf("failed to backup database: %w", err)
	}

	// Backup configuration
	if err := s.addConfigToBackup(tw); err != nil {
		return nil, fmt.Errorf("failed to backup config: %w", err)
	}

	// Backup uploads (avatars, images)
	if err := s.addUploadsToBackup(tw); err != nil {
		log.Printf("Warning: Failed to backup uploads: %v", err)
		// Continue anyway
	}

	// Get file info
	info, err := os.Stat(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat backup file: %w", err)
	}

	backup := &Backup{
		Filename:  filename,
		Size:      info.Size(),
		CreatedAt: time.Now(),
		Type:      backupType,
	}

	log.Printf("Backup created: %s (size: %d bytes)", filename, info.Size())

	// Rotate old backups
	if err := s.rotateBackups(backupDir); err != nil {
		log.Printf("Warning: Failed to rotate backups: %v", err)
	}

	return backup, nil
}

// RestoreBackup restores from a backup archive
func (s *BackupService) RestoreBackup(backupFilename string) error {
	backupPath := filepath.Join(s.config.DataDir, "backup", backupFilename)

	log.Printf("Restoring from backup: %s", backupPath)

	// Open backup file
	file, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Extract files
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Determine target path
		target := filepath.Join(s.config.DataDir, header.Name)

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract file
		if header.Typeflag == tar.TypeReg {
			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
		}
	}

	log.Println("Backup restored successfully")
	return nil
}

// ListBackups lists available backups
func (s *BackupService) ListBackups() ([]*Backup, error) {
	backupDir := filepath.Join(s.config.DataDir, "backup")

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Backup{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	backups := make([]*Backup, 0)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".gz" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backupType := "manual"

		backups = append(backups, &Backup{
			Filename:  entry.Name(),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
			Type:      backupType,
		})
	}

	return backups, nil
}

// addDatabaseToBackup adds database files to backup
func (s *BackupService) addDatabaseToBackup(tw *tar.Writer) error {
	dbDir := filepath.Join(s.config.DataDir, "db")

	return filepath.Walk(dbDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(s.config.DataDir, path)
		if err != nil {
			return err
		}

		// Add to tar
		return addFileToTar(tw, path, relPath)
	})
}

// addConfigToBackup adds configuration to backup
func (s *BackupService) addConfigToBackup(tw *tar.Writer) error {
	configFile := filepath.Join(s.config.ConfigDir, "server.yml")

	if _, err := os.Stat(configFile); err != nil {
		return nil // Config file doesn't exist yet
	}

	return addFileToTar(tw, configFile, "config/server.yml")
}

// addUploadsToBackup adds uploaded files to backup
func (s *BackupService) addUploadsToBackup(tw *tar.Writer) error {
	uploadsDir := filepath.Join(s.config.DataDir, "uploads")

	if _, err := os.Stat(uploadsDir); err != nil {
		return nil // Uploads directory doesn't exist
	}

	return filepath.Walk(uploadsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		relPath, err := filepath.Rel(s.config.DataDir, path)
		if err != nil {
			return err
		}

		return addFileToTar(tw, path, relPath)
	})
}

// rotateBackups keeps only the most recent N backups
func (s *BackupService) rotateBackups(backupDir string) error {
	maxBackups := 4 // Keep max 4 backups

	backups, err := s.ListBackups()
	if err != nil {
		return err
	}

	// Sort by creation time (newest first) - already sorted by ModTime
	if len(backups) <= maxBackups {
		return nil
	}

	// Delete oldest backups
	for i := maxBackups; i < len(backups); i++ {
		backupPath := filepath.Join(backupDir, backups[i].Filename)
		log.Printf("Rotating old backup: %s", backups[i].Filename)
		if err := os.Remove(backupPath); err != nil {
			log.Printf("Warning: Failed to delete old backup: %v", err)
		}
	}

	return nil
}

// addFileToTar adds a file to a tar archive
func addFileToTar(tw *tar.Writer, filePath, archivePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    archivePath,
		Size:    info.Size(),
		Mode:    int64(info.Mode()),
		ModTime: info.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tw, file); err != nil {
		return err
	}

	return nil
}
