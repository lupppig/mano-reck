#!/usr/bin/env bash
set -euo pipefail

if [[ ! -d backend ]]; then
  echo "backend/ is missing; complete P00-03 before running format-check." >&2
  exit 1
fi

unformatted="$(find backend -type f -name '*.go' -not -path '*/vendor/*' -exec gofmt -l {} +)"
if [[ -n "${unformatted}" ]]; then
  echo "Go files require formatting:" >&2
  echo "${unformatted}" >&2
  echo "Run 'make format'." >&2
  exit 1
fi
