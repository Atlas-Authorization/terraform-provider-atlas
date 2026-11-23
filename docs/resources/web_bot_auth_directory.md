# atlas_web_bot_auth_directory

Put one directory on the allow or block list; it leaves the other list if it was there.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `POST /v1/web_bot_auth/directories` (create); `GET /v1/web_bot_auth/directories` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/web_bot_auth/directories/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_web_bot_auth_directory" "example" {
  directory = "tf-web-bot-auth-directory"
  list = "allow"
}
```

## Schema

### Required

- `directory` (String) The `directory` field.
- `list` (String) The `list` field. One of: `allow`, `block`.

### Optional

- `name` (String) The `name` field.

### Read-Only

- `id` (String) Atlas object id.

## Import

The import id is `<id>`.

```sh
terraform import atlas_web_bot_auth_directory.example <id>
```
