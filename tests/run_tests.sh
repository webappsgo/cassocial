#!/usr/bin/env bash
# Integration test runner for cassocial.
# Auto-detects the preferred execution environment: incus > docker > host (not recommended).
# Usage: ./tests/run_tests.sh [--incus|--docker|--host]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "${SCRIPT_DIR}")"

# --- helper ---
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }
info() { printf '[RUN_TESTS] %s\n' "$*"; }

# --- argument parsing ---
FORCE_ENV=""
case "${1:-}" in
    --incus)  FORCE_ENV=incus  ;;
    --docker) FORCE_ENV=docker ;;
    --host)   FORCE_ENV=host   ;;
    "")       ;;
    *)        die "Unknown argument: ${1}" ;;
esac

# --- environment auto-detect ---
if [ -z "${FORCE_ENV}" ]; then
    if command -v incus >/dev/null 2>&1 && incus list >/dev/null 2>&1; then
        FORCE_ENV=incus
    elif command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
        FORCE_ENV=docker
    else
        FORCE_ENV=host
    fi
fi

info "Environment: ${FORCE_ENV}"

case "${FORCE_ENV}" in
    incus)
        exec "${SCRIPT_DIR}/incus.sh" "$@"
        ;;
    docker)
        exec "${SCRIPT_DIR}/docker.sh" "$@"
        ;;
    host)
        info "WARNING: running tests on the host is not recommended."
        info "Running: go test -timeout 300s -count=1 -p 1 ./src/..."
        cd "${PROJECT_DIR}"
        exec go test -timeout 300s -count=1 -p 1 ./src/...
        ;;
    *)
        die "Unknown environment: ${FORCE_ENV}"
        ;;
esac
