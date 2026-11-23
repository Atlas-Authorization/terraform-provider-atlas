# atlas_messaging_push_provider

The push sender used for push MFA and in-app sign-in approval: `apns_fcm` (APNS and/or FCM credentials; devices are routed by platform) or `expo` (an Expo access token). Credentials are write-only. Destroying the resource leaves the sender configured; the API has no read route for it, so refresh keeps the last applied values. Needs the `messaging:write` scope.

Scopes: `messaging:write` to write.

Backend API: `PUT /v1/messaging_providers/push` (create and update); no read route: refresh keeps the last applied values; no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_messaging_push_provider" "example" {
  transport = "expo"
  config    = jsonencode({ access_token = var.expo_access_token })
}
```

## Schema

### Required

- `transport` (String) Which transport. One of: `apns_fcm`, `expo`.

### Optional

- `config` (String, JSON, Sensitive) The transport settings as a JSON document (use `jsonencode`): for apns_fcm `apns_auth_key`, `apns_key_id`, `apns_team_id`, `apns_topic`, `apns_host`, `fcm_service_account`; for expo `access_token`. Write-only: sent to Atlas, never returned, and not detected as drift. An omitted secret keeps the stored one.

### Read-Only

- `id` (String) Atlas object id.

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_messaging_push_provider.example messaging_push_provider
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
