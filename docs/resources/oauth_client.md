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

# Your own app: end the session on RP-initiated logout without the
# "Sign out?" page (first-party clients only).
resource "atlas_oauth_client" "web" {
  name                     = "Acme web app"
  redirect_uris            = ["https://app.example.com/callback"]
  first_party              = true
  skip_logout_confirmation = true
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
- `skip_logout_confirmation` (Boolean) — When true, an RP-initiated logout (`/oauth2/logout`) from this client with an exactly registered `post_logout_redirect_uri` ends the session without the "Sign out?" page and redirects straight back (back- and front-channel logout still run). **First-party clients only**: set `first_party` to `true`. Atlas refuses it on a third-party client (`422`), so the provider fails the plan instead when it can tell the client is third-party. Defaults to `false`.
- `oidc_metadata` (String, JSON) — OpenID metadata for the client as a JSON object (use `jsonencode`): `jwks`, `jwks_uri`, `tls_client_auth_subject_dn`, the `*_signed_response_alg` and `*_encrypted_response_alg` / `_enc` keys for `id_token`, `userinfo`, `authorization` and `introspection`, `request_object_signing_alg`, `request_object_encryption_alg` / `_enc`, `require_signed_request_object`, `request_uris`, `response_types`, `backchannel_token_delivery_mode` (`poll`, `ping`, `push`), `backchannel_client_notification_endpoint`, `backchannel_user_code_parameter`, `backchannel_authentication_request_signing_alg`, `org_selection` (`none`, `ask`, `require`), `min_app_version`, `native_sso_group`, `wimse_workload_id`, `tls_client_auth_san_*`, `dpop_bound_access_tokens` and `txn_token_unsigned_subject`. Set a key to `null` to clear it. Only the keys you list are managed, and the attribute is not read back after an import.

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
