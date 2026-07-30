# TODO

Items flagged during audits/reviews but not yet fixed. Fix fully before removing a line.

## Audit findings (2026-07-30)

### Needs a design decision

- [ ] `src/paths/` — package is entirely dead (no importers anywhere in `src/`),
  and `src/server/config/config.go` duplicates the same path-resolution logic
  (data/config/cache dir resolution). Decide: delete `src/paths/` as dead code,
  OR consolidate `config.go`'s path logic to call into it and rename to the
  Go-singular `src/path/`. Do not just rename — the duplication is the real issue.

### GeoIP is a stub end-to-end (PART 20 country blocking)

- [ ] `src/server/geoip.go:37` — `GeoIP.Lookup` is hardcoded to return
  `("Unknown", "XX", nil)`; it never reads any database. Country blocking
  (`CheckCountryBlocked` / `GeoIPMiddleware`) therefore cannot actually block by
  real country.
- [ ] `src/server/geoip.go:21` vs `:66,79` — path mismatch: `NewGeoIP` sets
  `enabled` by checking for `GeoLite2-Country.mmdb`, but `DownloadDatabase`
  fetches and saves `geoip.csv.gz` (from ip-location-db). After a real download,
  `NewGeoIP` on next start still sees no `.mmdb` and stays disabled. Pick ONE
  format and make load/download/lookup consistent.
- [ ] `src/server/service/analytics_service.go:478` — `getCountryFromIP` only
  distinguishes Local/Unknown; not wired to the `GeoIP` service. Wire it once
  the GeoIP lookup above is real (AnalyticsService needs a GeoIP dependency).
- [ ] `src/server/geoip_test.go` — the whole suite encodes the stub (writes
  `.mmdb` placeholders, asserts `Lookup` returns `"XX"`). Must be rewritten
  alongside the real implementation. Decide implementation approach first:
  pure-Go CSV-range parser (matches the existing `.csv.gz` download, no new dep,
  preferred per CGO-free convention) vs. `oschwald/maxminddb-golang` (adds a dep;
  needs a licensed `.mmdb`).

### Stub handlers duplicating real, wired implementations

- [ ] `src/server/handler/analytics.go:152` — `HandleExportAnalytics` is a stub
  (writes only a CSV header, returns hardcoded-zero JSON, 501 for pdf, and does
  NO ownership check — an IDOR). A real, ownership-checked `AnalyticsHandler.
  ExportAnalytics` already exists in the same package (see
  `analytics_handlers_test.go`). Decide which is routed, wire that one, and
  delete the stub. Its tests (`analytics_handler_test.go`) currently assert the
  stub behavior and must be removed/rewritten with the stub.

- [ ] `src/server/handler/user.go:290` — `Handle2FASetup` returns 501
  "2FA setup not yet implemented" even though a TOTP 2FA service exists and the
  users table has `two_factor_enabled`. Wire the handler to the 2FA service
  (generate secret, return provisioning URI / QR, verify + persist on confirm).

### Minor

- [ ] `src/server/server.go:225` — health check reports `checks["cache"] = "ok"`
  unconditionally with comment "no cache implemented yet". Either implement the
  cache health probe or drop the `cache` key from the health payload so it does
  not report a non-existent subsystem as healthy.
