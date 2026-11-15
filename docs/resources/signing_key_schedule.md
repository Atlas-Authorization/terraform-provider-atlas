# atlas_signing_key_schedule

Rotate the instance signing key automatically. Every interval_days the worker publishes a new key (prepublish_hours before it signs), switches to it, and keeps the previous key published until every token it signed has expired: at least the longest token lifetime the instance uses plus the key set cache, never under 24 hours, and never dropped while one of its tokens can still be presented. Keys in a cloud KMS rotate at the scheduled time (they cannot be held back) and are limited to RS256 and ES256. An earlier rotation still inside its grace window postpones the next one. Needs a recent MFA sign-in.

Scopes: `instance:read` to read, `instance:write` to write.

Backend API: `PUT /v1/instance/signing_key_schedule` (create and update); `GET /v1/instance/signing_key_schedule` (read); `DELETE /v1/instance/signing_key_schedule` (destroy).

## Example Usage

```hcl
resource "atlas_signing_key_schedule" "example" {
  interval_days = 1
}
```

## Schema

### Required

- `interval_days` (Number) The `interval_days` field.

### Optional

- `enabled` (Boolean) Default true.
- `alg` (String) RS256, ES256, ES384, PS256 or EdDSA. Omit to keep the current algorithm.
- `grace_period_seconds` (Number) At least the instance minimum (see effective_grace_seconds).
- `prepublish_hours` (Number) Default 24.

### Read-Only

- `id` (String) Atlas object id.
- `effective_grace_seconds` (Number) What the previous key is kept for in practice: never shorter than the longest token the instance signs lives (plus the key set cache), and never under 24 hours.
- `next_rotation_at` (Number) Epoch ms of the next rotation.
- `last_rotated_at` (Number) Epoch ms of the last rotation the schedule made.
- `last_error` (String) The `last_error` field.
- `last_error_at` (Number) Epoch ms of the last failed attempt.

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_signing_key_schedule.example signing_key_schedule
```
