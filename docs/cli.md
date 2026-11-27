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

With no command, `display.mode: auto` opens the Bubble Tea interface when
stdin and stdout are interactive terminals. Use the filter input to search
available actions, `↑`/`↓` to move, `Enter` to select, `?` for help, and `q` or
`Ctrl+C` to quit. If no valid server is configured, the client first opens a
setup screen to enter a server URL, test its health endpoint, and save it to
`cli.yml`. Required action values are prompted after the menu closes.
Set `display.mode: tui` to require the TUI, or `display.mode: gui` to receive a
`gui_unavailable` error because native GUI support is not built in. Set
`tui.enabled: false` in `cli.yml` to keep command-only behavior. The TUI
defaults to the dark palette, mouse support, and Unicode symbols. Configure
`tui.theme` as `dark`, `light`, or `system`; system mode uses `COLORFGBG` and
falls back to dark. Configure `tui.mouse` and `tui.unicode` to change the other
preferences. A command is required when no interactive terminal is available.
Version and help are flags (`--version`/`-v`, `--help`/`-h`), not commands.
The version output uses the running binary name and includes the version,
short commit, build date, Go version, and OS/architecture.

The client accepts `--help`, `--version`, `--color`, `--format` (or `--output`),
`--lang`, `--debug`, `--shell`, `--server`, `--token`, `--token-file`,
`--config`, `--user`, and `--update`.
`--user` supplies the default profile target when `profile` has no slug;
`@name` is accepted for that user target. An organization prefix is rejected
while PART 35 remains inactive in this project.
Server URL resolution uses `--server`, then `CASSOCIAL_SERVER_PRIMARY`, then
the compatibility alias `CASSOCIAL_SERVER`, then `server.primary` in `cli.yml`,
then the compiled official-site default when available. Token resolution uses
`--token`, `--token-file`,
`CASSOCIAL_TOKEN`, `auth.token`, `auth.token_file`, and the default token file,
in that order. A relative `auth.token_file` path is resolved beside `cli.yml`.
The HTTP timeout comes from `CASSOCIAL_SERVER_TIMEOUT`, then
`server.timeout` in `cli.yml`, then `30s`; the environment value accepts
seconds or a Go duration such as `45s`. Transient GET/HEAD/OPTIONS failures
use `CASSOCIAL_SERVER_RETRY` and `CASSOCIAL_SERVER_RETRY_DELAY`, then
`server.retry` and `server.retry_delay` in `cli.yml`; defaults are three retry
attempts and a one-second exponential backoff. Client errors are not retried,
and requests that may change server data are not repeated against the same
node. If the node returns a server error or a connection failure, the client
tries the next configured cluster URL once with a fresh request; authentication
and other client errors do not trigger failover.
`--format` and `--output` work before the command or after its arguments.
Response format precedence is `--format`/`--output`,
`CASSOCIAL_OUTPUT_FORMAT`, `output.format` in `cli.yml`, then `table`.
Supported formats are `table`, `json`, `yaml`, `plain`, and `csv`.
Color precedence is `--color`, `output.color` in `cli.yml`, `NO_COLOR`, then
terminal detection. `output.color` accepts `auto`, `yes`, or `no`; boolean
values are also accepted for compatibility with the shared binary setting.

By default, configuration is read from
`~/.config/casapps/cassocial/cli.yml` on Unix-like systems and the matching
`%APPDATA%` directory on Windows. A relative `--config` path is resolved under
that configuration directory. Unix configuration and token files must have
mode `0600`; the client refuses files with broader permissions. Windows ACL
validation requires the credential file to be owned by the running user and
to grant access only to that user's SID. Login removes inherited file ACLs and
sets the current user's ACE before validating the result.

When `--server` supplies a valid server and `server.primary` is empty or
invalid, the client saves it as the primary server. When `--token` is supplied
and `auth.token` is empty, the client saves the token to `cli.yml`; a configured
token is never overwritten by the flag. Both writes preserve other config
values and keep Unix permissions at `0600`.

| Command | Purpose |
|---------|---------|
| `login` | Prompt for username and hidden password, complete optional 2FA, and save the token with restrictive permissions |
| `health` | Read the connected server's health response |
| `profile SLUG` | Read a public profile |
| `links PROFILE_ID` | Read links for a profile the token can access |
| `shortlink create --profile PROFILE_ID --url URL [--code CODE]` | Create a short link |
| `shortlink list` | List the authenticated user's short links |
| `shortlink delete --code CODE` | Delete one of the user's short links |

`sl` is an alias for `shortlink`. Responses default to table output; choose a
machine-readable format with `--format json`, `--format yaml`, or
`--format csv`. Login stores
the bearer token in the platform's default token file. When a token from a
cached source is revoked or expired, that source is cleared and the command
exits with status `77` and directs the user to sign in again.
At startup the client reads `/api/autodiscover`, saves the server's current API
version to `server.api_version`, and checks the release for its OS and
architecture. An interactive terminal prompts before installing an available
update. A non-interactive invocation installs automatically only when
`update.auto: true`; `cassocial-cli --update check` reports availability and
`cassocial-cli --update yes` installs it explicitly. Downloads require HTTPS
and a matching SHA-256 checksum. The server publishes a release only when the
operator has provisioned the binary and matching `version.txt` and
`sha256.txt` files in its CLI distribution directory.

Configure automatic non-interactive installation in the protected client
`cli.yml`:

```yaml
update:
  auto: false
  check_interval: per_invocation
  channel: stable
```

When `server.cluster` contains URLs, the client retries connection failures
and server errors against those nodes, including replayable mutations. It does
not fail over on authentication or other client errors. The current server
health response does not publish cluster URLs, so the client cannot discover
them automatically yet. Native GUI is unavailable in the default static
client build.

Authentication and permission errors exit with status `77`, connection errors
with `69`, missing resources with `66`, invalid server input with `65`, and
client configuration errors with `78`. Usage errors use `2`; other failures
use `1`.

## Examples

```sh
cassocial --help
cassocial --lang it --status
cassocial-cli --server "$CASSOCIAL_SERVER_PRIMARY" profile YOUR_PROFILE
cassocial-cli --server "$CASSOCIAL_SERVER_PRIMARY" health
cassocial-cli --update check
cassocial-cli --update yes
cassocial-cli --server "$CASSOCIAL_SERVER_PRIMARY" --token-file "$HOME/.config/casapps/cassocial/token" links PROFILE_ID
cassocial-cli --server "$CASSOCIAL_SERVER_PRIMARY" --token-file "$HOME/.config/casapps/cassocial/token" shortlink create --profile PROFILE_ID --url https://your-site.invalid
```

The client API version defaults to `v1`, is discovered from the server on each
invocation, and is saved as `server.api_version` in `cli.yml`.
