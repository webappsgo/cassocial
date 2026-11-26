# Cassocial

Cassocial is a self-hosted link aggregator and social profile platform. It lets people create profile pages with links, profile details, themes, and optional custom domains, managed from a server-rendered web interface.

[![CI](https://github.com/webappsgo/cassocial/actions/workflows/ci.yml/badge.svg)](https://github.com/webappsgo/cassocial/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/webappsgo/cassocial)](https://github.com/webappsgo/cassocial/releases)
[![License](https://img.shields.io/github/license/webappsgo/cassocial)](LICENSE.md)
[![Documentation](https://readthedocs.org/projects/webappsgo-cassocial/badge/?version=latest)](https://webappsgo-cassocial.readthedocs.io/)

## Features

- User accounts with multiple profiles, custom slugs, profile details, and ordered links
- Public and private profile visibility
- Responsive themes, QR codes, and short links
- Profile and link analytics
- Checksum-verified GeoIP datasets with optional country risk filtering
- Daily IP/domain blocklists with a built-in FireHOL source and trusted-IP allowlisting
- Separate Regular User and Server Admin accounts
- Optional custom profile domains
- Server-rendered pages with progressive enhancement
- Companion `cassocial-cli` client
- Eight interface languages: Arabic, German, English, Spanish, French, Italian, Japanese, and Chinese

## Install

Download the matching server and client binaries from [GitHub Releases](https://github.com/webappsgo/cassocial/releases/latest).

| Platform | Architectures |
|---|---|
| Linux | amd64, arm64 |
| macOS | amd64 (Intel), arm64 (Apple Silicon) |
| Windows | amd64, arm64 |
| FreeBSD | amd64, arm64 |

For example, install the Linux amd64 server:

```bash
curl -q -LSsf https://github.com/webappsgo/cassocial/releases/latest/download/cassocial-linux-amd64 -o cassocial
chmod +x cassocial
./cassocial --help
```

On macOS, remove quarantine from downloaded binaries when required:

```bash
xattr -d com.apple.quarantine ./cassocial ./cassocial-cli 2>/dev/null || true
```

For Docker, systemd, service setup, and platform-specific paths, see the [installation guide](docs/installation.md).

## Docker

For Compose, create a persistent deployment directory and download the Compose file there:

```bash
mkdir -p "$HOME/cassocial/volumes/config" "$HOME/cassocial/volumes/data"
cd "$HOME/cassocial"
curl -q -LSsf -o docker-compose.yml https://raw.githubusercontent.com/webappsgo/cassocial/main/docker/docker-compose.yml
docker compose up -d
```

The container publishes port `64580` on the Docker bridge address `172.17.0.1`. Keep the deployment directory and its `volumes/` data when updating the image.

For a single-container deployment with PostgreSQL and Valkey bundled in the
image, download `docker/all-in-one.yml` as `all-in-one.yml` and run
`docker compose -f all-in-one.yml up -d`. The AIO image also includes Tor and
i2pd; only the web port is published.

## Client CLI

Install the `cassocial-cli` binary for your platform from [GitHub Releases](https://github.com/webappsgo/cassocial/releases/latest), then point it at your server:

```bash
chmod +x cassocial-cli-linux-amd64
./cassocial-cli-linux-amd64 --server https://YOUR_SERVER --help
./cassocial-cli-linux-amd64 --server https://YOUR_SERVER health
```

The client can check server health and read public profiles without a token;
profile-link and short-link management use an API token. See the [CLI guide](docs/cli.md) for supported commands and authentication limitations.

## Configuration and API

The server reads `{config_dir}/server.yml`; default paths and supported environment overrides are in the [configuration reference](docs/configuration.md). Server Admins manage the settings currently exposed by the application at `/server/administration`.

The API is served by each Cassocial instance. See the [API reference](docs/api.md) for route families and authentication details. Use your own server address:

```bash
curl -q -LSsf https://YOUR_SERVER/server/healthz
```

## More documentation

- [Administration](docs/admin.md)
- [Security](docs/security.md)
- [Integrations](docs/integrations.md)
- [SMTP setup example](docs/examples/smtp.md)
- [Translation guide](docs/translating.md)
- [Development](docs/development.md)
- [Report a security issue](.github/SECURITY.md)

## Development

Contributor instructions are in the [development guide](docs/development.md). Builds and tests use the Docker Go toolchain described by the project Makefile and AI.md.

```bash
git clone https://github.com/webappsgo/cassocial.git
cd cassocial
make dev
make test
./tests/run_tests.sh
./tests/incus.sh
./tests/e2e.sh
```

`make test` runs the Go suite with coverage. `tests/run_tests.sh` builds the
server and CLI, then checks the compiled binaries and HTTP behavior in an
Alpine container. `tests/incus.sh` runs the full service lifecycle suite in
Incus containers using Debian with systemd and Alpine with OpenRC; pass
`systemd` or `openrc` to run one mode. To run an init-system suite after the Go
coverage gate, use `make test INCUS=systemd`, `make test INCUS=openrc`, or
`make test INCUS=all`.
`tests/e2e.sh` is a separate, on-demand browser suite: it checks server-rendered
HTML, core flows with JavaScript disabled, and full browser behavior with
headless Chromium. It runs its Go tests in Docker and starts a temporary
Chromium sidecar; it is not part of `make test` or `tests/run_tests.sh`.

## Disclaimer

This software is provided "as is" without warranty of any kind. Use it at your own risk. Evaluate its security and operational requirements before production use. Connected third-party services remain subject to their own terms.

## License

See [LICENSE.md](LICENSE.md).
