# Omisha OS licensing and on-premises installation (prototype)

## Security and scope
This is a development prototype, **not** production-grade billing, revocation, or secure offline enforcement. Never publish admin tokens, private keys or customer data. The server signing key is persisted in its Docker volume; back it up securely. Offline licenses can be checked for signature and expiry, but cannot learn of revocation until online. Customer software is not tamper-proof. The local runtime currently serves health endpoints only.

## Run licensing service on Ubuntu VM
```bash
cd omisha-os/licensing
export OMISHA_LICENSE_ADMIN_TOKEN="$(openssl rand -hex 32)"
docker compose up -d --build
curl http://127.0.0.1:8090/health
curl http://127.0.0.1:8090/v1/public-key
```
Store the admin token securely; do not paste it into shell history on a shared machine.

## Issue a signed license
```bash
curl -sS -X POST http://127.0.0.1:8090/v1/licenses \
 -H "X-Admin-Token: $OMISHA_LICENSE_ADMIN_TOKEN" \
 -H 'Content-Type: application/json' \
 -d '{"company":"Demo Manufacturing","modules":["core","crm","finance"],"expires_at":"2027-12-31T23:59:59Z","max_installations":2}' > issued.json
python3 -c 'import json; d=json.load(open("issued.json"));json.dump(d["license"],open("customer-license.json","w"));open("public-key.txt","w").write(d["public_key"])'
```

## Offline verification and customer-premises deployment
Copy customer-license.json and public-key.txt through a secure channel. Pin the public key obtained from the trusted licensing operator, not from the customer-supplied license file.
```bash
cd omisha-os
bash onprem/install.sh licensing/customer-license.json licensing/public-key.txt demo-001 core,crm,finance
curl http://127.0.0.1:18080/health
```
The installer checks Docker and Compose, verifies Ed25519 signature and expiry locally, renders a per-customer Compose stack and deploys it. It does not install Docker automatically and does not use remote shell execution. On-premises prototype uses a dedicated PostgreSQL container; hybrid shared database is intended for managed hosting and is not implemented by this installer.

## Online activation
```bash
curl -X POST http://127.0.0.1:8090/v1/activate -H 'Content-Type: application/json' -d '{"license":{...},"installation_id":"vm-demo-001"}'
```
Send the complete signed license object in place of {...}. Online activation records an installation and checks revocation and limits. Offline installation checks signature and expiry but **does not register an activation**; therefore activation quotas are not enforceable offline. Add signed activation certificates and periodic reconciliation before commercial release.

## Important limitations
No payment gateway, automatic renewals, secure update channel, true offline activation limits, enterprise RBAC, tenant billing, hardening or end-to-end VM tests yet. The existing Omisha 3.1 control-plane source must be added separately to GitHub. Modules currently have health endpoints, not production ERP logic.
