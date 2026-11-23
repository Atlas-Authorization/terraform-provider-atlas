# atlas_mfa_campaign

Start an MFA rollout campaign: ask everyone, a group, an organization or an organization role to enrol a second factor. Before grace_until users are nagged on the client prompt (GET /v1/client/me/mfa/campaign) every nag_interval_days; with enforce, sign-in after the deadline requires a second factor. The target must exist and each required factor must be enabled for the application (400 otherwise). Honours Idempotency-Key.

Scopes: `mfa_campaigns:read` to read, `mfa_campaigns:write` to write.

Backend API: `POST /v1/mfa_campaigns` (create); `GET /v1/mfa_campaigns/{id}` (read); `PATCH /v1/mfa_campaigns/{id}` (update); `DELETE /v1/mfa_campaigns/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_mfa_campaign" "example" {
  name = "tf-mfa-campaign"
}
```

## Schema

### Required

- `name` (String) The `name` field.

### Optional

- `target_type` (String) Default `all`. Changing it forces replacement.
- `target_value` (String) Required for group (slug), org (id) and role (key). Changing it forces replacement.
- `required_factors` (List of String) The `required_factors` field. One of: `totp`, `sms`, `passkey`, `push`.
- `starts_at` (Number) Start; defaults to now (epoch ms). Changing it forces replacement.
- `grace_until` (Number) The deadline; must be after starts_at (epoch ms).
- `nag_interval_days` (Number) Default 3.
- `enforce` (Boolean) Default false.

### Read-Only

- `id` (String) Atlas object id.
- `status` (String) The `status` field. One of: `active`, `ended`.
- `ended_at` (Number) The `ended_at` field. (epoch ms)
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).
- `metrics` (String, JSON) The `metrics` field. A JSON document (use `jsonencode`).

## Import

The import id is `<id>`.

```sh
terraform import atlas_mfa_campaign.example <id>
```
