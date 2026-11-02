# atlas_network_acl

Create a network ACL rule (allow|deny, a CIDR or single address or a `zone_id`, a priority). A malformed CIDR is refused. First match in priority order decides; an allow-list denies every unnamed IP.

Scopes: `network_acls:read` to read, `network_acls:write` to write.

Backend API: `POST /v1/network_acls` (create); `GET /v1/network_acls/{id}` (read); `PATCH /v1/network_acls/{id}` (update); `DELETE /v1/network_acls/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_network_acl" "example" {

}
```

## Schema

### Optional

- `action` (String) The `action` field. One of: `allow`, `deny`.
- `cidr` (String) The `cidr` field.
- `zone_id` (String) The `zone_id` field.
- `description` (String) The `description` field.
- `priority` (Number) The `priority` field.
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `expires_at` (Number) The `expires_at` field. (epoch ms)
- `source` (String) The `source` field.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_network_acl.example <id>
```
