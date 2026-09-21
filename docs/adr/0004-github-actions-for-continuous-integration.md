# ADR 0004: GitHub Actions for continuous integration

- Status: Accepted
- Date: 2026-09-21

## Context

The repository is hosted on GitHub and Phase 00 requires a production-grade
quality pipeline without introducing production deployment infrastructure.

## Decision

Use GitHub Actions for pull-request and main-branch continuous integration.
The workflow calls the same stable root commands that developers run locally
and will gate:

- formatting, static analysis, unit tests, and builds;
- frontend formatting, linting, type checking, tests, and production build;
- migration validation and contract generation drift;
- integration and critical end-to-end suites against pinned service images;
- secret, dependency, and source scanning;
- backend and frontend container builds.

Third-party actions are pinned to immutable commit SHAs. Caches are keyed only
from reproducible dependency inputs. Failure artifacts must be useful but must
not contain secrets or sensitive payloads.

CI will not contain cloud credentials, production deployment jobs, Terraform,
Kubernetes, domain, load-balancer, or TLS provisioning.

## Alternatives considered

- GitLab CI and other hosted systems can satisfy the requirements, but would
  add a second control plane to a GitHub-hosted repository.
- CI commands embedded only in workflow YAML tend to drift from local
  development and are harder to reproduce.

## Consequences

Repository commands become the stable interface and GitHub Actions becomes a
thin orchestrator. A future CI-provider change should not require rewriting
the underlying checks. Continuous deployment remains a separate, explicitly
out-of-scope decision.
