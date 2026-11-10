# Omisha OS — CRM MVP (development)

A minimal persistent contact directory is implemented in the Go Lite Core. Contacts are keyed by company ID and contact ID; list operations filter by company. Access requires the global administrator token, a registered company with the `crm` module, and an active signed license with CRM entitlement when licensing is enabled.

Endpoints:
- `GET /v1/companies/{company}/contacts`
- `PUT /v1/companies/{company}/contacts`

Example payload:
```json
{"id":"contact-one","name":"Example Contact","email":"example@example.com","phone":"+91 9000000000"}
```

This is **not a complete CRM**: no sales pipeline, contact ownership, per-user RBAC, search indexes, consent tracking, encryption at rest, multi-process locking, PostgreSQL storage or user interface. Global admin can access all companies; this is **not customer-facing tenant authentication**. Only use with trusted administrators in development.
