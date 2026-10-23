# atlas_policy_settings

Instance-wide policy acceptance settings — a singleton per instance. Destroying
the resource turns the `pol` session claim off.

API: `GET`/`PATCH /v1/instance/policy_settings`.

## Example Usage

```hcl
resource "atlas_policy_settings" "this" {
  session_claim = true
}
```

## Schema

### Required

- `session_claim` (Boolean) — Put the `pol` claim — the accepted version of every required policy, for example `{"tos": "2026-10-24"}` — on session tokens, so your backend can check acceptance without a lookup. Tokens minted after the change carry it; it can take up to 30 seconds to apply everywhere.

### Read-Only

- `id` (String) — Always `policy_settings`.

## Import

```sh
terraform import atlas_policy_settings.this policy_settings
```
