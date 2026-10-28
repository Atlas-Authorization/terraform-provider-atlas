# atlas_app_role

Create an app-level user role: a role not tied to any organization, granting permission keys. Users holding it carry app_roles / app_permissions on their session token. Emits appRole.created.

Scopes: `users:read` to read, `users:write` to write.

Backend API: `POST /v1/app_roles` (create); `GET /v1/app_roles/{id}` (read); `PATCH /v1/app_roles/{id}` (update); `DELETE /v1/app_roles/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_app_role" "example" {
  key = "tf-app-role"
  name = "tf-app-role"
}
```

## Schema

### Required

- `key` (String) Unique per instance; cannot change later. Changing it forces replacement.
- `name` (String) The `name` field.

### Optional

- `description` (String) The `description` field.
- `permissions` (List of String) resource:action keys. The org: namespace is reserved for organization permissions.

### Read-Only

- `id` (String) Atlas object id.
- `users_count` (Number) How many users hold the role (list and create only).
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_app_role.example <id>
```
