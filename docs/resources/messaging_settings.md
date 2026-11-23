# atlas_messaging_settings

Change messaging behaviour. Send only the fields to change. Audited.

Scopes: `messaging:read` to read, `messaging:write` to write.

Backend API: `PUT /v1/messaging/settings` (create and update); `GET /v1/messaging/settings` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_messaging_settings" "example" {

}
```

## Schema

### Optional

- `auto_suppress` (Boolean) Put an address on the suppression list when your provider reports a hard bounce, a spam complaint, an unsubscribe or an invalid number. Default true.
- `failover_enabled` (Boolean) Try your fallback providers when the first one errors. Default true.
- `ledger_retention_days` (Number) How long ledger rows are kept. Default 90.

### Read-Only

- `id` (String) Atlas object id.
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_messaging_settings.example messaging_settings
```
