# atlas_localization

Create or replace a localization override for one resource + locale. Validated on save.

Scopes: `localizations:read` to read, `localizations:write` to write.

Backend API: `POST /v1/localizations` (create); `GET /v1/localizations/{id}` (read); `PATCH /v1/localizations/{id}` (update); `DELETE /v1/localizations/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_localization" "example" {

}
```

## Schema

### Optional

- `resource_type` (String) The `resource_type` field. Changing it forces replacement.
- `resource_name` (String) The `resource_name` field. Changing it forces replacement.
- `locale` (String) The `locale` field. Changing it forces replacement.
- `content` (String, JSON) The `content` field. A JSON document (use `jsonencode`).
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (String) The `created_at` field.
- `updated_at` (String) The `updated_at` field.

## Import

The import id is `<id>`.

```sh
terraform import atlas_localization.example <id>
```
