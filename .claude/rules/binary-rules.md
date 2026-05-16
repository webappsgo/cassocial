# Binary Rules (PART 7, 8, 33)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Never use CGO (CGO_ENABLED=0 always)
- Never add -musl suffix to binary names
- Never build Go directly on host → always via Docker (make dev/local/build)
- Never omit client binary → cassocial-cli is REQUIRED

## CRITICAL - ALWAYS DO
- Build all 8 platforms: linux/darwin/windows/freebsd × amd64/arm64
- Binary names: `cassocial-{os}-{arch}` (cassocial-linux-amd64, etc.)
- Windows adds .exe: `cassocial-windows-amd64.exe`
- Local build: `cassocial` (no platform suffix)
- Client binary: `cassocial-cli`
- Build source: `./src` directory always

## CLI FLAGS (NON-NEGOTIABLE)
```
--help (-h)
--version (-v)
--mode {production|development}
--config {config_dir}
--data {data_dir}
--log {log_dir}
--pid {pid_file}
--address {listen}
--port {port}
--baseurl {path}
--debug
--status
--service {start,restart,stop,reload,--install,--uninstall,--disable,--help}
--daemon
--maintenance {backup,restore,update,mode,setup,--help}
--update [check|yes|branch {stable|beta|daily}]
```

Short flags: ONLY -h (help) and -v (version). All others: long form only.

## LDFLAGS
```
-s -w -X 'main.Version=...' -X 'main.CommitID=...' -X 'main.BuildDate=...' -X 'main.OfficialSite=...'
```

---
For complete details, see AI.md PART 7, 8, 33
