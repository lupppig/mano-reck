# ADR 0002: API, persistence, and schema tooling

- Status: Accepted
- Date: 2026-09-21

## Context

Later phases will add independently implemented HTTP handlers, frontend
consumers, workers, provider adapters, and public webhooks. They need common,
machine-readable contracts and predictable database changes before work can
split safely.

## Decision

- Describe the versioned HTTP API in OpenAPI 3.1.1 YAML under
  `contracts/openapi/`. JSON over HTTP is the initial transport.
- Define integration-event and public-webhook payloads as JSON Schema 2020-12
  documents under separate `contracts/events/internal/` and
  `contracts/events/public/` trees. Internal domain events remain Go domain
  types and are not automatically public contracts.
- Treat schema files as source. Generated Go/TypeScript clients, types, or
  documentation are reproducible outputs and must never be hand-edited.
- Use PostgreSQL-specific SQL through `pgx/v5` and `pgxpool`.
- Use `sqlc` to generate type-safe query code for stable repository queries.
  Hand-written `pgx` is allowed for transactions, bulk operations, or queries
  that `sqlc` cannot express clearly; an ORM is not introduced.
- Use `golang-migrate` v4 with ordered, paired plain-SQL `up` and `down`
  migrations. The initial tool version is 4.19.1.
- Application use cases own transaction boundaries. Repositories accept the
  transaction/query context they consume and do not silently commit multi-step
  financial operations.

## Alternatives considered

- GraphQL offers client query flexibility but adds schema/runtime complexity
  without a demonstrated product need.
- An ORM can accelerate simple CRUD, but it obscures PostgreSQL constraints,
  locking, and financial transaction behavior that must remain reviewable.
- Code-first HTTP schemas reduce initial files but allow transport behavior to
  drift before parallel consumers agree on a contract.
- Auto-migration at application startup is convenient locally but weakens
  review and operational control over irreversible schema changes.

## Consequences

Contract changes are reviewed before generated consumers change. PostgreSQL
features and constraints remain explicit. Developers must learn the small
`pgx`/`sqlc` toolchain, and migrations need deliberate compatibility and
rollback analysis.
