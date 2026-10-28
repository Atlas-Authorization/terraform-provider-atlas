# atlas_audit_retention

Set audit log retention in days (`null` returns to the default). Entries older than the retention are deleted by an hourly sweep, so shortening it needs `confirm_deletion: true`.

Scopes: `audit:read` to read, `audit:write` to write.

Backend API: `PUT /v1/audit_retention` (create and update); `GET /v1/audit_retention` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_audit_retention" "example" {
  retention_days = 1
}
```

## Schema

### Required

- `retention_days` (Number) The `retention_days` field.

### Optional

- `confirm_deletion` (Boolean) The `confirm_deletion` field. Not returned by the API, so it is not detected as drift.

### Read-Only

- `id` (String) Atlas object id.
- `configured` (Boolean) The `configured` field.
- `default_days` (Number) The `default_days` field.
- `min_days` (Number) The `min_days` field.
- `max_days` (Number) The `max_days` field.

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_audit_retention.example audit_retention
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
