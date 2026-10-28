# atlas_group_rule

Create a rule: when a user matches condition, run the actions (add to a group, grant an app role, set a trait). → 201.

Scopes: `users:read` to read, `users:write` to write.

Backend API: `POST /v1/group_rules` (create); `GET /v1/group_rules/{id}` (read); `PATCH /v1/group_rules/{id}` (update); `DELETE /v1/group_rules/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_group_rule" "example" {
  name = "tf-group-rule"
  condition = jsonencode({})
  actions = jsonencode({})
}
```

## Schema

### Required

- `name` (String) The `name` field.
- `condition` (String, JSON) A user-search filter tree (see POST /v1/users/search). A JSON document (use `jsonencode`).
- `actions` (String, JSON) The `actions` field. A JSON document (use `jsonencode`).

### Optional

- `enabled` (Boolean) The `enabled` field.
- `priority` (Number) Lower runs first (default 100).

### Read-Only

- `id` (String) Atlas object id.
- `created_by` (String) The `created_by` field.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_group_rule.example <id>
```
