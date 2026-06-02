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
	"github.com/casapps/cassocial/src/server/handler"
	"github.com/casapps/cassocial/src/server/store"
)

// Build info - set via -ldflags at build time
var (
	Version      = "dev"
	CommitID     = "unknown"
	BuildDate    = "unknown"
	OfficialSite = "" // Empty = users must use --server flag
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// run parses args and executes the appropriate action, returning an exit code.
// Extracted from main() so it can be tested without exec side-effects.
func run(args []string) int {
	fs := flag.NewFlagSet("cassocial", flag.ContinueOnError)

	// CLI flags (following TEMPLATE.md NON-NEGOTIABLE specification)
	var (
		showHelp     = fs.Bool("help", false, "Show help information")
		showHelpS    = fs.Bool("h", false, "Show help information")
		showVersion  = fs.Bool("version", false, "Show version information")
		showVersionS = fs.Bool("v", false, "Show version information")

		// Output control flags (PART 8 — NON-NEGOTIABLE)
		colorMode = fs.String("color", "auto", "Color output mode (always|never|auto)")
		lang      = fs.String("lang", "", "Language code (e.g. en, es, fr); auto-detected from LANG env var")

		// Directory flags
		configDir = fs.String("config", "", "Configuration directory")
		dataDir   = fs.String("data", "", "Data directory")
		logDir    = fs.String("log", "", "Log directory")
		pidFile   = fs.String("pid", "", "PID file path")

		// Server flags
		address = fs.String("address", "", "Listen address")
		port    = fs.Int("port", 0, "Listen port")
		mode    = fs.String("mode", "", "Application mode (production|development)")

		// Operation flags
		debug      = fs.Bool("debug", false, "Enable debug mode")
		showStatus = fs.Bool("status", false, "Show status and health")
		daemon     = fs.Bool("daemon", false, "Run as daemon")

		// Service management
		service = fs.String("service", "", "Service command (start|stop|restart|reload|--install|--uninstall)")

		// Maintenance
		maintenance = fs.String("maintenance", "", "Maintenance command (backup|restore|update|mode|setup)")
		update      = fs.String("update", "", "Update command (check|yes|branch)")
	)

	if err := fs.Parse(args); err != nil {
		// flag.ContinueOnError writes usage to os.Stderr automatically on error
		return 1
	}

	// Apply NO_COLOR / --color preference before any output
	applyColorPreference(*colorMode)

	// Handle version
	if *showVersion || *showVersionS {
		printVersion()
		return 0
	}

	// Handle help
	if *showHelp || *showHelpS {
		printHelp()
		return 0
	}

	// Apply language preference (auto-detect from LANG env if not set)
	if *lang == "" {
		if l := os.Getenv("LANG"); l != "" {
			*lang = l
		}
	}
	_ = *lang // Language handling deferred to i18n layer

	// Load configuration
	cfg, err := config.Load(*configDir, *dataDir, *logDir)
	if err != nil {
		log.Printf("Failed to load configuration: %v", err)
		return 1
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
		return 0
	}

	if *daemon {
		fmt.Fprintln(os.Stderr, "error: --daemon not yet implemented")
		return 1
	}

	if *service != "" {
		fmt.Fprintln(os.Stderr, "error: --service not yet implemented")
		return 1
	}

	if *maintenance != "" {
		fmt.Fprintln(os.Stderr, "error: --maintenance not yet implemented")
		return 1
	}

	if *update != "" {
		fmt.Fprintln(os.Stderr, "error: --update not yet implemented")
		return 1
	}

	// Write PID file
	if err := server.WritePIDFile(pidFilePath); err != nil {
		log.Printf("Failed to write PID file: %v", err)
		return 1
	}
	// Ensure PID file is removed on exit
	defer server.RemovePIDFile(pidFilePath)

	// Initialize database
	log.Printf("Connecting to database (driver: %s)...", cfg.Database.Driver)
	db, err := store.Connect(cfg.Database.Driver, cfg.Database.Name)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return 1
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

	jwtSecret := os.Getenv("JWT_SECRET")
	authSvc := server.NewAuth(db, jwtSecret)
	router := handler.NewRouter(db, authSvc, cfg)
	h := router.SetupRoutes()

	srv, err := server.New(cfg, db, h)
	if err != nil {
		log.Printf("Failed to create server: %v", err)
		return 1
	}

	// Start server (blocks until shutdown)
	if err := srv.Start(); err != nil {
		log.Printf("Server error: %v", err)
		return 1
	}

	return 0
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
	fmt.Println("  -h, --help                          Show this help message")
	fmt.Println("  -v, --version                       Show version information")
	fmt.Println("  --color {always|never|auto}         Color output (default: auto; respects NO_COLOR)")
	fmt.Println("  --lang CODE                         Language code (default: auto from LANG env)")
	fmt.Println("  --mode {production|development}     Set application mode")
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
	fmt.Println("Status: not yet implemented")
}

// applyColorPreference applies NO_COLOR and --color preference.
// When NO_COLOR is set (any non-empty value) or --color=never, color output is disabled.
func applyColorPreference(mode string) {
	if os.Getenv("NO_COLOR") != "" || mode == "never" {
		// Color disabled — future colored output checks this package-level flag
		os.Setenv("NO_COLOR", "1")
	}
}
