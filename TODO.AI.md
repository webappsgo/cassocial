# TODO

## [ ] Config struct completeness
Read: AI.md PART 5

`src/config/config.go`'s `Config` struct is far short of AI.md PART 5's full
example structure. Missing fields/sections: `admin_path`, `api_version`,
`healthz.root`, `branding`, `seo`, `user`/`group`, `pidfile`, `daemonize`,
`admin.email`, `ssl.letsencrypt`, `scheduler`, `rate_limit`, `web`.

## [ ] server.yml comments
Read: AI.md PART 5

Generated `server.yml` has zero comments because `yaml.Marshal(cfg)` cannot
emit the spec-mandated inline comments (single-line, <140 chars). Needs a
custom marshal/template approach instead of the plain `yaml.Marshal` call.

## [ ] Flags > env > file > defaults precedence
Read: AI.md PART 5

Full precedence layering is not implemented in `config.go`. `main.go`
currently applies `--mode`/`--debug` CLI flags after `Load()` as a partial
patch, but this is not integrated as a proper precedence layer, and
`--debug=false` cannot be explicitly expressed due to `flag.Bool`
limitations (need `flag.Func` or a tri-state flag instead).

## [ ] README.md missing Client and Disclaimer sections
Read: AI.md PART 1

Current README.md section order is Title, About, Official Site, Features,
Production, Configuration, API, Other, Development, License — missing the
mandatory `## Client` section (client install/usage, examples connecting to
`{official_site}` by default) and the mandatory `## Disclaimer` section
(before or after License).

## [ ] Debug endpoints (--debug / DEBUG=true)
Read: AI.md PART 6

No `src/server/debug.go` exists. PART 6 mandates pprof endpoints
(`/debug/pprof/*`), expvar (`/debug/vars`), and custom debug endpoints
(`/debug/config`, `/debug/routes`, `/debug/cache`, `/debug/db`,
`/debug/scheduler`) gated behind `IsDebug()`, returning 404 otherwise.

## [ ] Makefile missing GO_CACHE/GO_BUILD mkdir guard on test/dev targets
Read: AI.md PART 26

`test` target (Makefile line 219-221) and `dev` target (line 238-239) invoke
`GO_DOCKER` (which bind-mounts `GO_CACHE`/`GO_BUILD`) without a preceding
`@mkdir -p $(GO_CACHE) $(GO_BUILD)` as the first recipe line, unlike other
targets that create these host cache dirs before use.

## [ ] CLI missing --shell flag (server and client)
Read: AI.md PART 8

`src/main.go` and `src/client/main.go` are missing the `--shell` flag.
PART 8 requires all binaries to support the shared flag set: `--help`,
`--version`, `--shell`, `--debug`, `--color`, `--lang`.

## [ ] Client CLI --color values and missing --debug flag
Read: AI.md PART 8

`src/client/main.go` line 46: `--color` accepts `always|never|auto` but
PART 8 (line ~10679, ~10713) requires `auto|yes|no`. `src/client/main.go`
is also missing the `--debug` flag entirely (server has it, client does not).
