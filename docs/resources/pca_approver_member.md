# atlas_pca_approver_member

Notification channels and contact details of one member of an existing approver. Members and their passkeys are enrolled outside Terraform (and a passkey is revoked there too); this resource manages the member settings. Destroying the resource leaves the member as it is. The API has no read route for one member, so refresh keeps the last applied values.

Scopes: `pca:read` to read, `pca:write` to write.

Backend API: `PATCH /v1/pca/approvers/{id}/members/{memberId}` (create and update); no read route: refresh keeps the last applied values; no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_pca_approver_member" "example" {
  approver_id = "pcaapp_example"
  member_id   = "pcamem_example"
  channels    = ["email"]
}
```

## Schema

### Required

- `approver_id` (String) The id of the approver. Changing it forces replacement.
- `member_id` (String) The id of the member. Changing it forces replacement.

### Optional

- `channels` (List of String) Where the member is notified. Each of: `push`, `email`, `sms`, `ciba`.
- `email` (String) The member email.
- `phone` (String) The member phone, E.164.

### Read-Only

- `id` (String) Atlas object id.

## Import

The import id is `<approver_id>/<member_id>`.

```sh
terraform import atlas_pca_approver_member.example <approver_id>/<member_id>
```
