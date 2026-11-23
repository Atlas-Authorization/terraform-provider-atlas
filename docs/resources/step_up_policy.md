# atlas_step_up_policy

Create a named step-up policy: how recent and how strong a re-authentication a sensitive action needs. Names are unique (400 when taken). A policy no method could satisfy (phishing_resistant with allowed_methods that exclude passkey) is refused. Honours Idempotency-Key.

Scopes: `step_up:read` to read, `step_up:write` to write.

Backend API: `POST /v1/step_up/policies` (create); `GET /v1/step_up/policies/{id}` (read); `PATCH /v1/step_up/policies/{id}` (update); `DELETE /v1/step_up/policies/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_step_up_policy" "example" {
  name = "tf-step-up-policy"
}
```

## Schema

### Required

- `name` (String) Lowercase letters, digits and `_ . : -`, starting with a letter or digit. Unique.

### Optional

- `description` (String) The `description` field.
- `enabled` (Boolean) The `enabled` field.
- `max_age_seconds` (Number) Default 300.
- `required_acr` (List of String) Up to 20 acr values; the session acr must be one of them.
- `required_amr` (List of String) Up to 20 amr values, all required (for example `["pwd","mfa"]`).
- `required_amr_any` (List of String) Up to 20 amr values, at least one required.
- `phishing_resistant` (Boolean) Needs a passkey among allowed_methods, or no allowed_methods.
- `allowed_methods` (List of String) The `allowed_methods` field. One of: `password`, `passkey`, `totp`, `backup_code`, `email_code`, `phone_code`, `push`, `duo`, `yubikey`.
- `receipt_ttl_seconds` (Number) Default 60.

### Read-Only

- `id` (String) Atlas object id.
- `accepted_factors` (List of String) The step-up strategies that can satisfy it, after phishing resistance narrows allowed_methods. One of: `password`, `passkey`, `totp`, `backup_code`, `email_code`, `phone_code`, `push`, `duo`, `yubikey`.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_step_up_policy.example <id>
```
