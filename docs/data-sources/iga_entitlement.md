# atlas_iga_entitlement (data source)

An IGA entitlement, looked up by id.

Every attribute except `id` is computed. Scope: `iga:read`.

## Example Usage

```hcl
data "atlas_iga_entitlement" "example" {
  id = "<id>"
}
```

## Schema

### Required

- `id` (String) The id to look up.

### Read-Only

- `key` (String) The `key` field.
- `name` (String) The `name` field.
- `description` (String) The `description` field.
- `kind` (String) The `kind` field. One of: `role`, `org_membership`, `group`, `resource_role`, `app_assignment`, `custom`.
- `target` (String, JSON) What the entitlement points at, by kind: role {role_key}, org_membership {organization_id, role}, group {group}, resource_role {resource_id, role}, app_assignment {app_id}, custom (free-form). A JSON document (use `jsonencode`).
- `owner_user_id` (String) The `owner_user_id` field.
- `risk_level` (String) The `risk_level` field. One of: `low`, `medium`, `high`, `critical`.
- `approval_rules` (String, JSON) The `approval_rules` field. A JSON document (use `jsonencode`).
- `max_duration_seconds` (Number) The `max_duration_seconds` field.
- `requestable` (Boolean) The `requestable` field.
- `require_justification` (Boolean) The `require_justification` field.
- `activation_rules` (String, JSON) The `activation_rules` field. A JSON document (use `jsonencode`).
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).
