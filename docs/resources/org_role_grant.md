# atlas_org_role_grant

Let a grantee organization hand out a bounded set of the grantor organization’s roles (role_keys) to its own members. One active grant per pair.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `POST /v1/org_role_grants` (create); `GET /v1/org_role_grants/{id}` (read); `PATCH /v1/org_role_grants/{id}` (update); `DELETE /v1/org_role_grants/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_org_role_grant" "example" {

}
```

## Schema

### Optional

- `grantor_org_id` (String) The `grantor_org_id` field. Changing it forces replacement.
- `grantee_org_id` (String) The `grantee_org_id` field. Changing it forces replacement.
- `role_keys` (List of String) The `role_keys` field.

### Read-Only

- `id` (String) Atlas object id.
- `status` (String) The `status` field. One of: `active`, `revoked`.
- `created_at` (Number) Created (epoch ms).
- `revoked_at` (Number) The `revoked_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_org_role_grant.example <id>
```
