# P01-07 handoff — Full-stack integration harness

## Outcome

`make full-stack-test` now builds and starts the complete application and
infrastructure stack from isolated empty volumes, waits for health, probes the
Go and Next.js artifacts, and proves the committed PostgreSQL outbox path
through JetStream to one durable idempotency receipt. The harness verifies the
same correlation ID in HTTP, relay, and consumer logs and always removes its
containers, network, and volumes.

## Decisions

- The proof driver inserts only the existing infrastructure event directly in
  one explicit PostgreSQL transaction. No test-only HTTP route or fake product
  contract was added.
- Docker assigns ephemeral loopback ports to test services, allowing isolated
  projects to run without shared port assumptions. The backend URL is
  discovered before the frontend image is built and embedded as its public API
  origin.
- Compose owns application migration ordering. The backend starts only after
  PostgreSQL migration succeeds and required dependencies are healthy; the
  frontend starts only after backend readiness succeeds.
- The scratch backend image probes its own `/readyz` endpoint through the same
  binary, retaining a non-root, shell-free runtime image.

## Contracts and migrations

- Added backend and frontend services, health checks, build wiring, dependency
  ordering, and application migration execution to `deploy/compose.yaml`.
- Added optional backend/frontend host-port overrides. Existing API, event,
  persistence, tenant, authorization, audit, and financial contracts are
  unchanged.
- Relay and proof-consumer logs now expose lifecycle outcomes with event,
  aggregate, attempt/delivery, and correlation identifiers without payloads or
  secrets.

## Verification

- `make full-stack-test` passes against built non-root application images and
  fresh PostgreSQL, NATS, Redis, SeaweedFS, and Kratos state.
- The harness observes `published:1:1`: one relay attempt, published outbox
  state, and one durable consumer receipt.
- Backend unit tests, Compose validation, shell syntax validation, formatting,
  static analysis, contract validation, and diff checks pass.

## Remaining risks and next task

The infrastructure proof is deliberately not a domain event. Authentication,
tenant context, authorization, and an authenticated browser journey remain
Phase 02 work.

Next: **P02-01 — Freeze identity and authorization contracts**.
