# atlas_feature_flag

Create a feature flag: a typed value (`boolean`, `string`, `int` or `json`) with a `default_value`, an `off_value` (served while `enabled` is false; defaults to false for a boolean, else the default) and ordered targeting `rules` by org, user, plan and percentage. 409 when the key exists.

Scopes: `flags:read` to read, `flags:write` to write.

Backend API: `POST /v1/feature_flags` (create); `GET /v1/feature_flags/{key}` (read); `PATCH /v1/feature_flags/{key}` (update); `DELETE /v1/feature_flags/{key}` (destroy).

## Example Usage

```hcl
resource "atlas_feature_flag" "example" {
  key = "tf-feature-flag"
  type = "boolean"
  default_value = true
}
```

## Schema

### Required

- `key` (String) Lowercase letters, digits, - and _. Changing it forces replacement.
- `type` (String) The `type` field. One of: `boolean`, `string`, `int`, `json`. Changing it forces replacement.
- `default_value` (Boolean) A boolean, string, integer or JSON value matching the flag `type`.

### Optional

- `description` (String) The `description` field.
- `enabled` (Boolean) The `enabled` field.
- `off_value` (Boolean) A boolean, string, integer or JSON value matching the flag `type`.
- `rules` (String, JSON) Ordered targeting rules; the first match wins. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) When the flag was created (epoch ms).
- `updated_at` (Number) When the flag last changed (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_feature_flag.example <id>
```
