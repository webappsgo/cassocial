# Testing Rules (PART 29, 30, 31)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## TESTING (PART 29)
- Go unit tests: `*_test.go` files alongside code
- Integration tests: `tests/` directory at project root
- `tests/run_tests.sh` → auto-detects incus/docker
- `tests/docker.sh` → Docker-based integration tests
- `tests/incus.sh` → Incus-based integration tests (PREFERRED for systemd)
- Coverage goal: 100% (Makefile enforces this)
- Never run tests on host → always via Docker

## READTHEDOCS (PART 30)
- MkDocs configuration: `mkdocs.yml`
- ReadTheDocs config: `.readthedocs.yaml`
- Documentation: `docs/` directory
- Required docs pages: index.md, installation.md, configuration.md, api.md, admin.md, development.md
- Keep docs in sync with code at all times

## I18N & A11Y (PART 31)
- Translation files in `src/common/i18n/locales/`
- Required locales: en (base), es, fr, de, zh, ar, ja
- All 7 locale files MUST be complete (no missing keys)
- WCAG 2.1 AA accessibility minimum
- skip links, ARIA labels, keyboard navigation
- reduced-motion support via CSS
- Touch targets: minimum 44x44px
- `--lang` flag on all binaries (auto-detect from LANG env)

---
For complete details, see AI.md PART 29, 30, 31
