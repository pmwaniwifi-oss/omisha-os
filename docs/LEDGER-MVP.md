# Ledger master — Omisha OS development MVP

New administrator-only endpoints:
- `POST /v1/companies/{company}/ledgers` — create ledger with `id`, `name`, `group`
- `GET /v1/companies/{company}/ledgers` — list company ledgers

Supported groups: `asset`, `liability`, `equity`, `income`, `expense`.

Example: `{"id":"cash","name":"Cash in Hand","group":"asset"}`.

Journal posting checks that all referenced ledger IDs exist in the same company. Create ledger masters before posting entries. Ledger groups do not yet drive statutory P&L/Balance Sheet, and journal/ledger storage is separate JSON files, **not one atomic database transaction**. No per-user RBAC, fiscal period locks, or immutable audit log. Do not use for production accounting.
