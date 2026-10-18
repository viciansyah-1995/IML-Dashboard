# Decisions

- Existing project is not a Git checkout. Keep old prototype sources intact and additionally snapshot them under docs/legacy-prototype.
- PostgreSQL 17.10 native is installed; Docker/Compose and OpenClaw are absent. Use a dedicated project cluster on 127.0.0.1:55432 (never touch an existing user database), with DATABASE_URL override and PostgreSQL 17.10 Compose alternative.
- Go is installed only into ignored .local/toolchains, with official release checksum verification; no system installation.
- Money uses NUMERIC(24,4) in PostgreSQL and decimal strings with shopspring/decimal in Go/JSON. Counts are integers, durations seconds. No floating point monetary calculations.
- Definition contracts are per channel and immutable per version. Synthetic contracts are explicitly demo-only. Platform canonical mappings remain unconfirmed; source dashboards can publish valid source metrics without forcing unconfirmed canonical mappings.
- Global host DTOs exclude brand, account, private session/slot IDs, ID Line, source timestamps and evidence. Global filters reject brand_id. Private and global state/query keys are independent.
- Native run/worker; frontend /api proxy. Production integration credentials stay server-side. Mock extraction requires DEMO_MODE and explicit fixture scenario; no automatic fallback from OpenClaw.
- SapaGo SDK/auth specification is not supplied: report unavailable, expose deterministic insight only. No guessed endpoints.
