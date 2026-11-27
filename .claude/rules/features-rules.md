# Feature Rules (PART 18, 19, 20, 21, 22, 23)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Replace the internal scheduler with system cron or an external scheduler.
- Enable paid tiers, omit required admin controls, or expose backup/restore secrets.
- Report a feature as available when its provider, data source, or job is unavailable.

## CRITICAL - ALWAYS DO
- Implement email/notifications, scheduling, GeoIP, metrics, backup/restore, and update behavior using their specified services and config.
- Validate provider settings and all job inputs; report task state and failures through the required logs/admin views.
- Keep backup data protected, verify archives before restore, and preserve rollback safety.
- Make update checks and installs verify versions, artifacts, and platform compatibility.

## KEY RULES
- Optional providers remain disabled unless the spec defines their auto-detection behavior.
- Run recurring work inside the application scheduler and honor cancellation, retry, and retention rules.
- Keep every feature free and available without subscription gates.

For complete details, see AI.md PART 18-23.
