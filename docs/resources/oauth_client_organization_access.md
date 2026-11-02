# atlas_oauth_client_organization_access

Set who may use an application. `open` (default): any organization. `restricted`: only organizations granted access (and, when a grant lists roles, members holding them); a sign-in outside an entitled organization is refused with `access_denied` at the authorize endpoint.

Scopes: `oauth_clients:write` to read, `oauth_clients:write` to write.

Backend API: `PUT /v1/oauth_clients/{id}/organization_access` (create and update); no read route: refresh keeps the last applied values; no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_oauth_client_organization_access" "example" {
  client_id = "tf-oauth-client-organization-access"
  mode = "open"
}
```

## Schema

### Required

- `client_id` (String) The client id this object belongs to. Changing it forces replacement.
- `mode` (String) The `mode` field. One of: `open`, `restricted`.

### Read-Only

- `id` (String) Atlas object id.

## Import

The import id is `<client_id>`.

```sh
terraform import atlas_oauth_client_organization_access.example <client_id>
```
