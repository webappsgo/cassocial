package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/casapps/cassocial/src/config"
	"github.com/casapps/cassocial/src/server"
	"github.com/casapps/cassocial/src/server/store"
)

// Build info - set via -ldflags at build time
var (
	Version   = "dev"
	CommitID  = "unknown"
	BuildDate = "unknown"
)

func main() {
	// CLI flags (following TEMPLATE.md NON-NEGOTIABLE specification)
	var (
		showHelp    = flag.Bool("help", false, "Show help information")
		showHelpS   = flag.Bool("h", false, "Show help information")
		showVersion = flag.Bool("version", false, "Show version information")
		showVersionS = flag.Bool("v", false, "Show version information")

		// Directory flags
		configDir = flag.String("config", "", "Configuration directory")
		dataDir   = flag.String("data", "", "Data directory")
		logDir    = flag.String("log", "", "Log directory")
		pidFile   = flag.String("pid", "", "PID file path")

		// Server flags
		address = flag.String("address", "", "Listen address")
		port    = flag.Int("port", 0, "Listen port")
		mode    = flag.String("mode", "", "Application mode (production|development)")

		// Operation flags
		debug      = flag.Bool("debug", false, "Enable debug mode")
		showStatus = flag.Bool("status", false, "Show status and health")
		_          = flag.Bool("daemon", false, "Run as daemon") // TODO: Implement daemonization

		// Service management
		_ = flag.String("service", "", "Service command (start|stop|restart|reload|--install|--uninstall)") // TODO: Implement service management

		// Maintenance
		_ = flag.String("maintenance", "", "Maintenance command (backup|restore|update|mode|setup)") // TODO: Implement maintenance
		_ = flag.String("update", "", "Update command (check|yes|branch)") // TODO: Implement update
	)

	flag.Parse()

	// Handle version
	if *showVersion || *showVersionS {
		printVersion()
		os.Exit(0)
	}

	// Handle help
	if *showHelp || *showHelpS {
		printHelp()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configDir, *dataDir, *logDir)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override config with CLI flags
	if *address != "" {
		cfg.Server.Address = *address
	}
	if *port != 0 {
		cfg.Server.Port = *port
	}
	if *mode != "" {
		cfg.Server.Mode = *mode
	}
	if *debug {
		cfg.Server.Debug = true
	}

	// Determine PID file path
	pidFilePath := config.DeterminePIDFile(*pidFile)

	// Handle status
	if *showStatus {
		handleStatus(cfg, pidFilePath)
		os.Exit(0)
	}

	// TODO: Implement service, maintenance, and update handlers
	// For now, these are defined but not used

	// Write PID file
	if err := server.WritePIDFile(pidFilePath); err != nil {
		log.Fatalf("Failed to write PID file: %v", err)
	}
	// Ensure PID file is removed on exit
	defer server.RemovePIDFile(pidFilePath)

	// Initialize database
	log.Printf("Connecting to database (driver: %s)...", cfg.Database.Driver)
	db, err := store.Connect(cfg.Database.Driver, cfg.Database.Name)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	log.Println("Running database migrations...")
	if err := db.RunMigrations(); err != nil {
		log.Printf("Warning: Migration error: %v", err)
	}

	// Create and start server
	log.Printf("Cassocial v%s starting...", Version)
	log.Printf("Mode: %s, Address: %s, Port: %d", cfg.Server.Mode, cfg.Server.Address, cfg.Server.Port)

	srv, err := server.New(cfg, db)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server (blocks until shutdown)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func printVersion() {
	// Use actual binary name (AI.md PART 7 - binary naming rules)
	binaryName := filepath.Base(os.Args[0])

	// Format exactly as specified in AI.md PART 16
	fmt.Printf("%s v%s\n", binaryName, Version)
	fmt.Printf("Built: %s\n", BuildDate) // Already in ISO 8601 from Makefile
	fmt.Printf("Go: %s\n", runtime.Version()[2:]) // Remove "go" prefix
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func printHelp() {
	// Use actual binary name (AI.md PART 7 - binary naming rules)
	binaryName := filepath.Base(os.Args[0])

	fmt.Printf("%s - Self-hosted link aggregator and social profile\n", binaryName)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s [options]\n", binaryName)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help                       Show this help message")
	fmt.Println("  -v, --version                    Show version information")
	fmt.Println("  --mode {production|development}  Set application mode")
	fmt.Println("  --config {dir}                   Configuration directory")
	fmt.Println("  --data {dir}                     Data directory")
	fmt.Println("  --log {dir}                      Log directory")
	fmt.Println("  --pid {file}                     PID file path")
	fmt.Println("  --address {addr}                 Listen address")
	fmt.Println("  --port {port}                    Listen port")
	fmt.Println("  --debug                        Enable debug mode")
	fmt.Println("  --status                       Show status and health")
	fmt.Println("  --daemon                       Run as daemon")
	fmt.Println("  --service {cmd}                Service management")
	fmt.Println("  --maintenance {cmd}            Maintenance operations")
	fmt.Println("  --update {cmd}                 Update operations")
	fmt.Println()
	fmt.Println("Service Commands:")
	fmt.Println("  start, stop, restart, reload, --install, --uninstall")
	fmt.Println()
	fmt.Println("Maintenance Commands:")
	fmt.Println("  backup, restore, update, mode, setup")
	fmt.Println()
	fmt.Println("Update Commands:")
	fmt.Println("  check, yes, branch {stable|beta|daily}")
	fmt.Println()
}

func handleStatus(cfg *config.Config, pidFile string) {
	fmt.Println("Status: Not implemented yet")
	// TODO: Implement status check
}
