# Omisha OS 5.2 — Optional Scale Infrastructure

The `scale/` stack provides a private PostgreSQL 16 instance, optional Valkey cache and optional NATS JetStream event broker.

**Important:** These services are infrastructure building blocks, **not yet wired into Lite Core's company registry**. Lite Core continues to use its single-process JSON registry. Do not claim this is a completed PostgreSQL migration or production cluster.

## Setup
```bash
cd scale
bash init.sh
docker compose up -d
docker compose --profile cache --profile events up -d
```

PostgreSQL uses a file-based Docker secret; no database port is exposed to the host. Valkey and NATS are opt-in to save memory. Persistent volumes retain data on restart. Database credentials must be stored securely and backed up separately.

## Production requirements
Add version-pinned image digests, tested database backups/restores, TLS or trusted private overlay networking, connection pooling, metrics, migrations, high availability, NATS authentication, resource benchmarks, upgrade procedures and failure-injection tests before commercial use. PostgreSQL/Valkey/NATS each have their own resource requirements and do not guarantee unlimited scale.

## Roadmap
1. Implement a PostgreSQL-backed repository interface in Lite Core and transactional schema migrations.
2. Add license-entitlement binding and audit records.
3. Implement authenticated agent pairing, signed deployment jobs, result reporting, and rollback.
4. Run end-to-end integration and load tests in CI, then VMware tests when authorized.
