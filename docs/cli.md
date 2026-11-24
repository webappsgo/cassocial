# Command-line reference

Cassocial provides the `cassocial` server and the required `cassocial-cli`
companion client. Both binaries are built from this repository and share a
release version.

## Server

```text
cassocial [flags]
```

The server accepts `--help`, `--version`, `--color`, `--lang`, `--shell`,
`--config`, `--data`, `--log`, `--pid`, `--address`, `--port`, `--debug`,
`--status`, `--daemon`, `--service`, `--maintenance`, and `--update`. Use
`cassocial --help` for the current descriptions and accepted values. `--lang`
accepts a supported locale; unsupported values fall back to English.

Service operations are selected with `--service`: `start`, `stop`, `restart`,
`reload`, `--install`, or `--uninstall`. Installation is supported on Linux
with systemd.

Maintenance operations are selected with `--maintenance`: `backup`,
`restore FILE`, `update`, `mode enable|disable`, or `setup`. Update operations
are selected with `--update`: `check`, `yes`, or `branch stable|beta|daily`.

Configuration is read from `server.yml` under the configured config directory.
The command-line flags override the corresponding server settings. Do not put
secrets in command-line arguments; use the protected configuration and secrets
files described in [Configuration](configuration.md).

## Companion client

```text
cassocial-cli [flags] COMMAND [arguments]
```

The client talks to a Cassocial server over its HTTP API. It supports `--help`,
`--version`, `--color`, `--lang`, `--server`, `--token`, `--token-file`, and
`--user`. Set the server with `--server` or `CASSOCIAL_SERVER`. Token lookup
uses, in order, `--token`, `--token-file`, `CASSOCIAL_TOKEN`, and the protected
per-user token file. On Unix, credential files must be mode `0600`.

| Command | Purpose |
|---------|---------|
| `profile [slug]` | Read a public profile |
| `links [profile-id]` | Read profile links |
| `shortlink create --url URL [--code CODE]` | Create a short link |
| `shortlink list` | List the authenticated user's short links |
| `shortlink delete --code CODE` | Delete a short link |
| `version` | Print client version |

`sl` is an alias for `shortlink`. The client currently has no login command,
interactive TUI, saved YAML configuration, or cluster failover. Those are
required by PART 33 and remain implementation work; the command reference must
be updated when they are delivered. Authentication errors currently exit with
status 1.

Examples:

```sh
cassocial --help
cassocial --lang it --status
cassocial-cli --server "$CASSOCIAL_SERVER" profile YOUR_PROFILE
cassocial-cli --server "$CASSOCIAL_SERVER" --token-file "$HOME/.config/casapps/cassocial/token" links
```

See [AI.md PART 33](../AI.md) for the complete required client behavior.
