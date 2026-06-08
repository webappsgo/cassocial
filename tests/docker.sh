#!/usr/bin/env bash
# Docker-based integration test runner for cassocial.
# Runs the full Go test suite inside the casjaysdev/go:latest container,
# mirroring the CI environment exactly.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "${SCRIPT_DIR}")"
GO_CACHE="${HOME}/go/pkg"

die()  { printf 'ERROR: %s\n' "$*" >&2; exit 1; }
info() { printf '[DOCKER TESTS] %s\n' "$*"; }

command -v docker >/dev/null 2>&1 || die "docker is not installed or not in PATH"
docker info >/dev/null 2>&1 || die "docker daemon is not running"

info "Project: ${PROJECT_DIR}"
info "Go cache: ${GO_CACHE}"

mkdir -p "${GO_CACHE}"

docker run --rm \
    -v "${PROJECT_DIR}:/workspace" \
    -v "${GO_CACHE}:/go/pkg" \
    -w /workspace \
    casjaysdev/go:latest \
    sh -c "go test -timeout 300s -count=1 -p 1 ./src/... 2>&1"

info "All tests passed."
