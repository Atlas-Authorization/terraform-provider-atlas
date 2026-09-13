# atlas_resource_server

An API (resource server / audience) that accepts machine tokens from the OAuth2
`client_credentials` grant. Defines an audience identifier and the scopes it
grants.

Backend API (BAPI): `POST /v1/resource_servers` (create),
`GET /v1/resource_servers/:id` (read), `PATCH /v1/resource_servers/:id` (update),
`DELETE /v1/resource_servers/:id` (delete).

## Example Usage

```hcl
resource "atlas_resource_server" "billing_api" {
  identifier        = "https://api.example.com/billing"
  name              = "Billing API"
  token_ttl_seconds = 3600

  scopes {
    value       = "read:invoices"
    description = "Read invoices"
  }
  scopes {
    value = "write:invoices"
  }
}
```

## Schema

### Required

- `identifier` (String) — The audience (`aud`) stamped into issued tokens. Immutable — changing it forces replacement.
- `name` (String) — Human-readable name.

### Optional

- `scopes` (Block List) — Scopes this API defines. Each block has `value` (String, required) and `description` (String, optional).
- `token_ttl_seconds` (Number) — Lifetime of machine tokens for this API, in seconds.
- `signing_alg` (String) — JWT signing algorithm for issued tokens (e.g. `RS256`).

### Read-Only

- `id` (String) — Atlas object id.
- `created_at` (Number) — Creation time (epoch milliseconds).
- `updated_at` (Number) — Last update time (epoch milliseconds).

## Import

```sh
terraform import atlas_resource_server.billing_api rs_123
```
