#!/usr/bin/env bash
set -euo pipefail

readonly repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly compose_file="${repository_root}/deploy/compose.yaml"
readonly project_name="manoreck-events-test-$$"
readonly port_offset=$((40000 + ($$ % 10000)))
readonly postgres_port=$((port_offset + 1))
readonly nats_port=$((port_offset + 2))

export MANORECK_POSTGRES_HOST_PORT="${postgres_port}"
export MANORECK_NATS_HOST_PORT="${nats_port}"
export MANORECK_NATS_MONITOR_HOST_PORT=$((port_offset + 3))
export MANORECK_DATABASE_URL="postgres://manoreck:manoreck_local@postgres:5432/manoreck?sslmode=disable"

compose() {
  docker compose --project-name "${project_name}" --file "${compose_file}" "$@"
}

cleanup() {
  readonly exit_status=$?
  if ((exit_status != 0)); then
    mkdir -p "${repository_root}/artifacts"
    compose ps --all >"${repository_root}/artifacts/events-status.log" 2>&1 || true
    compose logs --no-color >"${repository_root}/artifacts/events-services.log" 2>&1 || true
    echo "Event infrastructure diagnostics written to artifacts/." >&2
  fi
  compose down --volumes --remove-orphans >/dev/null 2>&1 || true
  exit "${exit_status}"
}
trap cleanup EXIT

compose config --quiet
compose up --detach --wait --wait-timeout 120 postgres nats
compose run --rm migrate up

(
  cd "${repository_root}/backend"
  CGO_ENABLED=0 \
    MANORECK_TEST_DATABASE_URL="postgres://manoreck:manoreck_local@127.0.0.1:${postgres_port}/manoreck?sslmode=disable" \
    MANORECK_TEST_NATS_URL="nats://127.0.0.1:${nats_port}" \
    go test -tags=integration ./internal/platform/outbox
)

echo "Transactional outbox, lease recovery, JetStream delivery, and durable idempotency passed."
