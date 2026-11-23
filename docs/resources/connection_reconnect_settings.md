# atlas_connection_reconnect_settings

Replace the hosted-reconnect settings (a field left out returns to its default). With enabled true, the owner of an oauth connection that needs authorization again (a user, or an organization's technical contacts) gets one email per episode through the instance's own mail provider, never an Atlas key, with a link to a hosted page that restarts the provider consent. Nothing is sent while enabled is false, which is the default.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `PUT /v1/connection_reconnect_settings` (create and update); `GET /v1/connection_reconnect_settings` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_connection_reconnect_settings" "example" {

}
```

## Schema

### Optional

- `enabled` (Boolean) The `enabled` field.
- `link_ttl_hours` (Number) The `link_ttl_hours` field.
- `return_url` (String) An https URL on an allowed origin.
- `providers_value` (List of String) The `providers` field.

### Read-Only

- `id` (String) Atlas object id.

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_connection_reconnect_settings.example connection_reconnect_settings
```
