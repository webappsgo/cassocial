# Service Rules (PART 24, 25)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Never run as root in production (operator runs under non-root user)
- Never allow `Bash(sudo:*)` in Claude settings
- Never bypass privilege escalation security model

## CRITICAL - ALWAYS DO
- `--service --install` → create and enable systemd unit
- `--service --uninstall` → remove systemd unit
- `--service start|stop|restart|reload` → manage service state
- `--daemon` → daemonize (detach from terminal, write PID file)
- systemd unit file location: `/etc/systemd/system/cassocial.service`

## SYSTEMD UNIT REQUIREMENTS
- Service user/group: `cassocial`/`cassocial`
- `Restart=on-failure`
- `RestartSec=5`
- `WantedBy=multi-user.target`
- Never run as root (operator responsibility)

## SERVICE SUPPORT (PART 25)
- Systemd (Linux, primary)
- runit (alternative)
- rc.d (BSD)
- launchd (macOS LaunchDaemon)
- Windows SCM (via Service Control Manager)
- macOS plist: `io.github.casapps.cassocial`

---
For complete details, see AI.md PART 24, 25
