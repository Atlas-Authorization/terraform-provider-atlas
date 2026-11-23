# atlas_pca_approvals_settings

Human-approval settings for agent actions: the approval channels, the default channels for approver members, the CIBA backchannel and the Slack webhook. Destroying the resource leaves the settings in place.

Scopes: `pca:read` to read, `pca:write` to write.

Backend API: `PUT /v1/pca/approvals/settings` (create and update); `GET /v1/pca/approvals/settings` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_pca_approvals_settings" "example" {
  channels = jsonencode(["email"])
}
```

## Schema

### Optional

- `channels` (String, JSON) The enabled approval channels.
- `default_member_channels` (String, JSON) The channels a new approver member starts with.
- `ciba` (String, JSON) The CIBA backchannel settings (enabled, mode and its endpoints).
- `slack_webhook_url` (String, Sensitive) The Slack incoming-webhook URL. Write-only: sent to Atlas, never returned, and not detected as drift.
- `ciba_client_secret` (String, Sensitive) The CIBA client secret. Write-only: sent to Atlas, never returned, and not detected as drift.

### Read-Only

- `id` (String) Atlas object id.
- `slack_webhook_set` (Boolean) Whether a Slack webhook is stored.

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_pca_approvals_settings.example pca_approvals_settings
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
