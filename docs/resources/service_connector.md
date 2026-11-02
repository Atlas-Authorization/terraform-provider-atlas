# atlas_service_connector

Connect a service. `credentials` are checked against the kind's fields and stored in the vault (never returned again); `config` holds the non-secret settings. 409 when the name is taken.

Scopes: `service_connectors:read` to read, `service_connectors:write` to write.

Backend API: `POST /v1/service_connectors` (create); `GET /v1/service_connectors/{id}` (read); `PATCH /v1/service_connectors/{id}` (update); `DELETE /v1/service_connectors/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_service_connector" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `kind` (String) The `kind` field. Changing it forces replacement.
- `config` (String, JSON) The `config` field. A JSON document (use `jsonencode`).
- `credentials` (String, JSON, Sensitive) The `credentials` field. A JSON document (use `jsonencode`). Write-only: sent to Atlas, never returned, and not detected as drift.

### Read-Only

- `id` (String) Atlas object id.
- `category` (String) The `category` field. One of: `messaging`, `crm`, `analytics`, `fraud`, `idv`, `http`, `ticketing`, `null`.
- `enabled` (Boolean) The `enabled` field.
- `last_used_at` (Number) When an operation last ran (epoch ms).
- `last_status` (String) The `last_status` field. One of: `ok`, `error`, `null`.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_service_connector.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
