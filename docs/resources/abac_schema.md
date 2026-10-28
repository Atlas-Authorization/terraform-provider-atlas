# atlas_abac_schema

Set the ABAC entity schema: `entity_types` (attributes with `boolean`, `long`, `string`, `set`, `entity` and `record` types, `required: false` for optional ones, and `member_of`) and `actions` (`applies_to` principal and resource types and a `context`). Every problem is reported with its path. Once set, policies and templates are type-checked when saved: unknown attributes, optional attributes read without a `has` guard, wrong operand types, unknown entity types and actions, and actions that do not apply to the policy’s principal and resource types. The response lists the errors the schema would give the policies and links already stored.

Scopes: `abac:read` to read, `abac:write` to write.

Backend API: `PUT /v1/abac_schema` (create and update); `GET /v1/abac_schema` (read); `DELETE /v1/abac_schema` (destroy).

## Example Usage

```hcl
resource "atlas_abac_schema" "example" {
  schema = jsonencode({})
}
```

## Schema

### Required

- `schema` (String, JSON) The `schema` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `version` (Number) 0 when none is set.
- `updated_at` (Number) The `updated_at` field. (epoch ms)
- `policy_issues` (String, JSON) On a save: the errors this schema would give the policies and links already stored (the schema is saved anyway; they must be fixed before those are edited). A JSON document (use `jsonencode`).

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_abac_schema.example abac_schema
```
