# Development

Cassocial builds and tests inside the project Go toolchain container. Docker,
Git, and Make are required; installing Go on the host is not required.

## Checkout and build

```sh
git clone https://github.com/webappsgo/cassocial
cd cassocial
make dev
make build
```

`make dev` builds the server and CLI for the host platform into a random,
project-scoped temporary directory. `make build` builds the server and CLI for
Linux, macOS, Windows, and FreeBSD on amd64 and arm64. CGO is disabled for all
builds. Build output is placed under `binaries/` and should not be committed.

`make docker` builds the container image. Publishing the image is handled by
the release workflow; do not use the local build target to publish artifacts.

## Repository layout

```text
src/
├── client/       Companion CLI
├── common/       Shared code and embedded locale files
├── config/       Configuration loading and defaults
├── server/       HTTP handlers, services, models, templates, and storage
├── scheduler/    In-process scheduled tasks
├── service/      Email and notification services
└── main.go       Server entry point
docker/           Container definitions and compose files
docs/             MkDocs documentation
```

## Tests and lint

```sh
make test
bash "$HOME/.codex/hooks/test-lint-run.sh" lint -- docker run --rm \
  --name "cassocial-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
  -v "$PWD:/app" -w /app \
  -e CGO_ENABLED=0 -e GOFLAGS=-buildvcs=false \
  casjaysdev/go:latest sh -c 'go vet ./... && staticcheck ./...'
```

`make test` runs the full Go suite with coverage enforcement, then
`tests/run_tests.sh` builds and checks the server and CLI in a container.
`tests/run_tests.sh` can run that binary tier independently; `tests/incus.sh`
is the separate full service lifecycle check. The lint command uses the
session's test/lint gate wrapper, and all Go tools run in the Go container with
CGO disabled.

## Implementation conventions

- Read the relevant section of `AI.md` before implementing a feature. `AI.md`
  is read-only and is the implementation source of truth.
- Keep the project plan in `IDEA.md` and implementation tasks in `TODO.AI.md`.
- Use server-rendered Go templates and ensure core flows work without
  JavaScript. JavaScript may enhance progressive forms.
- Keep route definitions, API documentation, templates, translations, tests,
  and user documentation in sync.
- Use Argon2id for passwords and SHA-256 hashes for tokens. Do not log
  credentials or private user data.
- Keep Go source under `src/`; do not add a root `cmd/` directory.

## Database changes

AI.md PART 10 requires an idempotent schema bootstrap on every startup, with no
versioned migration files or migration tool. The statement list lives in
`src/server/store/schema_statements.go`; `DB.EnsureSchema` applies it on fresh
and already initialized databases and preserves existing rows. Add schema
changes there as additive, repeatable statements with the introduction
version documented beside each update.

The default SQLite database is `{data_dir}/db/server.db`. On first startup
after upgrading from the earlier `{data_dir}/db/cassocial.db` filename, the
server creates `server.db` from a SQLite online snapshot and leaves the old
database in place as a recovery copy.

## Documentation

The site is built with MkDocs Material. Edit the relevant page under `docs/`,
update `mkdocs.yml` navigation when adding a page, and check links and examples
against the shipped routes and configuration. Update `README.md` for changes
that affect users installing or operating Cassocial.

See the repository's `AI.md` and `IDEA.md` for complete implementation and
product requirements.
