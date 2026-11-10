# Omisha OS 5.0 staged development

## Stage 1: committed foundation
- Company registry with SQLite persistence
- Module catalog and company entitlements registry
- Deployment plan API (does NOT execute commands remotely)
- Draft invoice registry in INR minor units
- Audit trail and protected admin API
- Existing Ed25519 licensing service and on-prem signed-license installer

## Stage 2: not implemented
- Real per-company RBAC authentication, SSO/OIDC, token rotation, secure sessions
- Bind signed licensing entitlements to company module selection and deployment plans
- Authenticated agent pairing, signed jobs, idempotent execution, status reporting
- Marketplace package metadata, trusted image digests, verified updates, rollback

## Stage 3: not implemented
- Production payment integrations, GST invoicing, renewals and collections
- Monitoring, backup/restore, tenant isolation hardening, disaster recovery
- Functional ERP/CRM/carbon/manufacturing/finance/AI/IoT services

## Development run
```bash
export OMISHA_CONTROL_ADMIN_TOKEN="$(openssl rand -hex 32)"
docker compose -f control-plane/compose.yml up -d --build
curl http://127.0.0.1:8080/health
curl -H "Authorization: Bearer $OMISHA_CONTROL_ADMIN_TOKEN" http://127.0.0.1:8080/v1/catalog
```

Create a company:
```bash
curl -X POST http://127.0.0.1:8080/v1/companies \
 -H "Authorization: Bearer $OMISHA_CONTROL_ADMIN_TOKEN" \
 -H 'Content-Type: application/json' \
 -d '{"id":"demo-company","name":"Demo Company"}'
```

Assign modules:
```bash
curl -X PUT http://127.0.0.1:8080/v1/companies/demo-company/modules \
 -H "Authorization: Bearer $OMISHA_CONTROL_ADMIN_TOKEN" \
 -H 'Content-Type: application/json' \
 -d '{"modules":["core","crm","finance"]}'
```

Create a deployment PLAN:
```bash
curl -X POST http://127.0.0.1:8080/v1/deployments \
 -H "Authorization: Bearer $OMISHA_CONTROL_ADMIN_TOKEN" \
 -H 'Content-Type: application/json' \
 -d '{"company_id":"demo-company","target":"onprem"}'
```

**Security:** This is an internal development API. Do not expose ports to the public Internet. It has a single super-admin bearer token, no production RBAC or TLS termination, no payment capture, and no verified on-prem deployment orchestration. No VMware testing has been performed.
