#!/usr/bin/env bash
set -euo pipefail

readonly repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly migrations_directory="${repository_root}/backend/migrations"

if [[ ! -d "${migrations_directory}" ]]; then
  echo "Missing backend/migrations directory." >&2
  exit 1
fi

declare -A up_migrations=()
declare -A down_migrations=()
migration_count=0

while IFS= read -r migration_path; do
  migration_name="$(basename "${migration_path}")"
  if [[ ! "${migration_name}" =~ ^([0-9]{6}_[a-z0-9_]+)\.(up|down)\.sql$ ]]; then
    echo "Invalid migration name: ${migration_name}" >&2
    echo "Expected NNNNNN_lowercase_name.up.sql or .down.sql." >&2
    exit 1
  fi

  migration_key="${BASH_REMATCH[1]}"
  direction="${BASH_REMATCH[2]}"
  if [[ "${direction}" == "up" ]]; then
    up_migrations["${migration_key}"]="${migration_path}"
  else
    down_migrations["${migration_key}"]="${migration_path}"
  fi
  migration_count=$((migration_count + 1))
done < <(find "${migrations_directory}" -maxdepth 1 -type f -name '*.sql' -print | sort)

for migration_key in "${!up_migrations[@]}"; do
  if [[ -z "${down_migrations[${migration_key}]:-}" ]]; then
    echo "Missing down migration for ${migration_key}." >&2
    exit 1
  fi
done

for migration_key in "${!down_migrations[@]}"; do
  if [[ -z "${up_migrations[${migration_key}]:-}" ]]; then
    echo "Missing up migration for ${migration_key}." >&2
    exit 1
  fi
done

if ((migration_count % 2 != 0)); then
  echo "Migration files are not paired." >&2
  exit 1
fi

echo "Validated $((migration_count / 2)) migration pair(s)."
