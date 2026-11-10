# Omisha OS 5.1 Lite Core (development)

This introduces a **small Go runtime** with a read-only module catalog, health/ready endpoints, constant-time bearer token check, HTTP timeouts, graceful shutdown, and a scratch Docker image. It is a foundation, **not a replacement for the existing Python control plane or a functioning ERP**.

## Run
```bash
cd lite-core
export OMISHA_LITE_TOKEN="$(openssl rand -hex 32)"
export OMISHA_MODULES="core,crm,finance"
docker compose up -d --build
curl http://127.0.0.1:18081/health
curl -H "Authorization: Bearer $OMISHA_LITE_TOKEN" http://127.0.0.1:18081/v1/modules
```

## Test without Docker
```bash
cd lite-core
go test ./...
go vet ./...
```

## Deployment modes
- **Lite**: Go core and only required modules. Database, cache, broker are optional until actually needed.
- **Scale**: dedicated PostgreSQL, optional Valkey and NATS JetStream, then multi-node K3s where warranted.

## Honest boundaries
- Module names are metadata, not installed functional modules.
- No measured throughput, RAM benchmarks, SLOs, load tests, or multi-node validation yet.
- The token is an internal development credential, not SSO/RBAC.
- The Compose memory cap of 128 MB is a configured ceiling, not a proven runtime requirement.
- Do not publish the service directly on the internet; add HTTPS reverse proxy, secrets management, and rate limits before production.
