# atlas_iga_birthright_rule

Create a birthright rule: everyone matching `match` (a group, organization and role, app role or verified email domain; all conditions must hold) holds the listed entitlements. When someone stops matching, the grants the rule gave are revoked unless `revoke_on_unmatch` is false.

Scopes: `iga:read` to read, `iga:write` to write.

Backend API: `POST /v1/iga/birthright_rules` (create); `GET /v1/iga/birthright_rules` (read: the list is scanned for the id); `PATCH /v1/iga/birthright_rules/{id}` (update); `DELETE /v1/iga/birthright_rules/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_iga_birthright_rule" "example" {
  name = "tf-iga-birthright-rule"
  match = jsonencode({})
  entitlement_ids = ["example"]
}
```

## Schema

### Required

- `name` (String) The `name` field.
- `match` (String, JSON) The `match` field. A JSON document (use `jsonencode`).
- `entitlement_ids` (List of String) The `entitlement_ids` field.

### Optional

- `revoke_on_unmatch` (Boolean) The `revoke_on_unmatch` field.
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_iga_birthright_rule.example <id>
```
