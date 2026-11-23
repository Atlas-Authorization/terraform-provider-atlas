# atlas_org_role_set

Create a role set. role_keys must be instance-wide roles that exist; org:admin is always allowed and need not be listed. 409 when the key is taken. Honours Idempotency-Key.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `POST /v1/org_role_sets` (create); `GET /v1/org_role_sets/{id}` (read); `PATCH /v1/org_role_sets/{id}` (update); `DELETE /v1/org_role_sets/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_org_role_set" "example" {
  key = "tf-org-role-set"
  name = "tf-org-role-set"
}
```

## Schema

### Required

- `key` (String) Unique per instance. Fixed at create.
- `name` (String) The `name` field.

### Optional

- `description` (String) The `description` field.
- `role_keys` (List of String) Instance-wide roles an organization using this set may assign. org:admin is always allowed.
- `default_role_key` (String) The role a new seat gets when none is named. Must be one of role_keys (or org:admin).

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_org_role_set.example <id>
```
