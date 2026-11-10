# Omisha OS — Indian GST invoice MVP (development only)

Added company-scoped **draft calculation and persistence** for tax invoices, with item-level taxable values, manually supplied HSN and GST rates, and arithmetic CGST/SGST/IGST estimates. Monetary values are **integer paise**; rate is in basis points of a percentage (1800 = 18%).

The caller must provide `supply_type` as `intra` or `inter`. The software does not infer place of supply, validate GSTIN, check official HSN tax rates, determine reverse charge, exemptions, cess, ITC, e-invoice thresholds, or eligibility for GST returns.

The endpoint `POST /v1/companies/{company}/invoices` creates a stored invoice calculation; `GET /v1/companies/{company}/invoices` lists it. These records are **not legally validated GST tax invoices**. They lack statutory numbering, issuer identity, prescribed invoice fields, credit/debit notes, cancellation, signature and GSTN/e-invoice integration.

**No automated journal posting** is performed; invoices and journal records are separate. This avoids silently producing inconsistent accounts but requires reconciliation before real use. The storage is single-process JSON and not production-grade.

Example:
```json
{"id":"invoice-one","date":"2026-11-10","party_name":"Buyer","supply_type":"intra","items":[{"description":"Goods","hsn":"1234","taxable_paise":10000,"rate_basis_points":1800}]}
```
