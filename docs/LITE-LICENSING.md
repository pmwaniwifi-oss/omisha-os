# Omisha Lite Core — signed offline license validation

The Lite Core verifies the **existing Python licensing server's Ed25519 signed license** locally. It checks signature, expiration, module allowlist, and mandatory core entitlement. This does not contact the central licensing service.

## Configure
Set both variables to enable enforcement:
- `OMISHA_LICENSE_PATH=/licenses/license.json`
- `OMISHA_LICENSE_PUBLIC_KEY_PATH=/licenses/public-key.txt`

Mount the license and pinned public key read-only. The public key must come from a trusted operator channel; do not trust a public key bundled by an untrusted customer.

Without these variables, Lite Core remains in explicit **development mode**, and its module catalog is not license-enforced. Never use that mode for commercial deployments.

## Limitations
Offline validation cannot detect revocation, enforce activation limits, or stop a customer with full control of their host from modifying binaries. Revalidation happens on each `GET /v1/modules` request and on company writes. Company registry module selections must be a subset of the signed entitlements. License expiry stops protected API access, not the HTTP health endpoint. Production requires online activation, offline signed activation certificates, clock-tamper mitigation, and reconciliation.
