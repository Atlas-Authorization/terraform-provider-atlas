# atlas_policy_audience

Set a policy's audience. orgs asks only members of the listed organizations (1-200, all must exist) at their next sign-in, and never at sign-up; everyone (the default) clears the list. A user who is in none of the organizations is not asked and is not blocked.

Scopes: `instance:read` to read, `instance:write` to write.

Backend API: `PUT /v1/policies/{id}/audience` (create and update); `GET /v1/policies/{id}/audience` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_policy_audience" "example" {
  policy_id = "tf-policy-audience"
  audience = "everyone"
}
```

## Schema

### Required

- `policy_id` (String) The policy id this object belongs to. Changing it forces replacement.
- `audience` (String) The `audience` field. One of: `everyone`, `orgs`.

### Optional

- `organization_ids` (List of String) The `organization_ids` field.

### Read-Only

- `id` (String) Atlas object id.
- `policy` (String) The `policy` field.

## Import

The import id is `<policy_id>`.

```sh
terraform import atlas_policy_audience.example <policy_id>
```
