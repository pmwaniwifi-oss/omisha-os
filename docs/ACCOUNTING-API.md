# Omisha OS — Accounting API MVP

The Go Lite Core now exposes **administrator-only**, company-scoped accounting endpoints:
- `POST /v1/companies/{company}/journals` — create a balanced journal; duplicate IDs rejected
- `GET /v1/companies/{company}/journals` — list journal entries
- `GET /v1/companies/{company}/trial-balance` — calculate debit/credit net balances by ledger

Example journal:
```json
{"id":"sale-one","date":"2026-11-10","narration":"Sale example","lines":[{"ledger":"cash","debit_paise":11800,"credit_paise":0},{"ledger":"sales","debit_paise":0,"credit_paise":10000},{"ledger":"output-igst","debit_paise":0,"credit_paise":1800}]}
```
Amounts are **integer paise**, not rupees. A registered company with the finance module is required. When signed licensing is configured, finance entitlement is also required.

This is an early internal accounting API, **not statutory financial statements or GST returns**. There is no posting period lock, ledger master, vouchers, invoice numbering, input tax credit eligibility, reconciliation, role-based approvals, reversal workflow, immutable audit log, PostgreSQL transaction store or GSTN filing integration. The single-process JSON store is not suitable for production financial records.
