# Omisha OS — Indian Accounting + GST development module

Implemented backend primitives:
- Tenant-keyed, immutable-once-posted double-entry journals, stored in paise (integer amounts).
- Debit/credit balancing validation, duplicate journal ID rejection, basic ledger identifier validation.
- GST **arithmetic estimate only**: intrastate CGST+SGST or interstate IGST, supplied tax rate, half-up rounding to paise.

**Important:** GST rate, place-of-supply, exemption, reverse charge, composition, cess, HSN/SAC, ITC eligibility, e-invoice, e-way bill and return eligibility are not determined by this engine. The caller must not rely on this as a tax-compliance decision. It does not generate or file GSTR-1, GSTR-3B, GSTR-9 or GSTN submissions.

Planned: Indian chart of accounts, fiscal year, GSTIN validation, invoice sequencing, B2B/B2C invoice templates, purchase register, debit/credit notes, ledger and trial balance, P&L and balance sheet, TDS/TCS, bank reconciliation, audit logs, PostgreSQL transactions, per-company RBAC, tax-rule versioning, GSTR reconciliation and official GSTN integration.

Current functions are internal Go primitives with tests, **not exposed as HTTP endpoints** pending authorization and transaction controls. The JSON store is single-process only and is not suitable for regulated production bookkeeping.
