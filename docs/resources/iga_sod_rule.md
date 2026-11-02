# atlas_iga_sod_rule

A separation-of-duties rule: two entitlements one person should not hold
together. It either blocks the second one, or adds an approval step to the
request. Assignments made outside a request are checked as well.

Backend API (BAPI): `POST /v1/iga/sod_rules`, `PATCH /v1/iga/sod_rules/:id`,
`DELETE /v1/iga/sod_rules/:id` (and the list for reads). The API key needs the
`iga:read` and `iga:write` scopes.

## Example Usage

```hcl
resource "atlas_iga_sod_rule" "approve_vs_pay" {
  name          = "Approving and paying invoices"
  entitlement_a = atlas_iga_entitlement.invoice_approver.id
  entitlement_b = atlas_iga_entitlement.invoice_payer.id
  mode          = "require_approval"
  approver      = jsonencode({ type = "group", ids = ["risk"], mode = "any" })
}
```

## Argument Reference

- `name` (Required) Display name.
- `description` (Optional) Why the two conflict.
- `entitlement_a`, `entitlement_b` (Required, force replacement) The ids of the two entitlements.
- `mode` (Optional) `block` (default) or `require_approval`.
- `approver` (Optional) For `require_approval`: the extra approval step as JSON. Defaults to the entitlement owner.
- `enabled` (Optional) Defaults to `true`.

## Attribute Reference

- `id` The rule id.

## Import

```shell
terraform import atlas_iga_sod_rule.approve_vs_pay igs_…
```
