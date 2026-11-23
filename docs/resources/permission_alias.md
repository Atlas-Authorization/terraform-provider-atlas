# atlas_permission_alias

An alias for a permission key: the alias keeps working in checks and role definitions after the permission is renamed. The alias must not be a permission key itself and the target must exist.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `PUT /v1/permission_aliases/{alias}` (create and update); `GET /v1/permission_aliases` (read: the list is scanned for the key); `DELETE /v1/permission_aliases/{alias}` (destroy).

## Example Usage

```hcl
resource "atlas_permission_alias" "example" {
  alias          = "billing:view"
  permission_key = "invoices:read"
}
```

## Schema

### Required

- `alias` (String) The alias key. Changing it forces replacement.
- `permission_key` (String) The permission the alias resolves to.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Creation time (epoch ms).

## Import

The import id is `<alias>`.

```sh
terraform import atlas_permission_alias.example <alias>
```
