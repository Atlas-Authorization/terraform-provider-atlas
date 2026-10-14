# atlas_oauth_provider

A configured **social** sign-in provider — a key from the Atlas provider catalog
(`google`, `github`, ... ~193 providers). The client secret is write-only: the
API never returns it, so its presence is tracked only via `has_secret` and by
whatever `credentials` sets. The redirect/callback URI is **derived** from your
instance host and returned read-only — register that exact string at the
provider's console.

You can also re-brand a provider's button (`display_name`, `icon_url`),
choose its sign-in `mode`, and configure the providers that have no
credentials — Steam and the wallet sign-ins `siwe`, `siws` and `siwf` — with
`settings` only.

> The attribute is named `provider_key` (not `provider`) because `provider` is a
> reserved Terraform meta-argument and cannot be used as a resource attribute.

Backend API (BAPI, secret-key `sk_`):
`PUT /v1/oauth_providers/:provider` (create/update — idempotent),
`POST /v1/oauth_providers/:provider/scope` (sign-in/sign-up scope),
`POST /v1/oauth_providers/:provider/enabled` (enable toggle),
`GET /v1/oauth_providers/:provider` (read),
`DELETE /v1/oauth_providers/:provider` (delete).

## Example Usage

```hcl
resource "atlas_oauth_provider" "github" {
  provider_key = "github"

  credentials = {
    client_id     = "Iv1.abc123"
    client_secret = var.github_client_secret # write-only; never read back
  }

  scopes        = ["read:user", "user:email"]
  enabled       = true
  allow_sign_in = true
  allow_sign_up = true
}

variable "github_client_secret" {
  type      = string
  sensitive = true
}

output "github_redirect_uri" {
  # Register this exact string at the GitHub OAuth app.
  value = atlas_oauth_provider.github.redirect_uri
}
```

A wallet sign-in has no credentials — configure its verifier with `settings`:

```hcl
resource "atlas_oauth_provider" "ethereum" {
  provider_key = "siwe"
  settings = {
    domain    = "app.example.com"
    chain_ids = "1,10"
    rpc_url   = "https://mainnet.example-rpc.com"
  }
  display_name = "Ethereum wallet"
  enabled      = true
}
```

A provider that supports a native id_token flow (e.g. Google One-Tap, Apple)
can be configured with just a `client_id` in `credentials` — no secret required.

## Schema

### Required

- `provider_key` (String) — Catalog key of the social provider (e.g. `google`, `github`). Immutable — forces replacement.

### Optional

- `credentials` (Map of String, Sensitive) — Provider credentials, keyed by the provider's own field keys — typically `client_id` and `client_secret`. The secret is write-only; never returned by the API. Omit for providers without credentials (Steam, `siwe`, `siws`, `siwf`) or to manage only the button's display settings.
- `settings` (Map of String) — Provider-specific settings, keyed by setting key: a team/tenant domain, `redirect_uri` (your own registered callback, on a domain the instance has verified), or the wallet verifier settings — `siwe`: `domain`, `chain_ids`, `rpc_url`; `siws`: `domain`, `networks`; `siwf`: `domain`, `chain_ids`, `optimism_rpc_url`, `id_registry_address`. Kept as configured; Atlas's stored (normalized) values are in `setting_values`.
- `display_name` (String) — Button label override; unset keeps the catalog name.
- `icon_url` (String) — Button icon override (http(s) or `data:image/...` URL).
- `mode` (String) — `native` (in-page button such as Google One Tap, where supported) or `redirect`. Unset keeps the provider's default.
- `scopes` (List of String) — OAuth scopes to request. Defaults to the provider's catalog default scopes when unset.
- `enabled` (Boolean) — Whether the provider is enabled for sign-in. Defaults to false. Enabling one with no usable credentials is refused.
- `allow_sign_in` (Boolean) — Whether existing users may sign in with this provider. Defaults to true.
- `allow_sign_up` (Boolean) — Whether new users may sign up (JIT account creation). Defaults to true. `allow_sign_in` and `allow_sign_up` cannot both be false — disable the provider instead.

### Read-Only

- `id` (String) — Resource id (equals the provider key).
- `client_id` (String) — The configured client id (not a secret).
- `has_secret` (Boolean) — Whether a client secret is stored.
- `configured` (Boolean) — Whether the provider has credentials configured.
- `redirect_uri` (String) — Redirect/callback URI derived from your instance host. Register this exact string at the provider.
- `setting_values` (Map of String) — Stored settings of a provider without credentials (the wallet sign-ins).
- `native` (Boolean) — Whether the provider supports the native (in-page) flow.
- `atlas_callback` (String) — Atlas's own callback URL for this provider.
- `custom_redirect_uri` (String) — Your own redirect URI, when `settings.redirect_uri` overrides Atlas's callback.
- `javascript_origins` (List of String) — For a native provider, the Authorized JavaScript origins to register at the provider console.
- `updated_at` (Number) — Last update time (epoch milliseconds).

## Import

```sh
terraform import atlas_oauth_provider.github github
```

Write-only `credentials` and `settings` are not recovered on import.
