package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Build info - set via -ldflags at build time
var (
	Version      = "dev"
	CommitID     = "unknown"
	BuildDate    = "unknown"
	OfficialSite = "" // Empty = users must use --server flag
)

// osExit is the function called to terminate the process. Overridable in tests.
var osExit = os.Exit

// projectName is the hardcoded internal project name (never changes even if binary is renamed)
const projectName = "cassocial"

func main() {
	var (
		showHelp     = flag.Bool("help", false, "Show help information")
		showHelpS    = flag.Bool("h", false, "Show help information")
		showVersion  = flag.Bool("version", false, "Show version information")
		showVersionS = flag.Bool("v", false, "Show version information")

		// Output control flags (PART 8 — NON-NEGOTIABLE)
		colorMode = flag.String("color", "auto", "Color output mode (always|never|auto)")
		lang      = flag.String("lang", "", "Language code (e.g. en, es, fr); auto-detected from LANG env var")

		server    = flag.String("server", OfficialSite, "Server URL")
		token     = flag.String("token", "", "API token for authentication")
		tokenFile = flag.String("token-file", "", "Read token from file")
		user      = flag.String("user", "", "Target user or org context (@user or +org)")
	)

	flag.Parse()

	// Apply NO_COLOR / --color preference before any output
	if os.Getenv("NO_COLOR") != "" || *colorMode == "never" {
		os.Setenv("NO_COLOR", "1")
	}

	// Apply language preference (auto-detect from LANG env if not set)
	if *lang == "" {
		if l := os.Getenv("LANG"); l != "" {
			*lang = l
		}
	}
	_ = *lang // Language handling deferred to i18n layer

	if *showVersion || *showVersionS {
		printVersion()
		os.Exit(0)
	}

	if *showHelp || *showHelpS || flag.NArg() == 0 {
		printHelp()
		os.Exit(0)
	}

	// Resolve token from all sources
	apiToken := resolveToken(*token, *tokenFile)

	// Require server URL
	serverURL := *server
	if serverURL == "" {
		serverURL = os.Getenv(strings.ToUpper(projectName) + "_SERVER")
	}
	if serverURL == "" {
		fmt.Fprintf(os.Stderr, "error: no server URL specified. Use --server or set %s_SERVER environment variable.\n",
			strings.ToUpper(projectName))
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	c := &client{
		server: strings.TrimRight(serverURL, "/"),
		token:  apiToken,
		user:   *user,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	if err := c.run(cmd, args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type client struct {
	server string
	token  string
	user   string
	http   *http.Client
}

func (c *client) run(cmd string, args []string) error {
	switch cmd {
	case "profile":
		return c.cmdProfile(args)
	case "links":
		return c.cmdLinks(args)
	case "shortlink", "sl":
		return c.cmdShortlink(args)
	case "version":
		printVersion()
		return nil
	default:
		return fmt.Errorf("unknown command: %s (run --help for usage)", cmd)
	}
}

func (c *client) cmdProfile(args []string) error {
	fs := flag.NewFlagSet("profile", flag.ContinueOnError)
	slug := fs.String("slug", "", "Profile slug")
	if err := fs.Parse(args); err != nil {
		return err
	}

	target := *slug
	if target == "" && fs.NArg() > 0 {
		target = fs.Arg(0)
	}
	if target == "" && c.user != "" {
		target = strings.TrimLeft(c.user, "@+")
	}
	if target == "" {
		return fmt.Errorf("profile slug required")
	}

	body, err := c.get("/api/v1/profile?slug=" + target)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}

func (c *client) cmdLinks(args []string) error {
	fs := flag.NewFlagSet("links", flag.ContinueOnError)
	profileID := fs.String("profile", "", "Profile ID")
	if err := fs.Parse(args); err != nil {
		return err
	}

	id := *profileID
	if id == "" && fs.NArg() > 0 {
		id = fs.Arg(0)
	}
	if id == "" {
		return fmt.Errorf("profile ID required")
	}

	body, err := c.get("/api/v1/links?profile_id=" + id)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}

func (c *client) cmdShortlink(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("shortlink subcommand required: create, list, delete")
	}

	switch args[0] {
	case "list":
		body, err := c.get("/api/v1/shortlinks")
		if err != nil {
			return err
		}
		fmt.Println(string(body))
		return nil

	case "create":
		fs := flag.NewFlagSet("shortlink create", flag.ContinueOnError)
		url := fs.String("url", "", "Target URL")
		code := fs.String("code", "", "Custom short code (optional)")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *url == "" && fs.NArg() > 0 {
			*url = fs.Arg(0)
		}
		if *url == "" {
			return fmt.Errorf("--url required")
		}
		payload := map[string]interface{}{"url": *url}
		if *code != "" {
			payload["custom_code"] = *code
		}
		body, err := c.post("/api/v1/shortlinks", payload)
		if err != nil {
			return err
		}
		fmt.Println(string(body))
		return nil

	case "delete":
		fs := flag.NewFlagSet("shortlink delete", flag.ContinueOnError)
		code := fs.String("code", "", "Short code to delete")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *code == "" && fs.NArg() > 0 {
			*code = fs.Arg(0)
		}
		if *code == "" {
			return fmt.Errorf("--code required")
		}
		body, err := c.delete("/api/v1/shortlinks?code=" + *code)
		if err != nil {
			return err
		}
		fmt.Println(string(body))
		return nil

	default:
		return fmt.Errorf("unknown shortlink subcommand: %s", args[0])
	}
}

func (c *client) get(path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.server+path, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *client) post(path string, body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.server+path, strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

func (c *client) delete(path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodDelete, c.server+path, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *client) do(req *http.Request) ([]byte, error) {
	req.Header.Set("User-Agent", fmt.Sprintf("%s-cli/%s", projectName, Version))
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		fmt.Fprintf(os.Stderr, "error: your API token has been revoked or is invalid. Run '%s-cli login' to re-authenticate.\n",
			projectName)
		osExit(1)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return body, nil
}

func resolveToken(flagToken, flagTokenFile string) string {
	if flagToken != "" {
		return flagToken
	}
	if flagTokenFile != "" {
		data, err := os.ReadFile(flagTokenFile)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	if t := os.Getenv(strings.ToUpper(projectName) + "_TOKEN"); t != "" {
		return t
	}
	// Check default token file
	home, err := os.UserHomeDir()
	if err == nil {
		tokenPath := filepath.Join(home, ".config", "casapps", projectName, "token")
		if info, err := os.Stat(tokenPath); err == nil {
			if info.Mode().Perm()&0o077 != 0 {
				fmt.Fprintf(os.Stderr, "warning: token file %s has loose permissions; run 'chmod 0600 %s'\n",
					tokenPath, tokenPath)
			} else {
				data, err := os.ReadFile(tokenPath)
				if err == nil {
					return strings.TrimSpace(string(data))
				}
			}
		}
	}
	return ""
}

func printVersion() {
	binaryName := filepath.Base(os.Args[0])
	fmt.Printf("%s v%s\n", binaryName, Version)
	fmt.Printf("Built: %s\n", BuildDate)
	fmt.Printf("Go: %s\n", runtime.Version()[2:])
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func printHelp() {
	binaryName := filepath.Base(os.Args[0])
	fmt.Printf("%s - Command-line interface for Cassocial\n", binaryName)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s [options] <command> [arguments]\n", binaryName)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help                     Show this help message")
	fmt.Println("  -v, --version                  Show version information")
	fmt.Println("  --color {always|never|auto}    Color output (default: auto; respects NO_COLOR)")
	fmt.Println("  --lang CODE                    Language code (default: auto from LANG env)")
	fmt.Println("  --server URL                   Cassocial server URL")
	fmt.Println("  --token TOKEN                  API token for authentication")
	fmt.Println("  --token-file FILE              Read token from file")
	fmt.Println("  --user NAME                    Target user (@name) or org (+name) context")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  profile [slug]             Get profile information")
	fmt.Println("  links [profile-id]         List profile links")
	fmt.Println("  shortlink create --url URL Create a short link")
	fmt.Println("  shortlink list             List your short links")
	fmt.Println("  shortlink delete --code X  Delete a short link")
	fmt.Println("  version                    Show version information")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Printf("  %s_TOKEN   API token\n", strings.ToUpper(projectName))
	fmt.Printf("  %s_SERVER  Server URL\n", strings.ToUpper(projectName))
	fmt.Println()
}
