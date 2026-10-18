# Implementation plan — 11 October 2026

Preserve the previous prototype in docs/legacy-prototype. Implement the requested modular Go/PostgreSQL service and an API-driven React frontend in frontend/.

A. Config, isolated native PostgreSQL, SQL migrations, sessions/ACL, master data, opt-in seed.
B. Upload metadata/evidence, manual review, publication and revision, transaction-safe active version.
C. Persistent queue, leased worker, manual/mock/OpenClaw adapters and schema validation.
D. Definition-aware, exact decimal summaries, coverage and charts.
E. Allowlisted global host aggregates, compare/profile, CSV/XLSX and regression tests.
F. Deterministic insights, integration status, operations, backup/restore documentation and QA.

Status is recorded in requirements-traceability.md; actual validation results in verification.md. Run the full manual vertical slice before extending later phases. No cloud deployment or automatic publication.
