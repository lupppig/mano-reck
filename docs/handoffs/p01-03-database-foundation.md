# P01-03 handoff — Database and migration foundation

## Outcome

The API owns a bounded pgx pool as a required dependency, reports live
PostgreSQL connectivity through readiness, and exposes an application-owned
transaction boundary that repositories can consume without committing. A
pinned `golang-migrate` tool applies reviewed SQL explicitly, outside API
startup.

## Decisions

- Pool construction is lazy: an unavailable database fails `/readyz` but does
  not terminate the API or fail liveness. Maximum and minimum pool sizes are
  validated at startup.
- Database URLs are treated as secrets and redact themselves during ordinary
  and Go-syntax formatting.
- Application operations receive a deliberately small `DBTX` query surface.
  The transaction helper commits only after a successful operation and rolls
  back on returned errors or unfinished paths.
- The first reversible migration creates only the `platform` schema for
  runtime-owned infrastructure. Product and tenant tables remain owned by
  later domain phases.
- UUIDv7, `timestamptz`, tenant-leading constraints/indexes, and integer
  minor-unit money conventions are recorded beside the migrations.

## Contracts and migrations

- Added `000001_platform_schema.up.sql` and its down migration.
- Added validated database URL, maximum connections, and minimum connections
  configuration.
- Added a tools-profile migration service pinned to golang-migrate 4.19.1 and
  its immutable multi-platform digest.
- No domain API, event, tenant table, or financial table was introduced.

## Verification

- The backend builds with Go 1.27.1, and backend unit tests pass.
- Race-enabled backend tests and `go vet ./...` pass in the pinned Go image.
- `make database-test` starts PostgreSQL from an empty isolated volume, applies
  the migration, rolls it back, reapplies it, and proves pool readiness plus
  transaction commit and rollback behavior.
- Migration naming/pairing, Docker Compose validation, formatting, and
  `git diff --check` pass.
- CI now runs the same isolated PostgreSQL integration harness.

## Remaining risks and next task

The pool intentionally exposes no repository or domain table. NATS lifecycle,
the transactional outbox, relay, consumer idempotency, retry classification,
and dead-letter behavior remain the next bounded infrastructure slice.

Next: **P01-04 — Outbox and JetStream foundation**.
