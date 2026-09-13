# atlas_oauth_provider (Data Source)

Look up a social sign-in provider by its catalog key. The client secret is never
returned; only `has_secret` reports its presence.

> The lookup attribute is named `provider_key` (not `provider`) because
> `provider` is a reserved Terraform meta-argument.

Backend API (BAPI): `GET /v1/oauth_providers/:provider`.

## Example Usage

```hcl
data "atlas_oauth_provider" "google" {
  provider_key = "google"
}

output "google_redirect_uri" {
  value = data.atlas_oauth_provider.google.redirect_uri
}
```

## Schema

### Required

- `provider_key` (String) — Catalog key of the provider to look up (e.g. `google`).

### Read-Only

- `display_name` (String) — Human-readable provider name.
- `category` (String) — Catalog category.
- `tier` (String) — Catalog tier.
- `redirect_uri` (String) — Redirect/callback URI derived from your instance host.
- `default_scopes` (List of String) — The provider's default scopes.
- `configured` (Boolean) — Whether credentials are configured.
- `client_id` (String) — The configured client id (not a secret).
- `has_secret` (Boolean) — Whether a client secret is stored.
- `enabled` (Boolean) — Whether the provider is enabled.
- `allow_sign_in` (Boolean) — Whether existing users may sign in with this provider.
- `allow_sign_up` (Boolean) — Whether new users may sign up with this provider.
- `scopes` (List of String) — The configured scopes.
- `updated_at` (Number) — Last update time (epoch milliseconds).
