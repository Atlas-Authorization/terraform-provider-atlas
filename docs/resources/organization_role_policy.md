# atlas_organization_role_policy

Point an organization at a role set and/or give it its own default role (null clears either). Adding a member, accepting an invitation or changing a role to one outside the set is refused with 422 ROLE_NOT_IN_ROLE_SET; org:admin and the organization's own custom roles are always allowed, and directory syncs report what the directory says. Existing members are not changed.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `PUT /v1/organizations/{id}/role_policy` (create and update); `GET /v1/organizations/{id}/role_policy` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_organization_role_policy" "example" {
  organization_id = "tf-organization-role-policy"
}
```

## Schema

### Required

- `organization_id` (String) The organization id this object belongs to. Changing it forces replacement.

### Optional

- `role_set` (String) A role set id or key.
- `default_role` (String) The `default_role` field.

### Read-Only

- `id` (String) Atlas object id.
- `allowed_roles` (List of String) The role keys the organization may assign; null when unrestricted. Its own custom roles are always allowed too.
- `default_role_source` (String) The `default_role_source` field. One of: `organization`, `role_set`, `system`.

## Import

The import id is `<organization_id>`.

```sh
terraform import atlas_organization_role_policy.example <organization_id>
```
