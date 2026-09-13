# atlas_oauth_client (Data Source)

Look up an existing OAuth client by id. The client secret is never returned.

Backend API (BAPI): `GET /v1/oauth_clients/:id`.

## Example Usage

```hcl
data "atlas_oauth_client" "dashboard" {
  id = "oac_123"
}

output "dashboard_client_id" {
  value = data.atlas_oauth_client.dashboard.client_id
}
```

## Schema

### Required

- `id` (String) — Atlas object id to look up.

### Read-Only

- `client_id` (String) — The public `client_id`.
- `name` (String) — Client name.
- `redirect_uris` (List of String) — Allowed redirect URIs.
- `allowed_scopes` (List of String) — Allowed scopes.
- `grant_types` (List of String) — Allowed grant types.
- `token_endpoint_auth_method` (String) — Token endpoint auth method.
- `logo_uri` (String) — Logo URL.
- `first_party` (Boolean) — Whether the client is first-party.
- `secret_prefix` (String) — Non-secret prefix of the live secret.
- `is_public` (Boolean) — Whether the client is public (PKCE).
- `created_at` (Number) — Creation time (epoch ms).
- `updated_at` (Number) — Last update time (epoch ms).
