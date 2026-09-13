# atlas_role

A custom organization role: a stable key, a label, and a set of permission keys.
The key is published into authorization checks and is immutable — changing it
forces replacement.

Backend API (BAPI): `POST /v1/roles` (create), `GET /v1/roles` (list — the read
filters this list by id), `PATCH /v1/roles/:id` (label update),
`PUT /v1/roles/:id/permissions` (permission set), `DELETE /v1/roles/:id` (delete).

## Example Usage

```hcl
resource "atlas_role" "billing_admin" {
  key         = "billing_admin"
  name        = "Billing Admin"
  description = "Manage invoices and subscriptions"
  permissions = ["org:invoices:manage", "org:subscriptions:read"]
}
```

## Schema

### Required

- `key` (String) — Stable role key (e.g. `billing_admin`). Immutable; changing it forces replacement.
- `name` (String) — Human-readable label.

### Optional

- `description` (String) — Optional description.
- `permissions` (Set of String) — Permission keys granted by the role. Keys not defined on the instance are dropped by the API.

### Read-Only

- `id` (String) — Atlas object id.
- `is_system` (Boolean) — True for a built-in system role.
- `created_at` (Number) — Creation time (epoch ms).

## Import

```sh
terraform import atlas_role.billing_admin role_123
```
