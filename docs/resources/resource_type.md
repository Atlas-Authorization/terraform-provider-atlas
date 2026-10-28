# atlas_resource_type

Define a resource type: a `key`, its `roles` (each a set of permissions, optionally `includes` other roles), an optional `parent_type` resources nest under, and `inherit` (default true) controlling whether assignments on a parent flow down.

Scopes: `resources:read` to read, `resources:write` to write.

Backend API: `POST /v1/resource_types` (create); `GET /v1/resource_types/{key}` (read); `PATCH /v1/resource_types/{key}` (update); `DELETE /v1/resource_types/{key}` (destroy).

## Example Usage

```hcl
resource "atlas_resource_type" "example" {

}
```

## Schema

### Optional

- `key` (String) The `key` field. Changing it forces replacement.
- `name` (String) The `name` field.
- `description` (String) The `description` field.
- `parent_type` (String) The `parent_type` field.
- `inherit` (Boolean) The `inherit` field.
- `roles` (String, JSON) The `roles` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_resource_type.example <id>
```
