# P01-05 handoff — Infrastructure adapters

## Outcome

The API now owns narrow Redis and S3-compatible object-storage adapters as
required runtime dependencies. Both contribute to readiness. Redis operations
are namespaced and require expiry, while SeaweedFS startup ensures one configured
bucket and supports an opaque UUIDv7-keyed object round trip.

## Decisions

- Redis exposes only put, get, and delete for expiring bytes. Requiring a
  positive TTL prevents it from becoming authoritative application state.
- A configured key prefix isolates runtimes and tests without defining any
  business cache policy.
- Object-storage callers receive provider-neutral bytes and metadata. The
  adapter owns S3 translation, bucket lifecycle, and not-found normalization.
- Object keys must be UUIDv7 values. Content allowlists, size limits, checksums,
  tenant metadata, authorization, and retention remain with their future
  business owner rather than being guessed here.

## Contracts and migrations

- Added validated Redis URL/key-prefix and object endpoint/bucket/credential
  configuration. URLs and credentials that may contain secrets are redacted.
- Added `github.com/redis/go-redis/v9` and `github.com/minio/minio-go/v7` as the
  infrastructure client implementations.
- No HTTP API, event schema, database migration, domain cache, or document
  contract changed.

## Verification

- Unit tests cover adapter configuration, startup guards, mandatory Redis TTLs,
  and opaque object keys.
- `make adapters-test` starts isolated Redis and SeaweedFS instances under a
  unique Compose project, key prefix, bucket, and volumes.
- The integration test proves Redis health and round-trip deletion, then creates
  a SeaweedFS bucket and proves object upload, metadata, download, deletion, and
  provider-neutral not-found behavior.
- Race-enabled backend tests, `go vet`, formatting, Compose validation, and
  `git diff --check` pass.

## Remaining risks and next task

The adapters establish connectivity and safe primitives only. They deliberately
do not define business cache invalidation, document metadata, content security,
tenant authorization, or retention. SeaweedFS is still a single-node local
service, not a production durability design.

Next: **P01-06 — Next.js runtime shell**.
