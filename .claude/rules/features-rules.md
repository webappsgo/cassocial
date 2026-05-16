# Features Rules (PART 18-23)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## EMAIL (PART 18)
- SMTP credentials stored in settings table (never config file)
- Failure mode: degrade gracefully with clear user message
- Required: registration verification, password reset, admin notifications
- Never expose SMTP credentials in logs or API responses

## SCHEDULER (PART 19)
- NEVER use external cron or systemd timers
- Built-in scheduler only (robfig/cron or ticker loop)
- Required background jobs: cleanup, analytics retention, GeoIP refresh, backup

## GEOIP (PART 20)
- Local file only (no runtime network fetch)
- MaxMind GeoLite2 or compatible format
- IP hashed with SHA-256 before storage (never raw IP)
- Used for analytics only → NEVER as access gate
- Missing DB: silently return empty country code, still record event

## METRICS (PART 21)
- Prometheus format at `/metrics`
- INTERNAL ONLY → never expose publicly
- Optional bearer token auth for /metrics
- Version, uptime, request counts in healthz (safe)
- Latency histograms, DB connections in /metrics only (not healthz)

## BACKUP & RESTORE (PART 22)
- `--maintenance backup` → backs up DB and config
- Backup location: `/mnt/Backups/casapps/cassocial/`
- Scheduled backups via internal scheduler

## UPDATE COMMAND (PART 23)
- `--update check` → check for new version
- `--update yes` → download and install update
- `--update branch {stable|beta|daily}` → set update channel

---
For complete details, see AI.md PART 18, 19, 20, 21, 22, 23
