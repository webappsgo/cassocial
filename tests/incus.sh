#!/usr/bin/env bash
# Incus-based integration test runner for cassocial.
# Preferred over Docker for tests that require systemd (service lifecycle, socket activation).
# Creates a temporary Incus container, mounts the project directory, runs all tests,
# then destroys the container regardless of outcome.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "${SCRIPT_DIR}")"
CONTAINER_NAME="cassocial-test-$(date +%s)"
IMAGE="${INCUS_IMAGE:-images:alpine/3.20}"

die()  { printf 'ERROR: %s\n' "$*" >&2; exit 1; }
info() { printf '[INCUS TESTS] %s\n' "$*"; }

command -v incus >/dev/null 2>&1 || die "incus is not installed or not in PATH"
incus list >/dev/null 2>&1 || die "incus daemon is not running or current user lacks access"

cleanup() {
    info "Cleaning up container: ${CONTAINER_NAME}"
    incus delete --force "${CONTAINER_NAME}" 2>/dev/null || true
}
trap cleanup EXIT

info "Launching container: ${CONTAINER_NAME} (${IMAGE})"
incus launch "${IMAGE}" "${CONTAINER_NAME}"

info "Waiting for container to be ready..."
sleep 3

info "Installing Go toolchain..."
incus exec "${CONTAINER_NAME}" -- sh -c "apk add --no-cache go git bash curl 2>&1"

info "Mounting project directory..."
incus config device add "${CONTAINER_NAME}" workspace disk \
    source="${PROJECT_DIR}" path=/workspace

info "Running tests..."
incus exec "${CONTAINER_NAME}" -- sh -c "
    cd /workspace
    go test -timeout 300s -count=1 -p 1 ./src/... 2>&1
"

info "All tests passed."
