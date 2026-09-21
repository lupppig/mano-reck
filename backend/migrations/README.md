# Database migrations

Phase 01 introduces the database runtime and first migration. Migration files
are ordered, paired plain SQL files:

```text
000001_descriptive_name.up.sql
000001_descriptive_name.down.sql
```

Numbers are six digits and never reused. Names use lowercase snake case. Every
`up` file has a matching `down` file with the same number and name. A down
migration documents irreversible behavior instead of pretending data can be
restored.

Run `make migrations-check` before committing. Applying migrations against
PostgreSQL is added with the Phase 01 database foundation; this Phase 00 check
proves naming and pairing without inventing domain tables.
