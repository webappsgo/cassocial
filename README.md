# Cassocial

[![CI](https://github.com/webappsgo/cassocial/actions/workflows/ci.yml/badge.svg)](https://github.com/webappsgo/cassocial/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/webappsgo/cassocial)](https://github.com/webappsgo/cassocial/releases)
[![License](https://img.shields.io/github/license/webappsgo/cassocial)](LICENSE.md)

## About

Cassocial is a self-hosted link aggregator and social profile platform. It lets people create profile pages with links, profile details, themes, and optional custom domains, managed from a server-rendered web interface.

## Features

- Multiple profiles with custom slugs, profile details, and ordered links
- Public and private profile visibility
- Responsive themes, QR codes, and link shortlinks
- Profile and link analytics
- Regular User accounts (PART 34) and Server Admin accounts
- Optional custom profile domains (PART 36)
- Server-rendered pages with progressive enhancement
- A companion `cassocial-cli` client
- Eight interface languages: English, Spanish, French, Arabic, German, Italian,
  Japanese, and Chinese

## Production

### Docker

```bash
docker run -d \
  --name cassocial \
  -p 172.17.0.1:64580:80 \
  -v ./volumes/config:/config:z \
  -v ./volumes/data:/data:z \
  ghcr.io/webappsgo/cassocial:latest
```

### Docker Compose

```bash
curl -q -LSsf -O https://raw.githubusercontent.com/webappsgo/cassocial/main/docker/docker-compose.yml
docker compose up -d
```

### Binary

Download the binary for your operating system and architecture from the [GitHub releases page](https://github.com/webappsgo/cassocial/releases), then run it:

```bash
curl -q -LSsf -O https://github.com/webappsgo/cassocial/releases/latest/download/cassocial-linux-amd64
chmod +x cassocial-linux-amd64
./cassocial-linux-amd64
```

The server creates its configuration and data directories using the platform paths described in the [installation guide](docs/installation.md). Follow the first-run setup instructions printed by the server to create the primary Server Admin.

For systemd installation and service management, follow the [installation guide](docs/installation.md).

## Client

Download and install the `cassocial-cli` binary for your platform:

```bash
curl -q -LSsf -O https://github.com/webappsgo/cassocial/releases/latest/download/cassocial-cli-linux-amd64
chmod +x cassocial-cli-linux-amd64
sudo mv cassocial-cli-linux-amd64 /usr/local/bin/cassocial-cli
cassocial-cli --server https://YOUR_SERVER --help
```

The client requires the operator's server URL. See the [CLI guide](docs/cli.md) for supported commands and authentication.

## Configuration

Cassocial reads `{config_dir}/server.yml`, where `{config_dir}` follows the platform paths in the [installation guide](docs/installation.md). The server selects an available port in the 64xxx range by default. Server Admins can manage supported settings through `/server/administration`; see the [configuration reference](docs/configuration.md).

## API

The API is served by your Cassocial instance. See the [API reference](docs/api.md) for route families and authentication details. With no project-wide official site configured, use your own server address in requests:

```bash
curl -q -LSsf https://YOUR_SERVER/server/healthz
```

## Other

- [Installation](docs/installation.md)
- [Configuration](docs/configuration.md)
- [Administration](docs/admin.md)
- [Security](docs/security.md)
- [Integrations](docs/integrations.md)
- [Translation guide](docs/translating.md)
- [Development](docs/development.md)
- [Report a security issue](.github/SECURITY.md)

## Development

Development instructions are for contributors. Builds and tests use the Docker toolchain described by the project Makefile and AI.md.

```bash
git clone https://github.com/webappsgo/cassocial.git
cd cassocial
make dev
make test
```

Read the [development guide](docs/development.md) before making changes. The project specification is in `AI.md`, with product scope in `IDEA.md`.

## Disclaimer

This software is provided "as is" without warranty of any kind. Use at your own risk.

- **No Warranty**: The authors are not responsible for damages, data loss, or issues arising from use of this software.
- **Not Professional Advice**: This software does not constitute legal, financial, medical, or other professional advice.
- **Third-Party Services**: Connected services remain subject to their own terms.
- **Security**: No software is guaranteed to be free of vulnerabilities.
- **Production Use**: Evaluate the software and its operational requirements before deploying it.

By using this software, you acknowledge that you have read and understood this disclaimer.

## License

See [LICENSE.md](LICENSE.md).
