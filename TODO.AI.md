# TODO

Items flagged during audits/reviews but not yet fixed. Fix fully before removing a line.

## go-lint findings (2026-07-30)

- [ ] `Makefile:47` — `GO_DOCKER` image is `golang:alpine`; must be `casjaysdev/go:latest`
- [ ] `Makefile:41-47` — `GO_DOCKER` invocation missing `-e GOFLAGS=-buildvcs=false`
- [ ] `Makefile:67,77,90,104,127,133,140,236,239,243` — `go build` calls missing inline `-buildvcs=false`
- [ ] `Makefile:22` — `LDFLAGS` missing `-trimpath` for build/release/docker targets
- [ ] `src/main.go:42` — `--color` flag only accepts `always|never|auto`; must support `auto|yes|no`
- [ ] `src/paths` — directory name is plural; Go convention requires singular `src/path/`
- [ ] `go.mod:24` — `github.com/robfig/cron/v3` is an external cron dependency; spec requires built-in scheduler (no forbidden third-party cron lib)
