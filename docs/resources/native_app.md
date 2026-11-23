# atlas_native_app

Register a native app that signs users in (RFC 8252): an `ios` app (bundle_id + team_id), an `android` app (package_name + sha256_cert_fingerprints) or a `development` app for Expo Go and dev builds (allowed_users + expires_at, at most 90 days ahead; it stops working at expiry and expires_at can be extended in place).

Scopes: `native_apps:read` to read, `native_apps:write` to write.

Backend API: `POST /v1/native_apps` (create); `GET /v1/native_apps/{id}` (read); `PATCH /v1/native_apps/{id}` (update); `DELETE /v1/native_apps/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_native_app" "example" {
  platform      = "ios"
  bundle_id     = "com.example.app"
  team_id       = "ABCDE12345"
  callback_urls = ["com.example.app://auth/callback"]
}
```

## Schema

### Required

- `platform` (String) The app kind. One of: `ios`, `android`, `development`. Changing it forces replacement.
- `callback_urls` (List of String) The callback URLs a sign-in may return to.

### Optional

- `bundle_id` (String) ios only: the app bundle identifier.
- `team_id` (String) ios only: the Apple developer team id.
- `package_name` (String) android only: the application id.
- `sha256_cert_fingerprints` (List of String) android only: the signing certificate SHA-256 fingerprints.
- `allowed_users` (List of String) development only: 1 to 20 user ids (`user_...`) or email addresses allowed to sign in through the app.
- `expires_at` (Number) development only: when the app stops working, epoch milliseconds in the future, at most 90 days ahead. Raise it to extend the app.

### Read-Only

- `id` (String) Atlas object id.
- `expired` (Boolean) True once a development app has passed expires_at.
- `created_at` (Number) Creation time (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_native_app.example <id>
```
