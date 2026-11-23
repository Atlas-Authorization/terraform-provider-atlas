# atlas_suppression

Add one address or up to 1000 to the suppression list. Adding one that is already there updates its reason. Audited; emits message_suppression.created for new entries.

Scopes: `messaging:read` to read, `messaging:write` to write.

Backend API: `POST /v1/suppressions` (create); `GET /v1/suppressions/{id}` (read); no update route: every change replaces the object; `DELETE /v1/suppressions/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_suppression" "example" {
  address = "tf-suppression"
}
```

## Schema

### Required

- `address` (String) One address. Send this or addresses.

### Optional

- `channel` (String) Default email. Phone channels take E.164 numbers; push a device token.
- `reason` (String) Default manual. One of: `hard_bounce`, `complaint`, `unsubscribe`, `invalid`, `manual`.
- `detail` (String) The `detail` field.

### Read-Only

- `id` (String) Atlas object id.
- `address_masked` (String) The `address_masked` field.
- `source` (String) `api`, `admin`, or the provider that reported it (`ses`, `brevo`, `twilio`, ...).
- `message_id` (String) The ledger message that caused an automatic entry.
- `created_at` (Number) The `created_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_suppression.example <id>
```
