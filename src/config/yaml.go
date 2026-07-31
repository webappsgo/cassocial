package config

import (
	"bytes"
	"sort"

	"gopkg.in/yaml.v3"
)

// MarshalCommented renders the configuration as YAML with a single-line
// comment (< 140 characters, always ABOVE the setting — see AI.md PART 5
// "YAML Comment Style") above every field. Plain yaml.Marshal(cfg) cannot
// emit these comments, so the document is built manually as a yaml.Node
// tree: each field is encoded into a value node (so the yaml library still
// handles quoting/escaping correctly) with its comment attached as a
// HeadComment on the paired key node.
func (cfg *Config) MarshalCommented() ([]byte, error) {
	root := mapping(
		field("server", "Server configuration",
			mapping(
				field("port", "Default: random unused port in 64xxx range", cfg.Server.Port),
				field("fqdn", "Auto-detected from host", cfg.Server.FQDN),
				field("address", "[::] = all interfaces IPv4/IPv6", cfg.Server.Address),
				field("mode", "production or development", cfg.Server.Mode),
				field("debug", "Enable debug endpoints, pprof, and verbose diagnostics", cfg.Server.Debug),
				field("admin_path", "Admin panel path (default: admin) - see PART 17", cfg.Server.AdminPath),
				field("api_version", "API version prefix (default: v1) - used in /api/{api_version}/ routes", cfg.Server.APIVersion),
				field("healthz", "Optional /healthz root compatibility alias (canonical route stays /server/healthz)",
					mapping(
						field("root", "When true, mount /healthz to the SAME handler as /server/healthz (never redirect)",
							mapping(
								field("enabled", "", cfg.Server.Healthz.Root.Enabled),
							)),
					)),
				field("branding", "Site branding shown in the UI and page metadata - see PART 16",
					mapping(
						field("title", "Site title", cfg.Server.Branding.Title),
						field("tagline", "Short tagline shown alongside the title", cfg.Server.Branding.Tagline),
						field("description", "Longer description used for page metadata", cfg.Server.Branding.Description),
					)),
				field("seo", "Search-engine metadata",
					mapping(
						field("keywords", "Search-engine keywords", cfg.Server.SEO.Keywords),
					)),
				field("user", "System user the server runs as", cfg.Server.User),
				field("group", "System group the server runs as", cfg.Server.Group),
				field("pidfile", "Write a PID file on start", cfg.Server.PIDFile),
				field("daemonize", "Daemonize on start; default false since service managers prefer foreground", cfg.Server.Daemonize),
				field("admin", "Admin panel; username/password/token live in the database, not here",
					mapping(
						field("email", "Admin contact email", cfg.Server.Admin.Email),
					)),
			)),
		field("database", "Database connection settings",
			mapping(
				field("driver", "sqlite, postgres, or mysql", cfg.Database.Driver),
				field("url", "Connection string (overrides host/port/name/user/password when set)", cfg.Database.URL),
				field("host", "Database host (ignored for sqlite)", cfg.Database.Host),
				field("port", "Database port (ignored for sqlite)", cfg.Database.Port),
				field("name", "Database name, or for sqlite the file path", cfg.Database.Name),
				field("user", "Database user (ignored for sqlite)", cfg.Database.User),
				field("password", "Database password (ignored for sqlite)", cfg.Database.Password),
				field("ssl_mode", "Database TLS mode (ignored for sqlite)", cfg.Database.SSLMode),
				field("max_connections", "Maximum open connections", cfg.Database.MaxConns),
				field("max_idle_connections", "Maximum idle connections", cfg.Database.MaxIdleConns),
			)),
		field("logging", "Logging settings",
			mapping(
				field("level", "debug, info, warn, or error", cfg.Logging.Level),
				field("format", "text or json", cfg.Logging.Format),
			)),
		field("ssl", "SSL/TLS settings",
			mapping(
				field("enabled", "Enable HTTPS", cfg.SSL.Enabled),
				field("cert_file", "Manual cert path (optional; leave empty for auto-detection)", cfg.SSL.CertFile),
				field("key_file", "Manual key path (optional; leave empty for auto-detection)", cfg.SSL.KeyFile),
				field("letsencrypt", "Auto-manage certificates via Let's Encrypt", cfg.SSL.LetsEncrypt),
				field("domain", "Domain to request the certificate for", cfg.SSL.Domain),
			)),
		field("email", "SMTP settings for outbound notifications",
			mapping(
				field("enabled", "Enable outbound email", cfg.Email.Enabled),
				field("host", "SMTP server hostname (if set, skips autodetect)", cfg.Email.Host),
				field("port", "SMTP server port (default: 587)", cfg.Email.Port),
				field("username", "SMTP authentication username", cfg.Email.Username),
				field("password", "SMTP authentication password", cfg.Email.Password),
				field("from_name", "Sender display name (default: app title)", cfg.Email.FromName),
				field("from", "Sender email (default: no-reply@{fqdn})", cfg.Email.From),
				field("tls", "Derived from tls_mode; false only when tls_mode is none", cfg.Email.TLS),
				field("tls_mode", "TLS mode: auto, starttls, tls, or none", cfg.Email.TLSMode),
			)),
		field("cassocial", "Cassocial-specific settings",
			mapping(
				field("site_name", "Site name shown throughout the UI", cfg.Cassocial.SiteName),
				field("site_description", "Site description used for page metadata", cfg.Cassocial.SiteDescription),
				field("allow_registration", "Allow new users to self-register", cfg.Cassocial.AllowRegistration),
				field("max_profiles_per_user", "Maximum profiles a single user may create", cfg.Cassocial.MaxProfilesPerUser),
				field("max_links_per_profile", "Maximum links a single profile may contain", cfg.Cassocial.MaxLinksPerProfile),
			)),
		field("scheduler", "Manages all background tasks", schedulerNode(cfg.Scheduler)),
		field("rate_limit", "Per-IP request throttling", rateLimitNode(cfg.RateLimit)),
		field("web", "Frontend configuration",
			mapping(
				field("ui", "Frontend UI preferences",
					mapping(
						field("theme", "dark, light, or auto", cfg.Web.UI.Theme),
					)),
				field("cors", "Allowed CORS origin(s)", cfg.Web.CORS),
			)),
	)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(root); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// schedulerNode builds the scheduler mapping node, including the per-task
// sub-mapping, in a stable (sorted) task order.
func schedulerNode(s SchedulerConfig) *yaml.Node {
	names := make([]string, 0, len(s.Tasks))
	for name := range s.Tasks {
		names = append(names, name)
	}
	sort.Strings(names)

	var taskEntries []*yaml.Node
	for _, name := range names {
		t := s.Tasks[name]
		taskPairs := [][]*yaml.Node{
			field("enabled", "", t.Enabled),
			field("schedule", "Cron schedule expression", t.Schedule),
		}
		if t.RetryOnFail {
			taskPairs = append(taskPairs, field("retry_on_fail", "", t.RetryOnFail))
		}
		if t.RetryDelay != "" {
			taskPairs = append(taskPairs, field("retry_delay", "", t.RetryDelay))
		}
		if t.Retention != 0 {
			taskPairs = append(taskPairs, field("retention", "Maximum number of backups to keep", t.Retention))
		}
		if t.RenewBefore != "" {
			taskPairs = append(taskPairs, field("renew_before", "Renew this long before expiry", t.RenewBefore))
		}
		taskEntries = append(taskEntries, field(name, "", mapping(taskPairs...))...)
	}

	return mapping(
		field("enabled", "", s.Enabled),
		field("tasks", "Built-in tasks with sane defaults (all enabled by default)", &yaml.Node{Kind: yaml.MappingNode, Content: taskEntries}),
	)
}

// rateLimitNode builds the rate_limit mapping node.
func rateLimitNode(rl RateLimitConfig) *yaml.Node {
	rule := func(r RateLimitRule) *yaml.Node {
		return mapping(
			field("requests", "Requests allowed per window, per IP", r.Requests),
			field("window", "Window length in seconds", r.Window),
		)
	}
	return mapping(
		field("enabled", "", rl.Enabled),
		field("read", "Read (GET) endpoints", rule(rl.Read)),
		field("write", "Write (POST/PUT/DELETE) endpoints", rule(rl.Write)),
		field("health", "Health/status endpoints", rule(rl.Health)),
		field("global_burst", "Absolute per-IP ceiling across all endpoint types, per minute", rl.GlobalBurst),
		field("auth", "Auth endpoints - stricter limits, applied independently of the general limits above",
			mapping(
				field("login", "", rule(rl.Auth.Login)),
				field("password_reset", "", rule(rl.Auth.PasswordReset)),
				field("registration", "", rule(rl.Auth.Registration)),
			)),
	)
}

// field builds a key/value node pair for use inside mapping(). comment (if
// non-empty) is attached as a single-line HeadComment above the key, per
// AI.md PART 5 YAML comment style. value may be a Go scalar/slice (encoded
// via yaml.Node.Encode) or an already-built *yaml.Node (for nested mappings).
func field(key, comment string, value interface{}) []*yaml.Node {
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key, HeadComment: comment}

	var valNode *yaml.Node
	if n, ok := value.(*yaml.Node); ok {
		valNode = n
	} else {
		valNode = &yaml.Node{}
		_ = valNode.Encode(value)
	}

	return []*yaml.Node{keyNode, valNode}
}

// mapping builds a YAML mapping node from field() pairs, in the given order.
func mapping(pairs ...[]*yaml.Node) *yaml.Node {
	content := make([]*yaml.Node, 0, len(pairs)*2)
	for _, p := range pairs {
		content = append(content, p...)
	}
	return &yaml.Node{Kind: yaml.MappingNode, Content: content}
}
