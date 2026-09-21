# Mano Reck

Mano Reck is the implementation of a multi-tenant B2B credit infrastructure
platform. Businesses can borrow, provide financing to their customers, or do
both. BNPL is one financing product built on the platform rather than the
platform's entire domain model.

The product and delivery requirements are maintained in the sibling
`BNPL-skills` repository. This repository contains all application code,
contracts, migrations, local infrastructure, tests, and CI configuration.

## Delivery status

The dependency-ordered roadmap starts with Phase 00. Phase 01 cannot begin
until the foundational repository and CI contracts are complete.

| Phase | Status | Outcome |
|---|---|---|
| 00 — Foundation and production CI | In progress | Repository contracts and repeatable quality pipeline |
| 01 — Local runtime and infrastructure spine | Blocked by Phase 00 | Healthy local stack and transaction-to-event proof |

P00-01 through P00-03 are complete: foundational decisions and repository
commands are recorded, and both applications have buildable, smoke-tested
scaffolds. The next task is **P00-04 — Define shared contract conventions**.
Accepted decisions are recorded in [`docs/adr`](docs/adr/README.md).

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

| Command | Purpose |
|---|---|
| `make bootstrap` | Install locked dependencies |
| `make format` / `make format-check` | Apply or verify formatting |
| `make lint` | Run static analysis |
| `make unit` | Run deterministic unit tests |
| `make integration` | Run tests against real infrastructure boundaries |
| `make e2e` | Run critical browser journeys |
| `make build` | Build production application artifacts |
| `make smoke` | Start and probe both built applications |
| `make generate` / `make generate-check` | Refresh or verify derived source |
| `make migrate-up` / `make migrate-down` | Apply or revert local migrations |
| `make dev` / `make down` | Start or stop the complete local stack |
| `make check` | Run the fast local quality gate |
| `make ci` | Run the non-browser CI gate |

Configuration and generated-source rules are documented in
[`docs/configuration.md`](docs/configuration.md) and
[`docs/generated-files.md`](docs/generated-files.md).

## Architectural boundaries

- Go remains one deployable modular monolith; NATS does not imply
  microservices.
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
