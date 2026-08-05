package service

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server/store"
)

// manifestSchemaVersion is the manifest.json schema version (distinct from
// the running application's version, which is stored in Manifest.AppVersion).
const manifestSchemaVersion = "1.0.0"

// manifestFilename is the well-known archive entry holding backup metadata
// (AI.md PART 22 "Backup Format").
const manifestFilename = "manifest.json"

// Argon2id parameters for backup-encryption key derivation (OWASP 2023
// params, matching src/server/password_hash.go's password-hashing profile).
const (
	backupArgonTime    = 3
	backupArgonMemory  = 64 * 1024
	backupArgonThreads = 4
	backupArgonKeyLen  = 32
	backupArgonSaltLen = 16
)

// BackupService handles backup and restore operations
type BackupService struct {
	config *config.Config
	db     *store.DB
	// AppVersion is the running application's build-info version, embedded
	// into every backup's manifest.json and compared on restore (AI.md
	// PART 22 "Restore Verification" — version mismatch is a warning only).
	// Left unset by NewBackupService; callers (e.g. src/cli_ops.go) assign
	// it after construction so this constructor's signature stays stable.
	AppVersion string
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

// Manifest describes a backup archive's contents and integrity metadata,
// stored as the archive's final entry (manifest.json). Field names and
// shape follow AI.md PART 22 "Backup Format" exactly.
type Manifest struct {
	Version          string   `json:"version"`
	CreatedAt        string   `json:"created_at"`
	CreatedBy        string   `json:"created_by"`
	AppVersion       string   `json:"app_version"`
	Contents         []string `json:"contents"`
	Encrypted        bool     `json:"encrypted"`
	EncryptionMethod string   `json:"encryption_method,omitempty"`
	Checksum         string   `json:"checksum"`
}

// CreateBackup creates a full, unencrypted backup.
func (s *BackupService) CreateBackup(backupType string) (*Backup, error) {
	return s.createBackup(backupType, "")
}

// CreateBackupEncrypted creates a full backup encrypted with the given
// password (AES-256-GCM, Argon2id key derivation — AI.md PART 22 "Backup
// Encryption"). The unencrypted archive is built entirely in memory and
// never touches disk.
func (s *BackupService) CreateBackupEncrypted(backupType, password string) (*Backup, error) {
	if password == "" {
		return nil, fmt.Errorf("encryption password is required")
	}
	return s.createBackup(backupType, password)
}

// createBackup builds and writes a backup archive. When password is
// non-empty the archive is AES-256-GCM encrypted before the single disk
// write; when empty the archive is written as plain tar.gz.
func (s *BackupService) createBackup(backupType, password string) (*Backup, error) {
	if err := s.checkComplianceMode(password); err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("20060102-150405")
	ext := "tar.gz"
	if password != "" {
		ext = "tar.gz.enc"
	}
	filename := fmt.Sprintf("cassocial-backup-%s-%s.%s", backupType, timestamp, ext)

	backupDir := filepath.Join(s.config.DataDir, "backup")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	backupPath := filepath.Join(backupDir, filename)

	log.Printf("Creating backup: %s", backupPath)

	// Build the tar archive in memory so its checksum can be computed and
	// a manifest.json entry appended before compression/encryption.
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)

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

	if err := tw.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush backup archive: %w", err)
	}

	// Checksum + content listing are computed over the content entries
	// written so far, before the manifest entry itself is appended.
	checksum, contents, _, err := hashTarContents(tarBuf.Bytes(), "")
	if err != nil {
		return nil, fmt.Errorf("failed to checksum backup contents: %w", err)
	}

	manifest := Manifest{
		Version:    manifestSchemaVersion,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		CreatedBy:  "system",
		AppVersion: s.AppVersion,
		Contents:   contents,
		Encrypted:  password != "",
		Checksum:   "sha256:" + checksum,
	}
	if manifest.Encrypted {
		manifest.EncryptionMethod = "AES-256-GCM"
	}

	manifestBytes, err := json.MarshalIndent(&manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal backup manifest: %w", err)
	}
	if err := addBytesToTar(tw, manifestBytes, manifestFilename); err != nil {
		return nil, fmt.Errorf("failed to add manifest to backup: %w", err)
	}
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close backup archive: %w", err)
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to compress backup: %w", err)
	}
	if err := gzw.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize backup compression: %w", err)
	}

	output := gzBuf.Bytes()
	if password != "" {
		encrypted, err := encryptBackup(output, password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt backup: %w", err)
		}
		output = encrypted
	}

	// Single disk write — the unencrypted archive never touches disk
	// (AI.md PART 22 "Backup Encryption").
	if err := os.WriteFile(backupPath, output, 0600); err != nil {
		return nil, fmt.Errorf("failed to write backup file: %w", err)
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

// checkComplianceMode blocks unencrypted backups when compliance mode is
// enabled and no password was supplied (AI.md PART 22 "Compliance Mode
// Enforcement"). A missing/unparseable setting is treated as compliance
// mode disabled — GetSetting/ParseBool errors never block a normal backup.
func (s *BackupService) checkComplianceMode(password string) error {
	raw, err := s.db.GetSetting("backup_compliance_enabled")
	if err != nil {
		return nil
	}
	enabled, err := config.ParseBool(raw, false)
	if err != nil {
		return nil
	}
	if enabled && password == "" {
		return fmt.Errorf("compliance mode is enabled: an encryption password is required for backups")
	}
	return nil
}

// RestoreBackup restores from an unencrypted backup archive.
func (s *BackupService) RestoreBackup(backupFilename string) error {
	return s.restoreBackup(backupFilename, "")
}

// RestoreBackupWithPassword restores from a backup archive, decrypting it
// first if it is AES-256-GCM encrypted (`.tar.gz.enc`).
func (s *BackupService) RestoreBackupWithPassword(backupFilename, password string) error {
	return s.restoreBackup(backupFilename, password)
}

// restoreBackup runs the full AI.md PART 22 "Restore Verification" pipeline
// — file exists, file readable, format valid, decrypt test, checksum valid,
// manifest valid, version compatible — and only extracts once every check
// passes.
func (s *BackupService) restoreBackup(backupFilename, password string) error {
	backupPath := filepath.Join(s.config.DataDir, "backup", backupFilename)

	log.Printf("Restoring from backup: %s", backupPath)

	// 1. File exists
	info, err := os.Stat(backupPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("backup file does not exist: %s", backupFilename)
		}
		return fmt.Errorf("failed to stat backup file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("backup path is a directory, not a file: %s", backupFilename)
	}

	// 2. File readable
	raw, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("backup file is not readable: %w", err)
	}

	// 3. Decrypt test (only applicable to .tar.gz.enc archives)
	gzData := raw
	if strings.HasSuffix(backupFilename, ".enc") {
		if password == "" {
			return fmt.Errorf("backup is encrypted: a password is required to restore it")
		}
		decrypted, err := decryptBackup(raw, password)
		if err != nil {
			return fmt.Errorf("failed to decrypt backup (wrong password?): %w", err)
		}
		gzData = decrypted
	}

	// 4. Format valid (gzip layer)
	gzr, err := gzip.NewReader(bytes.NewReader(gzData))
	if err != nil {
		return fmt.Errorf("invalid backup format: %w", err)
	}
	tarData, err := io.ReadAll(gzr)
	gzr.Close()
	if err != nil {
		return fmt.Errorf("invalid backup format: %w", err)
	}

	// 4b. Format valid (tar layer) + 6. manifest valid + 5. checksum valid
	checksum, _, manifest, err := hashTarContents(tarData, manifestFilename)
	if err != nil {
		return fmt.Errorf("invalid backup format: %w", err)
	}
	if manifest == nil {
		return fmt.Errorf("backup is missing manifest.json — cannot verify integrity")
	}
	wantChecksum := strings.TrimPrefix(manifest.Checksum, "sha256:")
	if !strings.EqualFold(checksum, wantChecksum) {
		return fmt.Errorf("backup checksum mismatch — archive may be corrupted or tampered with")
	}

	// 7. Version compatible — a mismatch is a warning only, never a hard
	// failure (AI.md PART 22 "Restore Verification").
	if manifest.AppVersion != "" && s.AppVersion != "" && manifest.AppVersion != s.AppVersion {
		log.Printf("Warning: backup app_version %q does not match running version %q", manifest.AppVersion, s.AppVersion)
	}

	// All verification checks passed — extract.
	tr := tar.NewReader(bytes.NewReader(tarData))
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Name == manifestFilename {
			continue
		}

		// Reject symlinks/hardlinks outright - they can point outside DataDir
		// regardless of what the target path check below allows.
		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
			return fmt.Errorf("refusing to restore backup: %q is a link entry", header.Name)
		}

		// Determine target path and verify it stays within DataDir - reject
		// any entry whose cleaned path escapes via ".." or an absolute path
		// (Zip-Slip / path traversal).
		target := filepath.Join(s.config.DataDir, header.Name)
		rel, err := filepath.Rel(s.config.DataDir, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("refusing to restore backup: %q escapes data directory", header.Name)
		}

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
		name := entry.Name()
		if entry.IsDir() || !(strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tar.gz.enc")) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backupType := "manual"

		backups = append(backups, &Backup{
			Filename:  name,
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

// rotateBackups keeps only the most recent N backups.
// The count is derived from backup_retention_days in settings, defaulting to 30.
func (s *BackupService) rotateBackups(backupDir string) error {
	retentionDays := 30
	if retStr, err := s.db.GetSetting("backup_retention_days"); err == nil {
		var n int
		if _, err := fmt.Sscanf(retStr, "%d", &n); err == nil && n > 0 {
			retentionDays = n
		}
	}
	// Keep at least 1 backup per day over the retention period (daily schedule assumed)
	maxBackups := retentionDays

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

// addBytesToTar writes an in-memory byte slice as a tar entry (used for
// manifest.json, which has no on-disk source file).
func addBytesToTar(tw *tar.Writer, data []byte, archivePath string) error {
	header := &tar.Header{
		Name:    archivePath,
		Size:    int64(len(data)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

// hashTarContents reads every entry of a tar archive and returns a SHA-256
// checksum computed over the body bytes of all entries except skip (in
// archive order), the names of those entries, and — if an entry named skip
// is present — its JSON-decoded Manifest. It is used both to compute a
// backup's checksum/content-list at creation time (skip == "", no manifest
// entry exists yet) and to verify a backup's checksum/manifest at restore
// time (skip == manifestFilename) — AI.md PART 22 "Restore Verification".
func hashTarContents(tarData []byte, skip string) (checksum string, names []string, manifest *Manifest, err error) {
	h := sha256.New()
	tr := tar.NewReader(bytes.NewReader(tarData))
	for {
		header, terr := tr.Next()
		if terr == io.EOF {
			break
		}
		if terr != nil {
			return "", nil, nil, terr
		}

		if skip != "" && header.Name == skip {
			data, rerr := io.ReadAll(tr)
			if rerr != nil {
				return "", nil, nil, rerr
			}
			var m Manifest
			if jerr := json.Unmarshal(data, &m); jerr == nil {
				manifest = &m
			}
			continue
		}

		names = append(names, header.Name)
		if _, cerr := io.Copy(h, tr); cerr != nil {
			return "", nil, nil, cerr
		}
	}
	return hex.EncodeToString(h.Sum(nil)), names, manifest, nil
}

// deriveBackupKey derives a 256-bit AES key from a password + salt using
// Argon2id (OWASP 2023 params, matching the password-hashing profile in
// src/server/password_hash.go).
func deriveBackupKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, backupArgonTime, backupArgonMemory, backupArgonThreads, backupArgonKeyLen)
}

// encryptBackup encrypts plaintext with AES-256-GCM using an Argon2id-derived
// key, returning salt||nonce||ciphertext. The password itself is never
// stored anywhere (AI.md PART 22 "Backup Encryption").
func encryptBackup(plaintext []byte, password string) ([]byte, error) {
	salt := make([]byte, backupArgonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	key := deriveBackupKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	out := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	out = append(out, salt...)
	out = append(out, nonce...)
	out = append(out, ciphertext...)
	return out, nil
}

// decryptBackup reverses encryptBackup: it reads the salt+nonce prefix,
// re-derives the key with Argon2id, and AES-256-GCM decrypts the remainder.
// A wrong password surfaces as a GCM authentication failure.
func decryptBackup(data []byte, password string) ([]byte, error) {
	if len(data) < backupArgonSaltLen {
		return nil, fmt.Errorf("encrypted backup is too short")
	}
	salt := data[:backupArgonSaltLen]
	key := deriveBackupKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < backupArgonSaltLen+nonceSize {
		return nil, fmt.Errorf("encrypted backup is too short")
	}
	nonce := data[backupArgonSaltLen : backupArgonSaltLen+nonceSize]
	ciphertext := data[backupArgonSaltLen+nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong password?): %w", err)
	}
	return plaintext, nil
}
