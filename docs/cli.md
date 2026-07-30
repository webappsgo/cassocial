# CLI Reference

Cassocial ships two binaries: the server (`cassocial`, built from `src/main.go`) and a
lightweight API client (`cassocial-cli`, built from `src/client/main.go`). This page
documents both.

## Server Binary (`cassocial`)

```
cassocial [options]
```

### Options

| Flag | Description |
|------|-------------|
| `-h`, `--help` | Show help message |
| `-v`, `--version` | Show version information |
| `--color {always\|never\|auto}` | Color output (default: `auto`; respects `NO_COLOR`) |
| `--lang CODE` | Language code (default: auto-detected from `LANG` env var) |
| `--mode {production\|development}` | Set application mode |
| `--config {dir}` | Configuration directory |
| `--data {dir}` | Data directory |
| `--log {dir}` | Log directory |
| `--pid {file}` | PID file path |
| `--address {addr}` | Listen address |
| `--port {port}` | Listen port |
| `--debug` | Enable debug mode |
| `--status` | Show status and health, then exit |
| `--daemon` | Run as a background daemon |
| `--service {cmd}` | Service management (see below) |
| `--maintenance {cmd}` | Maintenance operations (see below) |
| `--update {cmd}` | Update operations (see below) |

### Service Commands (`--service`)

```
start, stop, restart, reload, --install, --uninstall
```

### Maintenance Commands (`--maintenance`)

```
backup, restore, update, mode, setup
```

`restore` takes the backup filename as a trailing positional argument; `mode` takes
`enable`/`disable` as a trailing positional argument.

### Update Commands (`--update`)

```
check, yes, branch {stable|beta|daily}
```

### Environment

| Variable | Purpose |
|----------|---------|
| `JWT_SECRET` | Secret used to sign session JWTs. If unset, a random secret is generated at startup (sessions do not survive a restart in that case). |

### Examples

```bash
# Run in development mode with debug logging
cassocial --mode development --debug

# Run with explicit config/data/log directories
cassocial --config /etc/cassocial --data /var/lib/cassocial --log /var/log/cassocial

# Check server status
cassocial --status

# Trigger a manual backup
cassocial --maintenance backup

# Install as a system service
cassocial --service --install
```

## API Client (`cassocial-cli`)

```
cassocial-cli [options] <command> [arguments]
```

The client is a thin wrapper around the JSON API — it does not talk to the
database directly and requires a running `cassocial` server to connect to.

### Options

| Flag | Description |
|------|-------------|
| `-h`, `--help` | Show help message |
| `-v`, `--version` | Show version information |
| `--color {always\|never\|auto}` | Color output (default: `auto`; respects `NO_COLOR`) |
| `--lang CODE` | Language code (default: auto-detected from `LANG` env var) |
| `--server URL` | Cassocial server URL |
| `--token TOKEN` | API token for authentication |
| `--token-file FILE` | Read the API token from a file |
| `--user NAME` | Target user (`@name`) or org (`+name`) context |

### Commands

| Command | Description |
|---------|-------------|
| `profile [slug]` | Get profile information (`--slug` also accepted) |
| `links [profile-id]` | List a profile's links (`--profile` also accepted) |
| `shortlink create --url URL [--code CODE]` | Create a short link, optionally with a custom code |
| `shortlink list` | List your short links |
| `shortlink delete --code X` | Delete a short link |
| `version` | Show version information |

`sl` is an accepted alias for `shortlink`.

### Token Resolution

The client resolves an API token in this order:

1. `--token` flag
2. `--token-file` flag
3. `CASSOCIAL_TOKEN` environment variable
4. `~/.config/casapps/cassocial/token` (a warning is printed if this file's
   permissions are looser than `0600`)

### Server URL Resolution

The client resolves the server URL in this order:

1. `--server` flag
2. `CASSOCIAL_SERVER` environment variable

If neither is set, the client exits with an error.

### Environment

| Variable | Purpose |
|----------|---------|
| `CASSOCIAL_TOKEN` | API token |
| `CASSOCIAL_SERVER` | Server URL |

### Examples

```bash
# Look up a public profile
cassocial-cli --server https://cassocial.example.com profile casjay

# List links for a profile ID, authenticated with a token file
cassocial-cli --server https://cassocial.example.com --token-file ~/.config/casapps/cassocial/token \
  links 3f9c1e2a-...

# Create a short link with a custom code
cassocial-cli --server https://cassocial.example.com --token "$CASSOCIAL_TOKEN" \
  shortlink create --url https://example.com/very/long/path --code demo

# List your short links
cassocial-cli --server https://cassocial.example.com --token "$CASSOCIAL_TOKEN" shortlink list
```

If the server responds with `401 Unauthorized`, the client prints a message
telling you to re-authenticate and exits with status `1`.
