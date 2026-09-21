# Phase 00 handoff — Foundation and production CI

## Outcome

P00-01 through P00-06 are complete. The repository now has an explicit
architecture and toolchain contract, stable non-interactive commands, minimal
Go and Next.js probes, machine-readable API/event/audit conventions, and a
production continuous-integration baseline. Phase 01 is unblocked.

## Decisions

- The backend remains a Go modular monolith; the frontend is Next.js with
  TypeScript and pnpm.
- PostgreSQL is authoritative state. NATS JetStream, Redis, SeaweedFS, and ORY
  Kratos retain the narrow responsibilities recorded in the ADRs.
- UUIDv7, integer minor currency units with ISO currency codes, and UTC
  timestamps are the shared primitive conventions.
- GitHub Actions performs CI only. Production deployment and credentials remain
  out of scope.

The binding rationale is indexed in [`docs/adr`](../adr/README.md).

## Contracts and migrations

- OpenAPI 3.1 defines versioning, request/correlation IDs, idempotency,
  pagination, and the error envelope.
- JSON Schema 2020-12 defines public/internal event envelopes and audit
  metadata.
- No domain API, event, table, or migration was introduced. Migration naming
  and up/down pairing are validated in preparation for P01-03.

## Verification

- Go formatting, `go vet`, race-enabled unit tests, coverage generation, and a
  production build pass.
- Frontend formatting, ESLint, TypeScript, Vitest, and the production Next.js
  build pass with the pinned Node and pnpm versions.
- OpenAPI and JSON Schemas validate; migration and shell-script checks pass.
- Backend and frontend process smoke tests pass.
- Backend and frontend production container images build and answer their
  expected health/shell probes as non-root users.
- The GitHub Actions workflow passes `actionlint`; dependency audit reports no
  known high-severity JavaScript vulnerabilities.

## Remaining risks and next task

CI has not run on GitHub until these commits are pushed, so hosted-runner
behavior remains the final external confirmation. Phase 01 must add real
dependency-backed integration tests; the current integration and E2E commands
are intentionally empty gates because the runtime spine does not exist yet.

Next: **P01-01 — Compose infrastructure**.
