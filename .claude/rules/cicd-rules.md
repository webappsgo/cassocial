# CI/CD Rules (PART 28)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Never use Makefile in CI/CD → explicit commands with all env vars
- Never use third-party action tags → pin to full commit SHA
- Never use `pull_request_target` for untrusted code execution
- Never expose secrets to fork PRs
- Never use gitleaks (commercial license) → use truffleHog (Apache-2.0)
- Never use `workflow_run` for cross-workflow ordering → use branch protection

## CRITICAL - ALWAYS DO
- Pin all third-party actions to full commit SHA
- Secret scan (truffleHog) on every push and PR
- govulncheck on every build (when go.sum present)
- Trivy image scan on every Docker build
- All 5 CI providers: GitHub, GitLab, Gitea, Forgejo, Jenkins
- Checksums, SBOM, release notes on every tagged release

## WORKFLOW PERMISSIONS
```yaml
permissions:
  contents: read   # baseline (all jobs)
```
Release job only:
```yaml
permissions:
  contents: write
  packages: write
  id-token: write
  attestations: write
```

## CI/CD PROVIDERS
| Provider | Location |
|----------|----------|
| GitHub | `.github/workflows/build.yml`, `release.yml`, `security.yml` |
| GitLab | `.gitlab-ci.yml` |
| Gitea | `.gitea/workflows/build.yml`, `release.yml`, `security.yml` |
| Forgejo | `.forgejo/workflows/build.yml`, `release.yml`, `security.yml` |
| Jenkins | `Jenkinsfile` |

## VERSION PRECEDENCE
`release.txt` → env var → tag → fallback timestamp

## LDFLAGS IN CI
```
-s -w -X 'main.Version=...' -X 'main.CommitID=...' -X 'main.BuildDate=...' -X 'main.OfficialSite=...'
```

---
For complete details, see AI.md PART 28
