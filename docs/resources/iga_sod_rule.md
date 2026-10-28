# atlas_iga_sod_rule

Create a rule. `mode` is `block` (refuse) or `require_approval` (an extra approval step on a request; `approver` names who, default the owner).

Scopes: `iga:read` to read, `iga:write` to write.

Backend API: `POST /v1/iga/sod_rules` (create); `GET /v1/iga/sod_rules` (read: the list is scanned for the id); `PATCH /v1/iga/sod_rules/{id}` (update); `DELETE /v1/iga/sod_rules/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_iga_sod_rule" "example" {
  name = "tf-iga-sod-rule"
  entitlement_a = "tf-iga-sod-rule"
  entitlement_b = "tf-iga-sod-rule"
}
```

## Schema

### Required

- `name` (String) The `name` field.
- `entitlement_a` (String) The `entitlement_a` field. Changing it forces replacement.
- `entitlement_b` (String) The `entitlement_b` field. Changing it forces replacement.

### Optional

- `description` (String) The `description` field.
- `mode` (String) The `mode` field. One of: `block`, `require_approval`.
- `approver` (String, JSON) The `approver` field. A JSON document (use `jsonencode`).
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_iga_sod_rule.example <id>
```
