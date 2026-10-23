# atlas_better_auth_bridge

The Better Auth session bridge: moving from Better Auth to Atlas without signing
anyone out. While the bridge is open, a returning user's Better Auth session
cookie is exchanged once for an Atlas session — by the Atlas SDK middleware on
your server, or by the browser SDK. Atlas finds the session in your Better Auth
import, or (with the live fallback) asks your old app whether it is still valid,
then runs the normal sign-in checks on it, MFA included.

A singleton per instance. Each apply sets every field this resource manages;
destroying the resource turns the bridge off and deletes the stored secret. If
the settings are removed outside Terraform, the next plan re-creates them.

API: `GET`/`PUT`/`DELETE /v1/migrations/better_auth/bridge`.

## Example Usage

```hcl
variable "better_auth_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "atlas_better_auth_bridge" "this" {
  enabled       = true
  ends_at       = "2026-12-31T00:00:00Z"
  cookie_prefix = "better-auth"

  # Sessions created after the export: ask the old app, server to server.
  live_fallback = {
    enabled           = true
    base_url          = "https://app.example.com/api/auth"
    secret_wo         = var.better_auth_secret # write-only, never in state
    secret_wo_version = 1                      # bump to send a new secret
  }
}
```

To close the window early, set `enabled = false` (or destroy the resource).

## Schema

### Optional

- `enabled` (Boolean) — Whether Atlas accepts Better Auth sessions for exchange. Default `false`.
- `ends_at` (String) — End of the migration window, an RFC 3339 timestamp (e.g. `2026-12-31T00:00:00Z`). After it every exchange is refused. Unset = open until disabled.
- `cookie_prefix` (String) — Better Auth's `advanced.cookiePrefix` (1-64 letters, digits, `.`, `_` or `-`). Default `better-auth`; the session cookie is `<prefix>.session_token`.
- `live_fallback` (Attributes) — Validate a session missing from the import (created after the export) by asking the old app, server to server. Unset = off, and any stored secret is cleared. See [below for nested schema](#nestedatt--live_fallback).

### Read-Only

- `id` (String) — Always `better_auth_bridge`.
- `exchange_path` (String) — The Frontend API path the SDKs call for the browser exchange.

<a id="nestedatt--live_fallback"></a>
### Nested Schema for `live_fallback`

Optional:

- `enabled` (Boolean) — Turn the live check on. Requires `base_url`. Default `false`.
- `base_url` (String) — The old app's Better Auth base URL including its base path, e.g. `https://app.example.com/api/auth`; Atlas calls `<base_url>/get-session`. Must be a public https URL on a production instance.
- `secret_wo` (String, Sensitive, [Write-only](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments)) — The old app's `BETTER_AUTH_SECRET` (at least 16 characters). Lets Atlas check the cookie signature and sign a bare token for the live call. Never stored in state or returned by Atlas. It is sent on create and whenever `secret_wo_version` changes; change the version to rotate it, or change the version with `secret_wo` unset to clear it. Needs Terraform 1.11 or later.
- `secret_wo_version` (Number) — Any number; changing it sends `secret_wo` again (or clears the secret when `secret_wo` is unset).

Read-Only:

- `has_secret` (Boolean) — Whether Atlas holds a secret for the old app (never the secret).

The provider checks `cookie_prefix`, `ends_at`, `base_url` and the secret length at plan time.

## Import

```sh
terraform import atlas_better_auth_bridge.this better_auth_bridge
```

An imported `live_fallback` has no `secret_wo_version`; set one together with
`secret_wo` to replace the stored secret.
