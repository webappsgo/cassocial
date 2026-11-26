# Installation

Cassocial is available as a Docker image or a standalone binary. Releases
include Linux, macOS, Windows, and FreeBSD builds for amd64 and arm64.

## Docker Compose

The Compose file mounts persistent data from a `volumes/` directory beside the
file. Keep this deployment directory outside the source checkout. For example:

```sh
mkdir -p "$HOME/cassocial/volumes/config" "$HOME/cassocial/volumes/data"
cd "$HOME/cassocial"
curl -q -LSsf -o docker-compose.yml https://raw.githubusercontent.com/webappsgo/cassocial/main/docker/docker-compose.yml
docker compose up -d
```

The container listens on port 80 and publishes it on the Docker bridge at
`172.17.0.1:64580`. Put a reverse proxy on the same network for external
access. The host-side configuration and data persist under the deployment
directory's `volumes/` tree.

### All-in-One image

The AIO image runs Cassocial, PostgreSQL, and Valkey in one container and
includes Tor and i2pd. Only the web port is exposed. Use a separate persistent
deployment directory:

```sh
mkdir -p "$HOME/cassocial-aio/volumes/config" "$HOME/cassocial-aio/volumes/data"
cd "$HOME/cassocial-aio"
curl -q -LSsf -o all-in-one.yml https://raw.githubusercontent.com/webappsgo/cassocial/main/docker/all-in-one.yml
docker compose -f all-in-one.yml up -d
```

The AIO Compose file publishes port `64580` on the Docker bridge address
`172.17.0.1`. PostgreSQL and Valkey stay internal to the container. Keep the
deployment directory and its `volumes/` data when updating the image.

## Direct Docker run

Run this from a deployment directory with `volumes/config` and `volumes/data`
created first:

```sh
docker run -d \
  --name cassocial \
  -p 172.17.0.1:64580:80 \
  -v ./volumes/config:/config:z \
  -v ./volumes/data:/data:z \
  ghcr.io/webappsgo/cassocial:latest
```

View startup output with `docker logs cassocial`.

## Reverse proxy path prefix

To publish Cassocial below a path such as `/cassocial`, send the trusted
`X-Forwarded-Prefix` header from the reverse proxy. The server uses that
prefix for requests and generated page links; when no trusted proxy header is
present, set `server.baseurl` in `server.yml` or pass `--baseurl`.

```nginx
location /cassocial/ {
    proxy_pass http://cassocial/;
    proxy_set_header X-Forwarded-Prefix /cassocial;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## Binary

Download the release matching the operating system and architecture from
[GitHub Releases](https://github.com/webappsgo/cassocial/releases). For Linux
amd64:

```sh
curl -q -LSsf -O https://github.com/webappsgo/cassocial/releases/latest/download/cassocial-linux-amd64
chmod +x cassocial-linux-amd64
./cassocial-linux-amd64
```

The server creates its configuration and data directories at the platform
paths below. See `cassocial --help` for service-management options supported by
the target operating system.

### Service management

Run `cassocial --service --help` to see the service commands. An administrator
can install a system service with `cassocial --service --install`; a regular
user installs a per-user service where the operating system supports one. The
server uses the systemd user manager on Linux, a LaunchAgent on macOS, and an
ONLOGON scheduled task on Windows. FreeBSD has no per-user service manager, so
the command prints guidance and exits successfully. Linux system installs use
the detected service manager (systemd, OpenRC, SysVinit, runit, or s6), while
FreeBSD uses rc.d and macOS uses launchd.

Use `cassocial --service start|stop|restart|status` to manage an installed
service, and `cassocial --service --disable` to disable automatic startup.
Uninstall with `cassocial --service --uninstall`; this asks for confirmation
before deleting the service definition and the application's configuration,
data, cache, logs, backups, and managed service account. Back up data first.

## First run

On first start, Cassocial writes its default configuration, chooses a port in
the 64000–64999 range, initializes the database, and prints startup details.
If no working SMTP relay is available, the server continues running with
email-dependent flows unavailable.

The first-run setup token and URL are printed to the server console. Open the
configured `/server/{admin_path}` URL and submit the token to enter the
server-rendered setup wizard. The six steps create the primary Server Admin,
generate an API token, configure branding and server settings, configure
security, choose optional services, and complete setup. The raw API token and
optional TOTP recovery codes are shown once. Completion invalidates the setup
token and signs the primary admin in. Keep the setup token private. Server
Admin credentials are separate from Regular User accounts.

## Platform paths

Cassocial keeps the stable `internal_org: casapps` and
`internal_name: cassocial` identity in native OS paths. System and user paths
are separate:

| Platform and run mode | Config | Data | Cache | Logs | Backups |
|-----------------------|--------|------|-------|-------|---------|
| Linux system | `/etc/casapps/cassocial` | `/var/lib/casapps/cassocial` | `/var/cache/casapps/cassocial` | `/var/log/casapps/cassocial` | `/mnt/Backups/casapps/cassocial` |
| Linux user | `~/.config/casapps/cassocial` | `~/.local/share/casapps/cassocial` | `~/.cache/casapps/cassocial` | `~/.local/log/casapps/cassocial` | `~/.local/share/Backups/casapps/cassocial` |
| macOS system | `/Library/Application Support/casapps/cassocial` | `/Library/Application Support/casapps/cassocial/data` | `/Library/Caches/casapps/cassocial` | `/Library/Logs/casapps/cassocial` | `/Library/Backups/casapps/cassocial` |
| macOS user | `~/Library/Application Support/casapps/cassocial` | `~/Library/Application Support/casapps/cassocial` | `~/Library/Caches/casapps/cassocial` | `~/Library/Logs/casapps/cassocial` | `~/Library/Backups/casapps/cassocial` |
| FreeBSD system | `/usr/local/etc/casapps/cassocial` | `/var/db/casapps/cassocial` | `/var/cache/casapps/cassocial` | `/var/log/casapps/cassocial` | `/var/backups/casapps/cassocial` |
| FreeBSD user | `~/.config/casapps/cassocial` | `~/.local/share/casapps/cassocial` | `~/.cache/casapps/cassocial` | `~/.local/log/casapps/cassocial` | `~/.local/share/Backups/casapps/cassocial` |
| Windows administrator | `%PROGRAMDATA%\casapps\cassocial` | `%PROGRAMDATA%\casapps\cassocial\data` | `%PROGRAMDATA%\casapps\cassocial\cache` | `%PROGRAMDATA%\casapps\cassocial\logs` | `%PROGRAMDATA%\Backups\casapps\cassocial` |
| Windows user | `%APPDATA%\casapps\cassocial` | `%LOCALAPPDATA%\casapps\cassocial` | `%LOCALAPPDATA%\casapps\cassocial\cache` | `%LOCALAPPDATA%\casapps\cassocial\logs` | `%LOCALAPPDATA%\Backups\casapps\cassocial` |

On Windows, SSL certificates and configuration live under the Config path,
and the database and security data live under Data; the Windows Service Control
Manager replaces PID files. On other native platforms, SSL and configuration
live under Config, while database and security data live under Data. Unix
system installs keep the PID file in `/var/run/casapps/cassocial/`; user
installs keep it in their data directory. Containers do not create PID files.
Docker uses `/config/cassocial` and `/data/cassocial` inside the container;
those container paths do not apply to native installs. `CONFIG_DIR`,
`DATA_DIR`, and `LOG_DIR` override their respective directories when set.

## Upgrading

Back up the deployment's data directory before replacing a binary or updating
the container image. Keep the same persistent mounts after an image update.
Review the release notes for configuration or behavior changes, then restart
the service.
