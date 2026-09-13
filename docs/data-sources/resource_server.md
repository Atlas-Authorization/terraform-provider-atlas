# atlas_resource_server (Data Source)

Look up an existing resource server (API) by id.

Backend API (BAPI): `GET /v1/resource_servers/:id`.

## Example Usage

```hcl
data "atlas_resource_server" "billing" {
  id = "rs_123"
}

output "billing_audience" {
  value = data.atlas_resource_server.billing.identifier
}
```

## Schema

### Required

- `id` (String) — Atlas object id to look up.

### Read-Only

- `identifier` (String) — The audience identifier.
- `name` (String) — Human-readable name.
- `scopes` (List of Object) — Scopes this API defines; each has `value` and `description`.
- `token_ttl_seconds` (Number) — Machine-token lifetime (seconds).
- `signing_alg` (String) — JWT signing algorithm.
- `created_at` (Number) — Creation time (epoch ms).
- `updated_at` (Number) — Last update time (epoch ms).
