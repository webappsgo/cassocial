# CI/CD Rules (PART 28)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Call Makefile targets in CI/CD.
- Use unsafe `pull_request_target` build/test/publish paths or expose secrets/write credentials to fork PRs.
- Use third-party actions without full-SHA pins or permit security jobs to be advisory.

## CRITICAL - ALWAYS DO
- Maintain workflows for GitHub, GitLab, Gitea, Forgejo, and Jenkins; run the provider matching the repository remote.
- Use explicit Docker toolchain commands, least-privilege permissions, and CGO-free builds.
- Include blocking secret, vulnerability, and workflow-policy checks.
- Verify the actual provider CI run after pushing and resolve failures before reporting completion.

## KEY RULES
- Keep local development commands separate from CI commands.
- Respect fork isolation and provider-specific secret handling.

For complete details, see AI.md PART 28.
