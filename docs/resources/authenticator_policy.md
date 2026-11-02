# atlas_authenticator_policy

Create an authenticator policy for a group: `required_factors` (members must pass a second factor, or enrol one), `grace_until` (members with no factor may still sign in until then), and a `passkey` profile (user verification, attestation, attachment, AAGUID allow / block lists) applied when members register passkeys. Policies only add protection to the instance settings; several policies combine to the strictest result.

Scopes: `access_policies:read` to read, `access_policies:write` to write.

Backend API: `POST /v1/authenticator_policies` (create); `GET /v1/authenticator_policies/{id}` (read); `PATCH /v1/authenticator_policies/{id}` (update); `DELETE /v1/authenticator_policies/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_authenticator_policy" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `group` (String) The `group` field.
- `enabled` (Boolean) The `enabled` field.
- `required_factors` (List of String) The `required_factors` field. One of: `totp`, `sms`, `passkey`, `push`.
- `grace_until` (Number) The `grace_until` field.
- `passkey` (String, JSON) The `passkey` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_authenticator_policy.example <id>
```
