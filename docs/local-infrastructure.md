# Local infrastructure

The Phase 01 runtime uses Docker Compose for PostgreSQL, NATS JetStream, Redis,
SeaweedFS, and ORY Kratos. Images are pinned by exact release and immutable
multi-platform digest in `deploy/compose.yaml`; upgrades intentionally change
both values in one reviewed commit.

## Start and stop

Docker Engine with Compose v2 is required. No `.env` file is required for the
safe local defaults:

```sh
make infra-up
make infra-down
```

`infra-up` returns only after all long-running services are healthy and the
one-shot Kratos database migration has succeeded. `infra-down` removes
containers and the project network but preserves named volumes.

Copy `.env.example` to `.env` when running applications or overriding a host
port. Compose automatically reads `.env`; container-to-container addresses
remain stable service names regardless of host-port overrides.

| Service       | Host endpoint           | Local purpose                       |
| ------------- | ----------------------- | ----------------------------------- |
| PostgreSQL    | `127.0.0.1:5432`        | Authoritative application state     |
| NATS          | `127.0.0.1:4222`        | JetStream event transport           |
| NATS monitor  | `http://127.0.0.1:8222` | Local health and diagnostics        |
| Redis         | `127.0.0.1:6379`        | Ephemeral coordination and cache    |
| SeaweedFS S3  | `http://127.0.0.1:8333` | Object bytes                        |
| Kratos public | `http://127.0.0.1:4433` | Browser and session API             |
| Kratos admin  | `http://127.0.0.1:4434` | Server-side identity administration |

All ports bind only to loopback. Credentials in the Compose and example files
are deliberately non-secret local values and must never be reused outside a
local/test environment.

## Application migrations

Database migrations are explicit and are never applied by API startup. After
the infrastructure is healthy, use:

```sh
make migrate-up
make migrate-version
make migrate-down
```

The commands run the pinned `golang-migrate` image against SQL files in
`backend/migrations`. `make database-test` creates an isolated PostgreSQL
project, applies all migrations, rolls them back, reapplies them, and verifies
pool readiness and transaction commit/rollback behavior.

`make event-test` creates isolated PostgreSQL and NATS services, applies every
migration, and proves that committed outbox rows reach a durable JetStream
consumer. It also verifies rollback exclusion, recovery from an expired relay
lease, and idempotent handling of duplicate event delivery.

`make adapters-test` creates isolated Redis and SeaweedFS services with a
unique key prefix and bucket. It proves Redis health and an expiring value
round trip, then creates the bucket and uploads, reads, checks, and deletes one
opaque object through the application adapter.

## Storage and reset

PostgreSQL, JetStream, Redis, and SeaweedFS use Compose-managed named volumes.
Redis data is still semantically disposable even though local persistence makes
ordinary restarts less surprising. Kratos stores its state in the dedicated
`kratos` PostgreSQL database initialized alongside the application database.

```sh
make infra-reset
```

`infra-reset` permanently deletes all local Compose volumes and their contents.
Use it to reproduce first-start behavior or recover from intentionally
incompatible local schema changes. It does not remove images or unrelated
Docker projects.

## Acceptance test

```sh
make infrastructure-test
```

The test creates a uniquely named Compose project on temporary high ports,
starts with new volumes, waits for readiness, and verifies PostgreSQL database
initialization, JetStream health, Redis PING, the SeaweedFS S3 listener, and
Kratos readiness. It always removes its containers, network, and test volumes.
On failure, service status and logs are written to the ignored `artifacts/`
directory.

Kratos self-service UI routes remain deferred to P01-06. SeaweedFS bucket
creation belongs to the application adapter; infrastructure startup alone does
not create application-owned objects or metadata.
