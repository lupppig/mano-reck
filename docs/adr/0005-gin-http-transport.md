# ADR 0005: Gin for the HTTP transport

- Status: Accepted
- Date: 2026-09-21

## Context

The initial backend probe used Go's `http.ServeMux` so Phase 00 could validate
the process without prematurely selecting a larger application framework. The
HTTP surface will now grow to include versioned APIs, middleware, request
binding, validation, and transport-level error handling. Contributors need one
router and middleware convention before those handlers multiply.

## Decision

Use Gin 1.12.x, pinned to an exact module version, for backend HTTP routing,
middleware composition, request binding, and response rendering.

Gin remains an adapter concern under the platform and transport packages. The
domain and application layers must not import Gin or expose `gin.Context` in
their interfaces. Handlers translate between Gin and application-owned command,
query, and result types. Go's `http.Server` remains the server lifecycle and
timeout boundary, which preserves standard graceful-shutdown behavior and
ordinary `net/http` testing and observability integration.

Create engines with `gin.New()` and install middleware explicitly. Do not use
Gin's default logger as an additional logging pipeline; application structured
logging remains authoritative. Trusted proxy configuration must be explicit
before forwarded client addresses influence security behavior.

## Alternatives considered

- Go's `http.ServeMux` has no external dependency and is sufficient for basic
  routing, but would require more local conventions and plumbing as the API
  surface grows.
- Chi stays close to `net/http` and has a smaller conceptual surface, but the
  selected framework provides integrated binding, validation, middleware, and
  rendering conventions.
- Fiber provides a broad framework API but does not use `net/http` as its core
  contract, increasing adapter and ecosystem friction for this codebase.

## Consequences

HTTP handlers and middleware can follow one consistent framework convention,
while the module gains Gin and its transitive dependencies. Transport tests
continue to exercise public behavior through `httptest`. Framework types are
reviewed as boundary types and must not leak into business modules.
