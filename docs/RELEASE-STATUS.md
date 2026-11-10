# Omisha OS release status — development, not production

## Implemented prototype components
- Go Lite Core: health/readiness, bearer-token protected module catalog, persistent single-instance company registry.
- CRM MVP: per-company contact CRUD subset (create/update/list), company module gate and admin-only API, with unit/HTTP tests.
- Finance MVP: balanced double-entry journals, tenant-scoped trial balance, GST arithmetic estimates, invoice calculation/list APIs and automated tests. No legal GST compliance claims.
- Offline Ed25519 license signature/expiry verification and entitlement checks on protected endpoints.
- Startup fails closed without license files unless `OMISHA_DEV_MODE=true` is explicitly set.
- Python licensing authority with issuance, activation tracking, and revocation.
- Python on-prem deployment plan and Docker Compose generation.
- Optional PostgreSQL, Valkey and NATS infrastructure templates, plus PostgreSQL backup/restore scripts.

## Not completed
- Shared PostgreSQL-backed multi-tenant Core, transactional migrations, full RBAC/SSO, UI, billing/payment processing.
- Customer deployment agent authentication and signed remote jobs.
- Online license revocation sync in Go Core; offline checks alone do not enforce activation counts.
- Full CRM (pipeline, UI, customer RBAC); statutory accounting and GST features (chart of accounts, financial statements, GSTN filing, reconciliations, audit trail); manufacturing, carbon, IoT and AI business modules.
- Production observability, encrypted backups, penetration tests, capacity benchmarks, tested disaster recovery.

## Local development
```bash
cd lite-core
export OMISHA_LITE_TOKEN="$(openssl rand -hex 32)"
export OMISHA_DEV_MODE=true
docker compose up -d --build
```
**Never set development mode for commercial customers.** For licensed use, provide a trusted pinned public key and signed license in the read-only `./licenses` directory, and set `OMISHA_LICENSE_PATH=/licenses/license.json` and `OMISHA_LICENSE_PUBLIC_KEY_PATH=/licenses/public-key.txt`.

## Quality gate
CI runs Go unit/integration tests, Go race detection, vet, Docker builds, Python syntax validation, and Compose parsing. CI success is not a substitute for end-to-end and security testing.
