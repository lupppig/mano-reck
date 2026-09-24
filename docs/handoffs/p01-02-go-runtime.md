# P01-02 handoff — Go runtime bootstrap

## Outcome

The API process now loads and validates configuration before accepting work,
emits environment-tagged JSON logs at the configured level, propagates UUIDv7
request and correlation identifiers on standard Go contexts, and performs
bounded graceful shutdown. Required dependencies have one lifecycle owner and
feed a separate readiness probe; ordinary dependency outages do not fail the
liveness probe.

## Decisions

- `/healthz` remains the backward-compatible process-health endpoint and has
  liveness semantics. `/livez` makes those semantics explicit, while `/readyz`
  reports whether every registered required dependency is available.
- Dependency startup follows declaration order and shutdown follows reverse
  order. Partial startup is retained for cleanup, and readiness does not expose
  provider errors or connection details.
- The runtime accepts `local`, `test`, and `sandbox` environments only. Timeout
  values are bounded whole seconds, and invalid configuration names the
  affected variable.
- Request metadata is copied from Gin into `context.Context`, keeping future
  application and infrastructure packages independent of the HTTP framework.

## Contracts and migrations

- OpenAPI now describes `/livez` and `/readyz`, including readiness component
  states and the `503` response.
- `.env.example` adds HTTP shutdown and readiness timeout settings.
- No database migration, domain API, event, tenant, or financial contract was
  introduced.

## Verification

- The backend builds with the pinned Go 1.27.1 container image.
- The final non-root image returns correlated `200` responses from `/healthz`,
  `/livez`, and `/readyz`, then exits cleanly on an interrupt signal.
- `go test -race ./...` passes for configuration, logging, request context,
  dependency lifecycle/readiness, HTTP middleware/probes, runtime cancellation,
  and existing identifier behavior.
- `go vet ./...` passes in the pinned Go container.
- OpenAPI and JSON Schema validation pass.
- Frontend formatting, unit tests, TypeScript checks, and ESLint pass unchanged.
- Docker Compose configuration remains valid, and `git diff --check` passes.

## Remaining risks and next task

No concrete PostgreSQL, NATS, Redis, or object-storage dependency is registered
yet. Their owning tasks will add implementations to the lifecycle/readiness
boundary; until then, readiness proves that the initialized process has no
failed registered dependencies.

Next: **P01-03 — Database and migration foundation**.
