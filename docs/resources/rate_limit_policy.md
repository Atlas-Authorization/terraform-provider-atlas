# atlas_rate_limit_policy

Create a rate-limit policy: a limit per window for one kind of subject (API key, user, organization, machine or IP). Atlas stores and distributes it; your service enforces it. Mark one policy per subject kind as the default, and add plan_overrides to give a billing plan its own limit. Precedence for a subject: a per-subject override, then an API key's own rate_limit, then the default policy's entry for the subject's plan, then the default policy. Names are unique (409); a second default for a subject kind is refused (409). Honours Idempotency-Key. Emits rate_limit_policy.updated.

Scopes: `rate_limit:read` to read, `rate_limit:write` to write.

Backend API: `POST /v1/rate_limit_policies` (create); `GET /v1/rate_limit_policies/{id}` (read); `PATCH /v1/rate_limit_policies/{id}` (update); `DELETE /v1/rate_limit_policies/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_rate_limit_policy" "example" {
  name = "tf-rate-limit-policy"
  subject = "key"
  limit = 1
  window_seconds = 1
}
```

## Schema

### Required

- `name` (String) Unique per instance.
- `subject` (String) What the limit counts: an API key, user, organization, machine or IP address. Fixed at create.
- `limit` (Number) Calls allowed per window.
- `window_seconds` (Number) The `window_seconds` field.

### Optional

- `burst` (Number) Extra calls the enforcing limiter may allow above the limit in one window. null: none.
- `plan_overrides` (String, JSON) Billing plan slug to its limit. A subject on that plan gets this limit instead of the policy limit; window_seconds and burst default to the policy values. A JSON document (use `jsonencode`).
- `is_default` (Boolean) The policy a subject of this kind gets when nothing more specific applies. One per subject kind.
- `enabled` (Boolean) A disabled policy applies to nobody.

### Read-Only

- `id` (String) Atlas object id.
- `version` (Number) Counts up on every change.
- `created_at` (Number) Epoch ms.
- `updated_at` (Number) Epoch ms.

## Import

The import id is `<id>`.

```sh
terraform import atlas_rate_limit_policy.example <id>
```
