# atlas_abac_policy

Create an ABAC policy from `source` written in the Cedar-subset policy language (permit/forbid with when/unless over principal, action, resource and context attributes; `fga.check(...)` consults the policy’s FGA store). Source that does not parse is refused with the character position; with an entity schema set it is also type-checked and every problem is reported. Optional `tests` are saved with version 1.

Scopes: `abac:read` to read, `abac:write` to write.

Backend API: `POST /v1/abac_policies` (create); `GET /v1/abac_policies/{id}` (read); `PATCH /v1/abac_policies/{id}` (update); `DELETE /v1/abac_policies/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_abac_policy" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `description` (String) The `description` field.
- `enabled` (Boolean) The `enabled` field.
- `fga_store_id` (String) The `fga_store_id` field.
- `source` (String) The `source` field.
- `tests` (String, JSON) The `tests` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `current_version` (Number) The `current_version` field.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_abac_policy.example <id>
```
