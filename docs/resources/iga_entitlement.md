# atlas_iga_entitlement

Add an entitlement to the catalog: a role, organization membership, group, resource role, installed-app assignment or custom item, with an owner, a risk level and approval rules. The target must exist. 409 for a key already in use.

Scopes: `iga:read` to read, `iga:write` to write.

Backend API: `POST /v1/iga/entitlements` (create); `GET /v1/iga/entitlements/{id}` (read); `PATCH /v1/iga/entitlements/{id}` (update); `DELETE /v1/iga/entitlements/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_iga_entitlement" "example" {
  key = "tf-iga-entitlement"
  name = "tf-iga-entitlement"
  kind = "role"
  target = jsonencode({})
}
```

## Schema

### Required

- `key` (String) The `key` field.
- `name` (String) The `name` field.
- `kind` (String) The `kind` field. One of: `role`, `org_membership`, `group`, `resource_role`, `app_assignment`, `custom`.
- `target` (String, JSON) The `target` field. A JSON document (use `jsonencode`).

### Optional

- `description` (String) The `description` field.
- `owner_user_id` (String) The `owner_user_id` field.
- `risk_level` (String) The `risk_level` field. One of: `low`, `medium`, `high`, `critical`.
- `approval_rules` (String, JSON) The `approval_rules` field. A JSON document (use `jsonencode`).
- `max_duration_seconds` (Number) The `max_duration_seconds` field.
- `requestable` (Boolean) The `requestable` field.
- `require_justification` (Boolean) The `require_justification` field.
- `activation_rules` (String, JSON) The `activation_rules` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_iga_entitlement.example <id>
```
