# Project Rules (PART 2, 3, 4)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Never put Dockerfile in project root → `docker/Dockerfile`
- Never create config/, data/, logs/ in project root
- Never create CHANGELOG.md (use GitHub/Gitea releases)
- Never create SUMMARY.md, COMPLIANCE.md, NOTES.md
- Never use plural directory names (handlers/, models/) → singular
- Never create vendor/ directory (use Go modules)
- Never hardcode project name or org → infer from git remote

## CRITICAL - ALWAYS DO
- All source in `src/` directory
- Docker files in `docker/` directory
- All 8 platforms: linux/darwin/windows/freebsd × amd64/arm64
- Binary naming: `cassocial-{os}-{arch}` (windows: `.exe`)
- MIT license in LICENSE.md
- Community files in `.github/` (CONTRIBUTING.md, SECURITY.md, etc.)

## PATH RULES
| Placeholder | Linux/BSD Default |
|-------------|-------------------|
| `{config_dir}` | `/etc/casapps/cassocial` |
| `{data_dir}` | `/var/lib/casapps/cassocial` |
| `{db_dir}` | `/var/lib/casapps/cassocial/db/` |
| `{log_dir}` | `/var/log/casapps/cassocial` |
| `{cache_dir}` | `/var/cache/casapps/cassocial` |
| `{backup_dir}` | `/mnt/Backups/casapps/cassocial` |
| `{pid_file}` | `/var/run/casapps/cassocial.pid` |

## ALLOWED ROOT FILES
`AI.md`, `IDEA.md`, `CLAUDE.md`, `README.md`, `LICENSE.md`, `go.mod`, `go.sum`, `Makefile`, `mkdocs.yml`, `.gitignore`, `.dockerignore`, `.gitattributes`, `release.txt`, `site.txt`, `Jenkinsfile`, `.readthedocs.yaml`, `renovate.json`

## ALLOWED ROOT DIRS
`src/`, `docker/`, `docs/`, `scripts/`, `tests/`, `.github/`, `.gitea/`, `.forgejo/`, `.claude/`, `.cursor/`, `.aider/`, `.ai/`, `.windsurf/`, `binaries/`, `releases/`, `volumes/`

---
For complete details, see AI.md PART 2, 3, 4
