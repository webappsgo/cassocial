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
`--version`, `--color`, `--lang`, `--server`, `--token`, `--token-file`,
`--config`, and `--user`. Set the server with `--server`, `CASSOCIAL_SERVER`,
or `server.primary` in `cli.yml`. Token lookup uses, in order, `--token`,
`--token-file`, `CASSOCIAL_TOKEN`, `auth.token` or `auth.token_file` in
`cli.yml`, and the protected per-user token file. `--config` selects a YAML configuration file; otherwise
the client reads `~/.config/casapps/cassocial/cli.yml` on Unix or the
`%APPDATA%` Cassocial config directory on Windows. On Unix, credential files
must be mode `0600` and loose permissions cause the client to refuse them.

| Command | Purpose |
|---------|---------|
| `profile [slug]` | Read a public profile |
| `links [profile-id]` | Read profile links |
| `shortlink create --profile PROFILE_ID --url URL [--code CODE]` | Create a short link |
| `shortlink list` | List the authenticated user's short links |
| `shortlink delete --code CODE` | Delete a short link |
| `version` | Print client version |

`sl` is an alias for `shortlink`. Authentication and permission errors exit
with status `77`, connection failures with `69`, missing resources with `66`,
invalid server input with `65`, and missing client configuration with `78`.
Unknown commands and general errors use `2` and `1`, respectively.

Examples:

```sh
cassocial --help
cassocial --lang it --status
cassocial-cli --server "$CASSOCIAL_SERVER" profile YOUR_PROFILE
cassocial-cli --server "$CASSOCIAL_SERVER" --token-file "$HOME/.config/casapps/cassocial/token" links PROFILE_ID
cassocial-cli --server "$CASSOCIAL_SERVER" --token-file "$HOME/.config/casapps/cassocial/token" shortlink create --profile PROFILE_ID --url https://your-site.invalid
```

The API version defaults to `v1` and can be set with `server.api_version` in
`cli.yml`. The client still lacks required PART 33 features, including login,
interactive mode, cached-token revocation handling, and cluster failover.
Those remain tracked in `TODO.AI.md`.
