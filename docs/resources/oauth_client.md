# atlas_oauth_client

A "Sign in with Atlas" relying party (OAuth/OIDC client). A confidential client's
secret is revealed once, on create, and stored (sensitive) in state; a public
(PKCE) client has none.

Backend API (BAPI): `POST /v1/oauth_clients` (create), `GET /v1/oauth_clients/:id`
(read), `PATCH /v1/oauth_clients/:id` (update), `DELETE /v1/oauth_clients/:id`
(delete).

## Example Usage

```hcl
resource "atlas_oauth_client" "dashboard" {
  name          = "Internal dashboard"
  redirect_uris = ["https://dashboard.example.com/callback"]
  # allowed_scopes and grant_types default sensibly when omitted.
}

output "dashboard_secret" {
  value     = atlas_oauth_client.dashboard.client_secret
  sensitive = true
}
```

## Schema

### Required

- `name` (String) — Human-readable name shown on the consent screen.
- `redirect_uris` (List of String) — Allowed redirect URIs. Absolute https (or `http://localhost` for development).

### Optional

- `allowed_scopes` (List of String) — OIDC scopes the client may request. Defaults to `[openid, profile, email]`; `openid` is required.
- `grant_types` (List of String) — Allowed grant types. Defaults to `[authorization_code, refresh_token]`.
- `token_endpoint_auth_method` (String) — One of `client_secret_basic`, `client_secret_post`, or `none` (public/PKCE). Defaults to `client_secret_basic`.
- `logo_uri` (String) — Optional logo URL shown on the consent screen.
- `first_party` (Boolean) — When true the client is first-party and may skip the consent screen. Defaults to `false`.

### Read-Only

- `id` (String) — Atlas object id.
- `client_id` (String) — The public `client_id` used at the authorize/token endpoints.
- `client_secret` (String, Sensitive) — The confidential client secret, revealed only on create. Empty for a public client; a read never returns it.
- `secret_prefix` (String) — Non-secret prefix of the live secret, to recognise which secret is deployed.
- `is_public` (Boolean) — True when the client authenticates with PKCE and holds no secret.
- `created_at` (Number) — Creation time (epoch milliseconds).
- `updated_at` (Number) — Last update time (epoch milliseconds).

## Import

```sh
terraform import atlas_oauth_client.dashboard oac_123
```

The one-time `client_secret` cannot be recovered on import (the API never returns
it); rotate the client out of band if you need a fresh secret in state.
