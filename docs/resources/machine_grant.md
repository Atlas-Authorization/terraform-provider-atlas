# atlas_machine_grant

Authorize a machine for a resource server and its scopes. 409 if it already has a grant there.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `POST /v1/machines/{id}/grants` (create); `GET /v1/machines/{id}/grants` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/machines/{id}/grants/{grantId}` (destroy).

## Example Usage

```hcl
resource "atlas_machine_grant" "example" {
  machine_id = "tf-machine-grant"
}
```

## Schema

### Required

- `machine_id` (String) The machine id this object belongs to. Changing it forces replacement.

### Optional

- `resource_server_id` (String) The `resource_server_id` field.
- `scopes` (List of String) The `scopes` field.
- `scope_presets` (List of String) The `scope_presets` field. Not returned by the API, so it is not detected as drift.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<machine_id>/<id>`.

```sh
terraform import atlas_machine_grant.example <machine_id>/<id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
