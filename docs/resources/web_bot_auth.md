# atlas_web_bot_auth

Set the allow and/or block list (each replaces its list; omit one to keep it). Entries are an https directory URL or host, or { directory, name }. At most 100 per list; a directory cannot be on both.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `PUT /v1/web_bot_auth` (create and update); `GET /v1/web_bot_auth` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_web_bot_auth" "example" {

}
```

## Schema

### Optional

- `allow` (String, JSON) The `allow` field. A JSON document (use `jsonencode`).
- `block` (String, JSON) The `block` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_web_bot_auth.example web_bot_auth
```
