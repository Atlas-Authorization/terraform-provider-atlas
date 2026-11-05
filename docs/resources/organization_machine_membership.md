# atlas_organization_machine_membership

Make a machine a member of an organization with a role. Holds no seat. 409 if it is already a member.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `POST /v1/organizations/{id}/machine_memberships` (create); `GET /v1/organizations/{id}/machine_memberships` (read: the list is scanned for the id); `PATCH /v1/organizations/{id}/machine_memberships/{machineId}` (update); `DELETE /v1/organizations/{id}/machine_memberships/{machineId}` (destroy).

## Example Usage

```hcl
resource "atlas_organization_machine_membership" "example" {
  organization_id = "tf-organization-machine-membership"
}
```

## Schema

### Required

- `organization_id` (String) The organization id this object belongs to. Changing it forces replacement.

### Optional

- `machine_id` (String) The `machine_id` field. Changing it forces replacement.
- `role` (String) The `role` field.

### Read-Only

- `id` (String) Atlas object id.
- `role_id` (String) The `role_id` field.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<organization_id>/<id>`.

```sh
terraform import atlas_organization_machine_membership.example <organization_id>/<id>
```
