# Makefile Rules (PART 26)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Never use Makefile in CI/CD → explicit commands with env vars in workflows
- Never run `go` directly on host → always via Docker (Makefile internals use golang:alpine)
- Never run `go build` directly → use `make dev` or `make local` or `make build`

## CRITICAL - ALWAYS DO
- All Makefile targets use Docker internally (golang:alpine)
- GODIR and GOCACHE for persistent module/build caching
- Infer PROJECTNAME and PROJECTORG from git remote or path

## MAKE TARGETS
| Target | Purpose | Output |
|--------|---------|--------|
| `make dev` | Development build | `$TMPDIR/casapps/cassocial-XXXXXX/` |
| `make local` | Production testing | `binaries/` (with version) |
| `make build` | Full release | `binaries/` (all 8 platforms) |
| `make test` | Unit tests | Coverage report |
| `make docker` | Docker image | Push to registry |
| `make clean` | Remove artifacts | Removes binaries/ releases/ |

## POST-BUILD DEBUG (via Docker)
```bash
BUILD_DIR=$(ls -td ${TMPDIR:-/tmp}/casapps/cassocial-*/ 2>/dev/null | head -1)
docker run --rm -it -v "$BUILD_DIR:/app" alpine:latest sh -c "
  apk add --no-cache curl bash file jq
  /app/cassocial --help
"
```

---
For complete details, see AI.md PART 26
