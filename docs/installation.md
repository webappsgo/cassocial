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

## First run

On first start, Cassocial writes its default configuration, chooses a port in
the 64000–64999 range, initializes the database, and prints startup details.
If no working SMTP relay is available, the server continues running with
email-dependent flows unavailable.

The first-run setup token and URL are printed to the server console. Open
`/server/administration/setup` (or the configured `server.admin_path` setup
URL) and use the token to create the primary Server Admin. The token is
one-time; keep it private. Server Admin credentials are separate from Regular
User accounts.

## Platform paths

The server uses `internal_org: casapps` and `internal_name: cassocial` for
stable system paths:

| Purpose | Linux / BSD | macOS | Windows |
|---------|-------------|-------|---------|
| Configuration | `/etc/casapps/cassocial` | `/Library/Application Support/casapps/cassocial` | `%PROGRAMDATA%\casapps\cassocial` |
| Data | `/var/lib/casapps/cassocial` | `/Library/Application Support/casapps/cassocial` | `%PROGRAMDATA%\casapps\cassocial` |
| Logs | `/var/log/casapps/cassocial` | `/Library/Logs/casapps/cassocial` | `%PROGRAMDATA%\casapps\cassocial\logs` |

User installs use `~/.config/casapps/cassocial` for configuration and
`~/.local/share/casapps/cassocial` for data and logs. Docker uses `/config`
and `/data` inside the container. `CONFIG_DIR`, `DATA_DIR`, and `LOG_DIR` can
override these locations.

## Upgrading

Back up the deployment's data directory before replacing a binary or updating
the container image. Keep the same persistent mounts after an image update.
Review the release notes for configuration or behavior changes, then restart
the service.
