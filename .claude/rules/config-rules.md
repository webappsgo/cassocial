# Config Rules (PART 5, 6, 12)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Never commit config files to repo (server.yml is runtime-generated)
- Never hardcode machine-specific values (hostname, IP, CPU count, memory)
- Never use .env files
- Never use strconv.ParseBool() → use config.ParseBool()
- Never hardcode default port below 1024 in code → detect privilege

## CRITICAL - ALWAYS DO
- Config file: `{config_dir}/server.yml`
- Detect hostname, IP, CPU cores at runtime on target machine
- Support all 40+ boolean variations via config.ParseBool()
- Application modes: production, development
- Default port: 80 (container), random 64xxx (local)

## APPLICATION MODES
| Mode | Description |
|------|-------------|
| production | Full security, minimal logging, no debug endpoints |
| development | Debug logging, profiling endpoints enabled |

## KEY CONFIG PATHS
- Config: `/etc/casapps/cassocial/server.yml`
- Data: `/var/lib/casapps/cassocial/`
- Logs: `/var/log/casapps/cassocial/`
- DB: `/var/lib/casapps/cassocial/db/server.db`

---
For complete details, see AI.md PART 5, 6, 12
