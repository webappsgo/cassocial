# Command-line reference

Cassocial provides the `cassocial` server and the required `cassocial-cli`
companion client.

## Server

```text
cassocial [flags]
```

The server supports `--help`, `--version`, `--color`, `--lang`, `--shell`,
`--config`, `--data`, `--cache`, `--log`, `--backup`, `--pid`, `--address`,
`--port`, `--baseurl`, `--mode`,
`--debug`, `--status`, `--daemon`, `--service`, `--maintenance`, and `--update`.
Run `cassocial --help` for the current descriptions and accepted values.

Configuration is read from `server.yml`; command-line flags take precedence
over environment variables and file values. Service and maintenance commands
depend on the target operating system and installation. Do not place passwords
or tokens in command-line arguments.

The directory flags accept filesystem paths. They create the selected
directories when needed. Relative paths are resolved from the current working
directory; paths containing `..` segments or control characters are rejected.
`--backup` overrides the platform backup location, while `--cache` selects the
cache directory. Equivalent `CONFIG_DIR`, `DATA_DIR`, `CACHE_DIR`, `LOG_DIR`,
and `BACKUP_DIR` environment variables are supported.
`--baseurl` sets the installation URL prefix. Trusted reverse-proxy prefix
headers take precedence when a request arrives through a trusted proxy.

## Companion client

```text
cassocial-cli [options] COMMAND [arguments]
```

The client accepts `--help`, `--version`, `--color`, `--lang`, `--debug`,
`--shell`, `--server`, `--token`, `--token-file`, `--config`, and `--user`.
Server URL resolution uses `--server`, then `CASSOCIAL_SERVER`, then
`server.primary` in `cli.yml`. Token resolution uses `--token`, `--token-file`,
`CASSOCIAL_TOKEN`, `auth.token`, `auth.token_file`, and the default token file,
in that order.

By default, configuration is read from
`~/.config/casapps/cassocial/cli.yml` on Unix-like systems and the matching
`%APPDATA%` directory on Windows. A relative `--config` path is resolved under
that configuration directory. Unix configuration and token files must have
mode `0600`; the client refuses files with broader permissions. Windows ACL
validation is not implemented yet.

| Command | Purpose |
|---------|---------|
| `health` | Read the connected server's health response |
| `profile SLUG` | Read a public profile |
| `links PROFILE_ID` | Read links for a profile the token can access |
| `shortlink create --profile PROFILE_ID --url URL [--code CODE]` | Create a short link |
| `shortlink list` | List the authenticated user's short links |
| `shortlink delete --code CODE` | Delete one of the user's short links |
| `version` | Print client version and build information |

`sl` is an alias for `shortlink`. Responses are printed as JSON. The client
does not currently implement interactive login, a TUI, automatic cluster
failover, or self-update. The `server.cluster` client configuration is not
used for failover yet.

Authentication and permission errors exit with status `77`, connection errors
with `69`, missing resources with `66`, invalid server input with `65`, and
client configuration errors with `78`. Usage errors use `2`; other failures
use `1`.

## Examples

```sh
cassocial --help
cassocial --lang it --status
cassocial-cli --server "$CASSOCIAL_SERVER" profile YOUR_PROFILE
cassocial-cli --server "$CASSOCIAL_SERVER" health
cassocial-cli --server "$CASSOCIAL_SERVER" --token-file "$HOME/.config/casapps/cassocial/token" links PROFILE_ID
cassocial-cli --server "$CASSOCIAL_SERVER" --token-file "$HOME/.config/casapps/cassocial/token" shortlink create --profile PROFILE_ID --url https://your-site.invalid
```

The client API version defaults to `v1` and can be set with `server.api_version`
in `cli.yml`.
