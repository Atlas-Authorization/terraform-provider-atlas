# atlas_radius_client

Register a RADIUS NAS client (name, NAS-Identifier, optional source-IP bind, write-only shared secret). The NAS-Identifier is unique across the deployment; a duplicate is a 409.

Scopes: `radius_clients:read` to read, `radius_clients:write` to write.

Backend API: `POST /v1/radius_clients` (create); `GET /v1/radius_clients/{id}` (read); `PATCH /v1/radius_clients/{id}` (update); `DELETE /v1/radius_clients/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_radius_client" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `nas_identifier` (String) The `nas_identifier` field.
- `ip_address` (String) The `ip_address` field.
- `ip_cidr` (String) The `ip_cidr` field.
- `reply_attributes` (String, JSON) The `reply_attributes` field. A JSON document (use `jsonencode`).
- `shared_secret` (String, Sensitive) The `shared_secret` field. Write-only: sent to Atlas, never returned, and not detected as drift.
- `require_message_authenticator` (Boolean) The `require_message_authenticator` field.
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `has_secret` (Boolean) The `has_secret` field.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_radius_client.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
