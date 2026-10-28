# atlas_administrative_unit

Create an administrative unit. Without a `filter` it is static and you list its members; with one it is dynamic and its members are whoever match.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `POST /v1/administrative_units` (create); `GET /v1/administrative_units/{id}` (read); `PATCH /v1/administrative_units/{id}` (update); `DELETE /v1/administrative_units/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_administrative_unit" "example" {
  key = "tf-administrative-unit"
  name = "tf-administrative-unit"
}
```

## Schema

### Required

- `key` (String) The `key` field. Changing it forces replacement.
- `name` (String) The `name` field.

### Optional

- `description` (String) The `description` field.
- `organization_id` (String) The `organization_id` field. Changing it forces replacement.
- `filter` (String, JSON) All conditions given must hold. A unit with a filter is dynamic: its members are whoever match now. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `kind` (String) The `kind` field. One of: `static`, `dynamic`.
- `member_count` (Number) Listed members of a static unit.
- `created_at` (Number) Created, Unix ms.
- `updated_at` (Number) Updated, Unix ms.

## Import

The import id is `<id>`.

```sh
terraform import atlas_administrative_unit.example <id>
```
