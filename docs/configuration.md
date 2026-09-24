# Configuration contract

Configuration is read from environment variables at process startup and
validated before the process accepts work. Missing, malformed, conflicting, or
unsafe values produce an actionable startup error naming the variable. Runtime
code does not read configuration opportunistically after startup.

## Naming

- Backend and server-only frontend values use the `MANORECK_` prefix.
- Browser-visible values use the `NEXT_PUBLIC_MANORECK_` prefix.
- A value with the `NEXT_PUBLIC_` prefix is public by definition and must never
  contain credentials, signing material, personal data, or internal-only
  endpoints.
- Variables use uppercase snake case and include a unit in the name when the
  value is not self-describing, such as `_SECONDS` or `_BYTES`.

`.env.example` documents a complete non-secret local configuration. Developers
copy it to the ignored `.env` file. The values in `.env.example` are local-only
defaults and must never be reused as production credentials.

## Secrets

- Real secrets are never committed, placed in `NEXT_PUBLIC_` variables, logged,
  included in error messages, or uploaded as CI artifacts.
- CI secrets are scoped to the smallest job that consumes them. The Phase 00 CI
  baseline does not require production or deployment credentials.
- Application configuration types must redact sensitive fields when formatted.
- Secret rotation must not require recompiling either application.

## Environment identity

`MANORECK_ENVIRONMENT` identifies the runtime safety boundary. Initial accepted
values are `local`, `test`, and `sandbox`. No implicit `production` behavior is
introduced during the local-platform roadmap. Sandbox controls must later
verify this value and their persisted environment scope; checking an
environment variable alone is not sufficient isolation.

## API runtime

- `MANORECK_LOG_LEVEL` accepts the structured logger levels `debug`, `info`,
  `warn`, or `error`.
- `MANORECK_HTTP_ADDRESS` is a `host:port` listen address. An empty host binds
  all container interfaces.
- `MANORECK_HTTP_SHUTDOWN_TIMEOUT_SECONDS` bounds graceful HTTP and dependency
  shutdown.
- `MANORECK_READINESS_TIMEOUT_SECONDS` bounds each aggregate dependency
  readiness evaluation.

Timeout values are whole seconds from 1 through 60. Invalid startup
configuration names the affected variable and stops the process before it
accepts work.
