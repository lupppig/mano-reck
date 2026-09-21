# Generated-file policy

Source schemas and migrations are authored; generated clients and query code
are derived artifacts.

- OpenAPI and event schemas under `contracts/` are source files unless their
  header explicitly marks them as generated.
- Generated Go and TypeScript files must contain a standard generated-file
  header and the command that recreates them.
- Generated source required to build a clean checkout is committed. Build
  output, coverage, reports, dependency directories, and local runtime data are
  not committed.
- Generated files are never edited by hand. Change the source schema, SQL, or
  generator configuration and run `make generate`.
- `make generate-check` regenerates outputs and fails if the working tree
  changes, preventing stale committed artifacts.
- Generator versions are pinned in repository configuration. A generator
  upgrade is reviewed with its output in the same change.

Lockfiles, checksums, and migration files are authored/reproducibility inputs;
they are not disposable build output.
