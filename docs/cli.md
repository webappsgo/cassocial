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
`Ctrl+C` to quit. Required action values are prompted after the menu closes.
Set `display.mode: tui` to require the TUI, or `display.mode: gui` to receive a
`gui_unavailable` error because native GUI support is not built in. Set
`tui.enabled: false` in `cli.yml` to keep command-only behavior. The TUI
defaults to the dark palette, mouse support, and Unicode symbols. Configure
`tui.theme` as `dark`, `light`, or `system`; system mode uses `COLORFGBG` and
falls back to dark. Configure `tui.mouse` and `tui.unicode` to change the other
preferences. A command is required when no interactive terminal is available.

The client accepts `--help`, `--version`, `--color`, `--lang`, `--debug`,
`--shell`, `--server`, `--token`, `--token-file`, `--config`, and `--user`.
`--user` supplies the default profile target when `profile` has no slug;
`@name` is accepted for that user target. An organization prefix is rejected
while PART 35 remains inactive in this project.
Server URL resolution uses `--server`, then `CASSOCIAL_SERVER`, then
`server.primary` in `cli.yml`, then the compiled official-site default when
available. Token resolution uses `--token`, `--token-file`,
`CASSOCIAL_TOKEN`, `auth.token`, `auth.token_file`, and the default token file,
in that order. A relative `auth.token_file` path is resolved beside `cli.yml`.

By default, configuration is read from
`~/.config/casapps/cassocial/cli.yml` on Unix-like systems and the matching
`%APPDATA%` directory on Windows. A relative `--config` path is resolved under
that configuration directory. Unix configuration and token files must have
mode `0600`; the client refuses files with broader permissions. Windows ACL
validation requires the credential file to be owned by the running user and
to grant access only to that user's SID. Login removes inherited file ACLs and
sets the current user's ACE before validating the result.

| Command | Purpose |
|---------|---------|
| `login` | Prompt for username and hidden password, complete optional 2FA, and save the token with restrictive permissions |
| `health` | Read the connected server's health response |
| `profile SLUG` | Read a public profile |
| `links PROFILE_ID` | Read links for a profile the token can access |
| `shortlink create --profile PROFILE_ID --url URL [--code CODE]` | Create a short link |
| `shortlink list` | List the authenticated user's short links |
| `shortlink delete --code CODE` | Delete one of the user's short links |
| `version` | Print client version and build information |

`sl` is an alias for `shortlink`. Responses are printed as JSON. Login stores
the bearer token in the platform's default token file. When a token from a
cached source is revoked or expired, that source is cleared and the command
exits with status `77` and directs the user to sign in again.
When `server.cluster` contains URLs learned from a server health response, the
client retries connection failures and server errors against those nodes and
uses the first working node for the current process. The server health
response does not publish cluster URLs because the AI.md health requirements
conflict on whether addresses may be exposed. Native GUI and self-update
remain pending.

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
