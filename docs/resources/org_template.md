# atlas_org_template

Create an organization template: custom roles (with existing permissions), policy, role set and default role, seat limit and warning level, and metadata. 409 when the key is taken. Honours Idempotency-Key.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `POST /v1/org_templates` (create); `GET /v1/org_templates/{id}` (read); `PATCH /v1/org_templates/{id}` (update); `DELETE /v1/org_templates/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_org_template" "example" {
  key = "tf-org-template"
  name = "tf-org-template"
}
```

## Schema

### Required

- `key` (String) Unique per instance. Fixed at create.
- `name` (String) The `name` field.

### Optional

- `description` (String) The `description` field.
- `roles` (String, JSON) The `roles` field. A JSON document (use `jsonencode`).
- `policy` (String, JSON) An organization policy, as PATCH /v1/organizations/{id}/policy takes it. A JSON document (use `jsonencode`).
- `role_set` (String) A role set key or id.
- `default_role` (String) One of the template's roles, or a role in its role set.
- `max_allowed_memberships` (Number) The `max_allowed_memberships` field.
- `seat_warn_percent` (Number) The `seat_warn_percent` field.
- `public_metadata` (String, JSON) The `public_metadata` field. A JSON document (use `jsonencode`).
- `private_metadata` (String, JSON) Backend only. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_org_template.example <id>
```
