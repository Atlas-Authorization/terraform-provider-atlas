# atlas_email_suppression

Add one or many addresses (a bounce or complaint webhook forwarder can post here). Existing entries are updated.

Scopes: `email_suppressions:read` to read, `email_suppressions:write` to write.

Backend API: `POST /v1/messaging/suppressions` (create); `GET /v1/messaging/suppressions` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/messaging/suppressions/{email}` (destroy).

## Example Usage

```hcl
resource "atlas_email_suppression" "example" {
  email = "user@example.com"
}
```

## Schema

### Required

- `email` (String) The `email` field.

### Optional

- `reason` (String) The `reason` field. One of: `hard_bounce`, `complaint`, `unsubscribe`, `manual`, `invalid`.
- `source` (String) The `source` field.
- `detail` (String) The `detail` field.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Added (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_email_suppression.example <id>
```
