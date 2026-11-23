# atlas_rate_limit_override

Pin one subject (an API key, user, organization or machine of this instance, or one IP address) to a policy of its kind (policy_id) or to an inline limit (limit with window_seconds, optional burst). An override beats every other rule, including the key's own rate_limit. Replaces any earlier override for the subject; the same content again changes nothing. expires_at ends it on its own. An override pointing at a disabled policy does not apply. Emits rate_limit_policy.updated.

Scopes: `rate_limit:read` to read, `rate_limit:write` to write.

Backend API: `PUT /v1/rate_limit_overrides/{subject_type}/{subject_id}` (create and update); `GET /v1/rate_limit_overrides/{subject_type}/{subject_id}` (read); `DELETE /v1/rate_limit_overrides/{subject_type}/{subject_id}` (destroy).

## Example Usage

```hcl
resource "atlas_rate_limit_override" "example" {
  subject_type = "tf-rate-limit-override"
  subject_id = "tf-rate-limit-override"
}
```

## Schema

### Required

- `subject_type` (String) The subject type this object belongs to. Changing it forces replacement.
- `subject_id` (String) The subject id this object belongs to. Changing it forces replacement.

### Optional

- `policy_id` (String) Pin the subject to this policy (its subject kind must match). Send this, or limit with window_seconds.
- `limit` (Number) The `limit` field.
- `window_seconds` (Number) The `window_seconds` field.
- `burst` (Number) The `burst` field.
- `reason` (String) The `reason` field.
- `expires_at` (Number) A future time (epoch ms or ISO 8601). null: until removed.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<subject_type>/<subject_id>`.

```sh
terraform import atlas_rate_limit_override.example <subject_type>/<subject_id>
```
