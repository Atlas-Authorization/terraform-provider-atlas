# atlas_email_risk_settings

Change email risk settings. `mode`: off, mx or provider (your own verification account; the key is write-only). With `gate_sign_up` a `block` verdict refuses the sign-up.

Scopes: `email_suppressions:read` to read, `email_suppressions:write` to write.

Backend API: `PUT /v1/fraud/email_risk/settings` (create and update); `GET /v1/fraud/email_risk/settings` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_email_risk_settings" "example" {

}
```

## Schema

### Optional

- `mode` (String) The `mode` field. One of: `off`, `mx`, `provider`.
- `provider_value` (String) The `provider` field. One of: `zerobounce`, `kickbox`.
- `credentials` (String, JSON, Sensitive) The `credentials` field. A JSON document (use `jsonencode`). Write-only: sent to Atlas, never returned, and not detected as drift.
- `block_disposable` (Boolean) The `block_disposable` field.
- `block_suppressed` (Boolean) The `block_suppressed` field.
- `block_undeliverable` (Boolean) The `block_undeliverable` field.
- `block_role` (Boolean) The `block_role` field.
- `gate_sign_up` (Boolean) The `gate_sign_up` field.

### Read-Only

- `id` (String) Atlas object id.
- `has_credentials` (Boolean) The `has_credentials` field.
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_email_risk_settings.example email_risk_settings
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
