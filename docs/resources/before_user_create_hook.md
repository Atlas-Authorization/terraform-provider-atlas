# atlas_before_user_create_hook

The `before_user_create` hook — a singleton per instance. Before Atlas creates a
user it calls your `url`, signed with the signing secret; your answer can set
`external_id` and metadata, or deny the sign-up. Destroying the resource switches
the hook off and clears its URL and secret.

API: `GET`/`PATCH /v1/instance/hooks/before_user_create`. Needs the
`instance:read` and `instance:write` scopes.

## Example Usage

```hcl
resource "atlas_before_user_create_hook" "this" {
  enabled                   = true
  url                       = "https://app.example.com/atlas/before-user-create"
  signing_secret_wo         = var.hook_signing_secret
  signing_secret_wo_version = 1
  timeout_ms                = 2000
  fail_closed               = true
}
```

## Schema

### Optional

- `enabled` (Boolean) — Call the hook on sign-up. Needs `url` and a signing secret. Default `false`.
- `url` (String) — The HTTPS endpoint Atlas calls (HTTP to localhost is allowed on a development instance). Unset clears it.
- `signing_secret_wo` (String, Sensitive, Write-only) — The secret Atlas signs each call with, 16 to 512 characters. Never stored in state or returned by Atlas. Sent on create and whenever `signing_secret_wo_version` changes (unset clears it).
- `signing_secret_wo_version` (Number) — Any number; changing it sends `signing_secret_wo` again (or clears the secret when it is unset).
- `timeout_ms` (Number) — How long Atlas waits for your answer, 500 to 5000 ms. Default 3000.
- `fail_closed` (Boolean) — Refuse the sign-up (`403 SIGN_UP_DENIED`) when the hook times out, fails or answers something Atlas cannot read, instead of creating the user without your fields. While your endpoint is down nobody can sign up. Default `false`.

### Read-Only

- `id` (String) — Always `before_user_create_hook`.
- `signing_secret_set` (Boolean) — Whether Atlas holds a signing secret for the hook.

## Import

```sh
terraform import atlas_before_user_create_hook.this before_user_create_hook
```
