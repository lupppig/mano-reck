#!/usr/bin/env bash
set -euo pipefail

readonly repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly compose_file="${repository_root}/deploy/compose.yaml"
readonly project_name="manoreck-database-test-$$"
readonly postgres_port=$((30000 + ($$ % 10000)))

export MANORECK_POSTGRES_HOST_PORT="${postgres_port}"
export MANORECK_DATABASE_URL="postgres://manoreck:manoreck_local@postgres:5432/manoreck?sslmode=disable"

compose() {
  docker compose --project-name "${project_name}" --file "${compose_file}" "$@"
}

cleanup() {
  readonly exit_status=$?
  if ((exit_status != 0)); then
    mkdir -p "${repository_root}/artifacts"
    compose ps --all >"${repository_root}/artifacts/database-status.log" 2>&1 || true
    compose logs --no-color >"${repository_root}/artifacts/database-services.log" 2>&1 || true
    echo "Database diagnostics written to artifacts/." >&2
  fi
  compose down --volumes --remove-orphans >/dev/null 2>&1 || true
  exit "${exit_status}"
}
trap cleanup EXIT

compose config --quiet
compose up --detach --wait --wait-timeout 120 postgres

compose run --rm migrate up
schema_exists="$(compose exec -T postgres psql --username manoreck --dbname manoreck --tuples-only --no-align --command "SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'platform')")"
[[ "${schema_exists}" == "t" ]] || {
  echo "The platform schema was not created." >&2
  exit 1
}

compose run --rm migrate down 1
schema_exists="$(compose exec -T postgres psql --username manoreck --dbname manoreck --tuples-only --no-align --command "SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'platform')")"
[[ "${schema_exists}" == "f" ]] || {
  echo "The platform schema was not removed by rollback." >&2
  exit 1
}

compose run --rm migrate up
(
  cd "${repository_root}/backend"
  CGO_ENABLED=0 \
    MANORECK_TEST_DATABASE_URL="postgres://manoreck:manoreck_local@127.0.0.1:${postgres_port}/manoreck?sslmode=disable" \
    go test -tags=integration ./internal/platform/database
)

echo "Database migrations, pooling, readiness, commit, and rollback behavior passed."
