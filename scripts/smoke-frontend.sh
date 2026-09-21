#!/usr/bin/env bash
set -euo pipefail

readonly repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly smoke_url="http://127.0.0.1:13000"
readonly log_path="$(mktemp /tmp/manoreck-frontend-smoke.XXXXXX.log)"

if [[ "${MANORECK_SMOKE_SKIP_BUILD:-false}" != "true" ]]; then
  (
    cd "${repository_root}"
    pnpm --dir frontend build
  )
fi

(
  cd "${repository_root}"
  pnpm --dir frontend start --hostname 127.0.0.1 --port 13000
) >"${log_path}" 2>&1 &
server_pid=$!
trap 'kill "${server_pid}" >/dev/null 2>&1 || true; wait "${server_pid}" >/dev/null 2>&1 || true; rm -f "${log_path}"' EXIT

for _ in {1..100}; do
  if response="$(curl --fail --silent --show-error "${smoke_url}" 2>/dev/null)"; then
    if [[ "${response}" == *"Credit infrastructure, built deliberately."* ]]; then
      echo "Frontend process smoke test passed."
      exit 0
    fi
    echo "Frontend shell did not contain its expected heading." >&2
    exit 1
  fi
  sleep 0.1
done

echo "Frontend did not become ready at ${smoke_url}." >&2
cat "${log_path}" >&2
exit 1
