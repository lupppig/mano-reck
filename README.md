# Mano Reck

Mano Reck is the implementation of a multi-tenant B2B credit infrastructure
platform. Businesses can borrow, provide financing to their customers, or do
both. BNPL is one financing product built on the platform rather than the
platform's entire domain model.

## Delivery status

The dependency-ordered roadmap starts with Phase 00. Foundation acceptance is
complete, so Phase 01 can now establish the local runtime and infrastructure
spine.

| Phase                                       | Status      | Outcome                                              |
| ------------------------------------------- | ----------- | ---------------------------------------------------- |
| 00 — Foundation and production CI           | Complete    | Repository contracts and repeatable quality pipeline |
| 01 — Local runtime and infrastructure spine | In progress | Healthy local stack and transaction-to-event proof   |

P00-01 through P00-06 are complete. Foundational decisions and repository
commands are recorded, both applications and production images build and pass
process-level smoke checks, shared transport contracts are machine-readable,
and CI enforces the same quality gates. See the
[Phase 00 handoff](docs/handoffs/phase-00.md) for acceptance evidence. P01-01 is
also complete: the pinned local infrastructure starts from empty volumes and
passes dependency health checks. The next task is **P01-02 — Bootstrap the Go
runtime**. Accepted decisions are recorded in [`docs/adr`](docs/adr/README.md).

## Intended repository layout

```text
backend/       Go modular monolith and background workers
frontend/      Next.js enterprise and developer application
contracts/     OpenAPI and versioned event schemas
deploy/        Docker Compose and local service configuration
docs/          Architecture decisions and engineering documentation
scripts/       Stable, non-interactive repository commands
```

Only directories with implemented content are added to version control.

## Repository commands

Run `make help` for the command list and `make doctor` to validate local
toolchains. The stable command surface is:

| Command                                 | Purpose                                          |
| --------------------------------------- | ------------------------------------------------ |
| `make bootstrap`                        | Install locked dependencies                      |
| `make format` / `make format-check`     | Apply or verify formatting                       |
| `make lint`                             | Run static analysis                              |
| `make contracts`                        | Validate OpenAPI and event/audit schemas         |
| `make migrations-check`                 | Validate migration naming and up/down pairing    |
| `make unit`                             | Run deterministic unit tests                     |
| `make integration`                      | Run tests against real infrastructure boundaries |
| `make infrastructure-test`              | Verify dependencies from isolated empty volumes  |
| `make e2e`                              | Run critical browser journeys                    |
| `make build`                            | Build production application artifacts           |
| `make smoke`                            | Start and probe both built applications          |
| `make generate` / `make generate-check` | Refresh or verify derived source                 |
| `make migrate-up` / `make migrate-down` | Apply or revert local migrations                 |
| `make infra-up` / `make infra-down`     | Start or stop local dependency services          |
| `make infra-reset`                      | Delete local dependency services and their data  |
| `make dev` / `make down`                | Start or stop the complete local stack           |
| `make check`                            | Run the fast local quality gate                  |
| `make ci`                               | Run the non-browser CI gate                      |

Configuration and generated-source rules are documented in
[`docs/configuration.md`](docs/configuration.md) and
[`docs/generated-files.md`](docs/generated-files.md).
Local service endpoints, persistence, and reset behavior are documented in
[`docs/local-infrastructure.md`](docs/local-infrastructure.md).

## Architectural boundaries

- Go remains one deployable modular monolith; NATS does not imply
  microservices.
- Gin is the HTTP transport adapter; domain and application packages do not
  depend on Gin contexts or types.
- PostgreSQL is authoritative transactional state.
- NATS JetStream carries committed asynchronous events through an outbox.
- Redis is limited to ephemeral coordination and caching.
- SeaweedFS stores object bytes behind an application-owned adapter.
- ORY Kratos authenticates identities; the application owns tenants,
  memberships, roles, permissions, and resource authorization.
- External financial, credit, identity, and banking providers begin as
  deterministic mocks behind replaceable interfaces.
- Tenant isolation, authorization, auditability, idempotency, immutable
  financial history, and balanced ledger entries are system invariants.

Production cloud infrastructure and deployment automation are intentionally
out of scope.
