# atlas_instance_write_budget

Set the write-budget mode and any limit. A limit is an integer from 1 to 10,000,000 per minute; null returns one to its default. Unknown fields are refused.

Scopes: `instance:read` to read, `instance:write` to write.

Backend API: `PUT /v1/instance/write_budget` (create and update); `GET /v1/instance/write_budget` (read); `DELETE /v1/instance/write_budget` (destroy).

## Example Usage

```hcl
resource "atlas_instance_write_budget" "example" {

}
```

## Schema

### Optional

- `mode` (String) The `mode` field. One of: `observe`, `enforce`.
- `writes_per_minute` (Number) The `writes_per_minute` field. Not returned by the API, so it is not detected as drift.
- `counter_writes_per_minute` (Number) The `counter_writes_per_minute` field. Not returned by the API, so it is not detected as drift.
- `events_per_minute` (Number) The `events_per_minute` field. Not returned by the API, so it is not detected as drift.
- `webhook_fanout_per_minute` (Number) The `webhook_fanout_per_minute` field. Not returned by the API, so it is not detected as drift.

### Read-Only

- `id` (String) Atlas object id.
- `window_seconds` (Number) The `window_seconds` field.
- `budgets` (String, JSON) The `budgets` field. A JSON document (use `jsonencode`).
- `exceeded_in_window` (List of String) The `exceeded_in_window` field. One of: `writes`, `counter_writes`, `events`, `webhook_fanout`.
- `resets_in_seconds` (Number) The `resets_in_seconds` field.
- `last_exceeded_at` (Number) The `last_exceeded_at` field. (epoch ms)
- `last_exceeded_budget` (String) The `last_exceeded_budget` field.
- `exceeded_total` (Number) The `exceeded_total` field.
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_instance_write_budget.example instance_write_budget
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
