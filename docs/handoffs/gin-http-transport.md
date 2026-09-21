# Gin HTTP transport handoff

## Outcome

The backend HTTP adapter now uses Gin 1.12.0 instead of `http.ServeMux` while
retaining Go's `http.Server` as the lifecycle, timeout, and graceful-shutdown
boundary. Gin is restricted to transport packages and is not a domain or
application-layer dependency.

## Contracts and migrations

- `GET /healthz`, its JSON body, and UUIDv7 request/correlation response headers
  are preserved.
- Unsupported methods continue to return `405 Method Not Allowed` with an
  `Allow` header.
- No OpenAPI, event, persistence, or database migration changed.

## Verification

- Race-enabled backend tests, `go vet`, and the production Go build pass.
- The backend process and rebuilt non-root production container pass the health
  smoke probe.
- `govulncheck` reports no reachable vulnerabilities. Gin initially resolved a
  vulnerable `quic-go` release, so the fixed v0.59.1 transitive version is
  pinned explicitly.

## Remaining risks

P01-02 still needs the complete runtime middleware contract, including
structured panic recovery, validated configuration, dependency-aware readiness,
and graceful-shutdown integration tests. Trusted proxies are disabled until a
real deployment topology defines them explicitly.
