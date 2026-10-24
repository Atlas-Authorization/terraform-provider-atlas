# atlas_id_app

A **Sign in with Atlas** app: an OpenID Connect client on the shared Atlas
issuer (`https://atlasauth.net`), registered to your platform account. Any site
can then offer "Sign in with Atlas" using the app's `client_id`.

This resource uses the platform surface, so the provider needs a **platform
key** (`psk_...`) — set `platform_key` in the provider block or
`ATLAS_PLATFORM_KEY` in the environment.

A confidential app's `client_secret` is returned once, on create (and again on
each rotation), and kept as a sensitive value in state. Change
`rotate_secret_trigger` to rotate it in place; the previous secret keeps working
for 24 hours.

`branding` is a JSON object (write it with `jsonencode`). It is compared
**semantically**: the key order, formatting and `null` members of what Atlas
returns never cause drift. (Editing only the whitespace or key order of a hand-written JSON string still shows one
in-place update, because Terraform compares configuration text; `jsonencode` output
is canonical, so it never does.)
A key you remove from the configuration is cleared on the server.

**Adoption.** Set `grants_managed_by_instance_id` to an instance of the same
account to hand it the app's grants: that instance's scoped key can then list and
revoke them, while the app stays editable here. Removing the attribute releases the
app. `managed_via` reports `instance_toggle`, `adopted` or unset.

API: `POST /v1/platform/atlas_id_apps`, `GET`/`PATCH`/`DELETE
/v1/platform/atlas_id_apps/:id`, `POST /v1/platform/atlas_id_apps/:id/rotate_secret`, `PUT /v1/platform/atlas_id_apps/:id/managed_by`.

## Example Usage

```hcl
provider "atlas" {
  # platform_key = "psk_live_..."  # or ATLAS_PLATFORM_KEY (preferred)
}

resource "atlas_id_app" "acme" {
  name          = "Acme"
  redirect_uris = ["https://acme.example.com/auth/callback"]

  post_logout_redirect_uris           = ["https://acme.example.com/"]
  backchannel_logout_uri              = "https://acme.example.com/auth/backchannel-logout"
  backchannel_logout_session_required = true

  allowed_scopes = ["openid", "profile", "email", "offline_access"]
  logo_url       = "https://acme.example.com/logo.png"
  homepage_url   = "https://acme.example.com"

  branding = jsonencode({
    theme                = "dark"
    accent_color         = "#4f46e5"
    privacy_policy_url   = "https://acme.example.com/privacy"
    terms_of_service_url = "https://acme.example.com/terms"
    copy = {
      sign_in_title   = "Sign in to {app}"
      consent_message = "Acme will see your name and email address."
      support_url     = "https://acme.example.com/help"
    }
  })

  # Bump to rotate the client secret.
  rotate_secret_trigger = "2026-10"
}

output "acme_client_id" {
  value = atlas_id_app.acme.client_id
}
```

## Schema

### Required

- `name` (String) — App name shown on the sign-in and consent screens (at most 100 characters).
- `redirect_uris` (List of String) — Exact-match redirect URIs (https, or http on localhost).

### Optional

- `type` (String) — `confidential` (default; gets a client secret) or `public` (PKCE, no secret). Changing it forces replacement.
- `post_logout_redirect_uris` (List of String) — Where `/oauth2/logout` may send the user after sign-out. Exact match. Defaults to none.
- `backchannel_logout_uri` (String) — OIDC Back-Channel Logout: a signed `logout_token` is POSTed here when the user signs out.
- `backchannel_logout_session_required` (Boolean) — Whether the app needs `sid` in the logout token. Default `false`.
- `frontchannel_logout_uri` (String) — OIDC Front-Channel Logout: loaded in a hidden iframe on the Atlas sign-out page.
- `frontchannel_logout_session_required` (Boolean) — Add `sid` (and `iss`) to the front-channel logout URL. Default `false`.
- `allowed_scopes` (List of String) — `openid` (required), `profile`, `email`, `offline_access` (refresh tokens). Defaults to `openid profile email`.
- `logo_url` (String) — https URL of the app logo.
- `homepage_url` (String) — https URL of the app homepage.
- `branding` (String, JSON object) — Page look for this app: `theme` (`auto` | `light` | `dark`), `accent_color` and `background_color` (`#RRGGBB`), `privacy_policy_url`, `terms_of_service_url`, `copy` (`sign_in_title`, `sign_in_subtitle`, `sign_up_title`, `sign_up_subtitle`, `consent_message`, `support_url`, `support_label`; `{app}` is replaced with the app name) and `design` (sign-in page design settings, the same shape as `atlas_instance_design.design`; images must be uploaded to your account).
- `grants_managed_by_instance_id` (String) — Adopt the app into an instance of the same account; remove to release.
- `rotate_secret_trigger` (String) — Any value; changing it rotates the client secret in place. Confidential apps only.

### Read-Only

- `id` (String) — App id.
- `client_id` (String) — The OAuth client id.
- `client_secret` (String, Sensitive) — The client secret, returned on create and on rotation only.
- `secret_prefix` (String) — The first characters of the current secret.
- `managed_by_instance_id` (String) — The instance that manages the app's grants (a Sign in with Atlas switch, or adoption).
- `managed_via` (String) — `instance_toggle`, `adopted`, or unset.
- `created_at` (Number) — Creation time (epoch ms).
- `updated_at` (Number) — Last update time (epoch ms).

## Import

```sh
terraform import atlas_id_app.acme oac_123
```

The client secret cannot be recovered on import. Set or change
`rotate_secret_trigger` to mint a new one into state.
