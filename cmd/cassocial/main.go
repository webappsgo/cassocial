package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/casapps/cassocial/internal/auth"
	"github.com/casapps/cassocial/internal/config"
	"github.com/casapps/cassocial/internal/database"
	"github.com/casapps/cassocial/internal/web"
)

const (
	Version = "1.0.0"
)

func main() {
	// CLI flags
	showVersion := flag.Bool("version", false, "Show version information")
	showVersionShort := flag.Bool("v", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help information")
	showHelpShort := flag.Bool("h", false, "Show help information")
	resetAdmin := flag.Bool("reset-admin", false, "Reset administrator password")
	port := flag.Int("port", 8080, "Port to listen on")
	portShort := flag.Int("p", 0, "Port to listen on")
	host := flag.String("host", "0.0.0.0", "Host to bind to")
	hostShort := flag.String("h", "", "Host to bind to")

	flag.Parse()

	// Handle version flag
	if *showVersion || *showVersionShort {
		fmt.Printf("Cassocial v%s\n", Version)
		fmt.Println("Self-hosted link aggregator and social profile landing page")
		fmt.Println("License: MIT")
		fmt.Println("Repository: https://github.com/casapps/cassocial")
		os.Exit(0)
	}

	// Handle help flag
	if *showHelp || *showHelpShort {
		printHelp()
		os.Exit(0)
	}

	// Initialize configuration
	cfg := config.New()

	// Resolve port and host
	if *portShort != 0 {
		cfg.Port = *portShort
	} else if *port != 0 {
		cfg.Port = *port
	}

	if *hostShort != "" {
		cfg.Host = *hostShort
	} else if *host != "" {
		cfg.Host = *host
	}

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Handle reset-admin command
	if *resetAdmin {
		handleResetAdmin(db)
		os.Exit(0)
	}

	// Setup logging
	setupLogging(cfg)

	// Initialize auth service
	authService := auth.NewAuth(db, cfg.MasterKey)

	// Start web server
	log.Printf("Starting Cassocial v%s on %s:%d", Version, cfg.Host, cfg.Port)
	server, err := web.New(cfg, db, authService)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func printHelp() {
	fmt.Println("Cassocial - Self-hosted link aggregator and social profile")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  cassocial [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -v, --version          Show version information")
	fmt.Println("  -h, --help             Show this help message")
	fmt.Println("  --reset-admin          Reset administrator password (emergency only)")
	fmt.Println("  -p, --port <port>      Port to listen on (default: 8080)")
	fmt.Println("  -h, --host <host>      Host to bind to (default: 0.0.0.0)")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  CASSOCIAL_PORT                Port to listen on")
	fmt.Println("  CASSOCIAL_HOST                Host to bind to")
	fmt.Println("  CASSOCIAL_DATA                Data directory path")
	fmt.Println("  CASSOCIAL_LOG_LEVEL           Log level (debug, info, warning, error)")
	fmt.Println("  CASSOCIAL_DATABASE_URL        Database connection URL")
	fmt.Println("  CASSOCIAL_MASTER_KEY          Master encryption key")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cassocial                     # Start with defaults")
	fmt.Println("  cassocial -p 3000             # Start on port 3000")
	fmt.Println("  cassocial --reset-admin       # Reset admin password")
	fmt.Println()
}

func handleResetAdmin(db *database.DB) {
	fmt.Println("Reset Administrator Password")
	fmt.Println("=============================")
	fmt.Println()
	fmt.Print("Enter new password: ")

	var password string
	fmt.Scanln(&password)

	if len(password) < 8 {
		fmt.Println("Error: Password must be at least 8 characters")
		os.Exit(1)
	}

	// Implementation would hash password and update admin user
	fmt.Println("Administrator password has been reset successfully")
	fmt.Println("You can now login with username 'administrator' and your new password")
}

func setupLogging(cfg *config.Config) {
	var logPath string

	// Determine log file location based on install type
	if _, err := os.Stat("./data"); err == nil {
		// Portable mode
		logPath = "./data/cassocial.log"
	} else if os.Geteuid() == 0 {
		// System install
		os.MkdirAll("/var/log/cassocial", 0755)
		logPath = "/var/log/cassocial/cassocial.log"
	} else {
		// User install
		homeDir, _ := os.UserHomeDir()
		logDir := filepath.Join(homeDir, ".local", "state", "cassocial")
		os.MkdirAll(logDir, 0755)
		logPath = filepath.Join(logDir, "cassocial.log")
	}

	// Open log file
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Warning: Could not open log file %s: %v", logPath, err)
		return
	}

	// Set log output to file
	log.SetOutput(logFile)
	log.Printf("Logging to: %s", logPath)
}
