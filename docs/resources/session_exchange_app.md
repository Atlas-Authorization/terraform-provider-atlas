# atlas_session_exchange_app

Register a session-exchange app and its backchannel_logout_uri. When an Atlas session ends — a single revoke or 'sign out everywhere' — Atlas POSTs a signed logout_token to the URI so the app can end its own session. The URI must be a public https endpoint.

Scopes: `session_exchange_apps:read` to read, `session_exchange_apps:write` to write.

Backend API: `POST /v1/session_exchange_apps` (create); `GET /v1/session_exchange_apps/{id}` (read); `PATCH /v1/session_exchange_apps/{id}` (update); `DELETE /v1/session_exchange_apps/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_session_exchange_app" "example" {
  name = "tf-session-exchange-app"
  backchannel_logout_uri = "https://example.com/backchannel_logout_uri"
}
```

## Schema

### Required

- `name` (String) A human label for the downstream app.
- `backchannel_logout_uri` (String) Public https endpoint Atlas POSTs the signed logout_token to (application/x-www-form-urlencoded: logout_token=<jwt>) when a session ends.

### Optional

- `audience` (String) Optional override for the logout_token `aud` claim. Defaults to the app id when omitted.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_session_exchange_app.example <id>
```
