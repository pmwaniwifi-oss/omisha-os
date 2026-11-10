# Omisha OS 5.1 — Lightweight persistent company registry

The Lite Core now supports `GET /v1/companies` and `PUT /v1/companies` with a bearer token. A JSON registry is atomically replaced on each successful write, and persisted in a Docker volume. Each company has an ID, name and selected module identifiers; core is mandatory.

Example:
```bash
curl -X PUT http://127.0.0.1:18081/v1/companies \
 -H "Authorization: Bearer $OMISHA_LITE_TOKEN" \
 -H 'Content-Type: application/json' \
 -d '{"id":"demo-company","name":"Demo Company","modules":["core","crm"]}'
curl -H "Authorization: Bearer $OMISHA_LITE_TOKEN" http://127.0.0.1:18081/v1/companies
```

**Limitations:** This is a single-instance local development registry, not a PostgreSQL implementation. It does not support multiple writer processes, user RBAC, tenant-specific authentication, licensing enforcement, real module execution or multi-node clustering. Keep it on a single instance. Future Scale Mode will use PostgreSQL transactions and migration tooling.
