# atlas_before_first_session_hook

The `before_first_session` hook: called once per user, before their first session token, so account linking can stamp `external_id` on the user. Same shape and signing as the `before_user_create` hook. Destroying the resource switches the hook off and clears its URL and secret.

Scopes: `instance:read` to read, `instance:write` to write.

Backend API: `PATCH /v1/instance/hooks/before_first_session` (create and update); `GET /v1/instance/hooks/before_first_session` (read); `DELETE /v1/instance/hooks/before_first_session` (destroy).

## Example Usage

```hcl
resource "atlas_before_first_session_hook" "example" {
  enabled        = true
  url            = "https://example.com/hooks/first-session"
  signing_secret = var.hook_signing_secret
}
```

## Schema

### Optional

- `enabled` (Boolean) Whether the hook is called.
- `url` (String) The https URL Atlas calls.
- `signing_secret` (String, Sensitive) The secret the call is signed with. Write-only: sent to Atlas, never returned, and not detected as drift.
- `timeout_ms` (Number) How long Atlas waits for the answer, in milliseconds.

### Read-Only

- `id` (String) Atlas object id.
- `signing_secret_set` (Boolean) Whether a signing secret is stored.
- `last_call` (String, JSON) The most recent call: when, its status and any error.
- `failed_calls` (Number) Calls that failed since the last good one.
- `updated_at` (Number) Last change (epoch ms).

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_before_first_session_hook.example before_first_session_hook
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
