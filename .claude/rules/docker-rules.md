# Docker Rules (PART 27)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Never put Dockerfile in project root → `docker/Dockerfile`
- Never put docker-compose.yml in project root → `docker/docker-compose.yml`
- Never use .env files with docker-compose
- Never modify ENTRYPOINT or CMD → customize via entrypoint.sh
- Never symlink or copy pre-built binaries into Docker build context (multi-stage builds only)
- Never use -musl suffix on Alpine builds

## CRITICAL - ALWAYS DO
- Multi-stage Dockerfile: builder (golang:alpine) + runtime (alpine:latest)
- Dockerfile: `docker/Dockerfile`
- docker-compose.yml: `docker/docker-compose.yml`
- Container root overlay: `docker/rootfs/`
- ENTRYPOINT: `["tini", "-p", "SIGTERM", "--", "/usr/local/bin/entrypoint.sh"]`
- STOPSIGNAL: `SIGRTMIN+3`
- Default timezone: `America/New_York` (override with TZ env var)
- Internal port: 80 (override with PORT env var)
- External port: random 64xxx mapped to 80 (e.g., 64580:80)
- Required packages: git, curl, bash, tini, tor

## PORT BEHAVIOR
| Context | Address | Port |
|---------|---------|------|
| Container (default) | 0.0.0.0 | 80 |
| Container (custom) | 0.0.0.0 | PORT env |
| Local (dev) | 0.0.0.0 | Random 64xxx |

## DOCKER TAGS
| Trigger | Tags |
|---------|------|
| Any push | `devel`, `{commit}` |
| Beta branch | adds `beta` |
| Release tag | `{version}`, `latest`, `YYMM`, `{commit}` |

---
For complete details, see AI.md PART 27
