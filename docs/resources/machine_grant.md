# atlas_machine_grant

Authorize a machine for a resource server and its scopes. 409 if it already has a grant there.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `POST /v1/machines/{id}/grants` (create); `GET /v1/machines/{id}/grants` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/machines/{id}/grants/{grantId}` (destroy).

## Example Usage

```hcl
resource "atlas_machine_grant" "example" {
  machine_id = "tf-machine-grant"
  resource_server_id = "tf-machine-grant"
}
```

## Schema

### Required

- `machine_id` (String) The machine id this object belongs to. Changing it forces replacement.
- `resource_server_id` (String) The `resource_server_id` field.

### Optional

- `scopes` (List of String) The `scopes` field.
- `scope_presets` (List of String) The `scope_presets` field. Not returned by the API, so it is not detected as drift.
- `constraints` (String, JSON) Limits carried in the machine token's `constraints` claim (a JSON object of at most 2 KB). A JSON document (use `jsonencode`).
- `expires_in_seconds` (Number) The `expires_in_seconds` field. Not returned by the API, so it is not detected as drift.
- `expires_at` (String) Alternative to `expires_in_seconds`, not both. ISO 8601 or epoch ms, as text.

### Read-Only

- `id` (String) Atlas object id.
- `expired` (Boolean) True once `expires_at` has passed. An expired grant mints nothing and is dropped from verify results.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<machine_id>/<id>`.

```sh
terraform import atlas_machine_grant.example <machine_id>/<id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
