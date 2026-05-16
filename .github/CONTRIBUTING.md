# Contributing to Cassocial

Thank you for your interest in contributing to Cassocial!

## Getting Started

### Prerequisites

- Docker (all builds run in containers — no local Go installation required)
- Git
- `make`

### Local Setup

```bash
git clone https://github.com/casapps/cassocial.git
cd cassocial

# Quick development build (outputs to temp dir)
make dev

# Run tests
make test

# Build all platforms
make build
```

### Project Structure

```
src/          Go source code
docker/       Dockerfile and compose files
docs/         MkDocs documentation
scripts/      Production/install scripts
tests/        Integration test scripts
```

### Build Targets

| Command | Purpose |
|---------|---------|
| `make dev` | Quick build to temp dir for local testing |
| `make local` | Production build to `binaries/` |
| `make build` | All 8 platforms to `binaries/` |
| `make test` | Unit tests with coverage |
| `make docker` | Build and push Docker image |

> **Note:** All Go operations run inside Docker containers. Do not run `go build` directly.

## Making Changes

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Run `make test` to confirm tests pass
5. Update documentation in `docs/` if you changed user-facing behavior
6. Update `IDEA.md` if you changed features or data models
7. Submit a pull request

### Code Style

- Go code must be formatted with `gofmt`
- Follow the patterns already established in the codebase
- All new behavior needs a test
- No TODO comments in committed code — implement fully or open a tracked issue

### Security

Please **do not** file security vulnerabilities as public issues. See [SECURITY.md](.github/SECURITY.md) for the responsible disclosure process.

## Reporting Bugs

Use the [bug report template](https://github.com/casapps/cassocial/issues/new?template=bug_report.md). Include:
- Cassocial version (`cassocial --version`)
- Operating system and architecture
- Steps to reproduce
- Expected vs actual behavior
- Relevant log output (redact any credentials)

## Requesting Features

Use the [feature request template](https://github.com/casapps/cassocial/issues/new?template=feature_request.md).

## Pull Request Requirements

- Summary of what changed and why
- Test evidence (test output or manual verification steps)
- Documentation updated if behavior changed
- No placeholder, stub, or TODO code introduced
- Breaking changes noted explicitly
