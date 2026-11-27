# Service Rules (PART 24, 25)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Run the application as OS root when the service contract requires an unprivileged account.
- Change ownership, service files, or host configuration as a workaround for project defects.
- Assume a service manager or privilege helper is present without detection.

## CRITICAL - ALWAYS DO
- Follow the specified privilege boundary for privileged operations and validate every requested action.
- Support the documented service managers with correct user/system service paths and lifecycle behavior.
- Keep service installation/removal idempotent, explicit, and safe to run without unsupported managers.
- Use the service's actual binary and runtime paths from configuration.

## KEY RULES
- Server Admin means an application administrator and never implies OS root.
- Do not modify global host service configuration during ordinary project tests or development.

For complete details, see AI.md PART 24, 25.
