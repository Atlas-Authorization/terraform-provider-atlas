# atlas_user_traits_schema

Store the user traits schema and/or change its enforcement. Omit schemas or enforcement to keep the stored one. Each schema is checked before it is stored (a 422 names the offending path): a subset of JSON Schema 2020-12 with type, properties, required, enum, default, items, maxLength, pattern, description and additionalProperties (true or false); any other keyword is refused, a root must have type "object", patterns that repeat a repeated group are refused, and a default must satisfy its own schema (it is documentation, never written into metadata). Enforcement: off checks nothing; warn lets every write through and records each violation in the audit log (field paths and messages, never values); enforce answers 422 with one error per field path on a write that breaks the schema. It applies to backend (secret key) writes, the dashboard and a signed-in user writing their own unsafe_metadata. Imports, migrations, SCIM provisioning and other Atlas-internal writes are never checked, so a bulk job is never stopped halfway. Existing metadata is not re-validated. Send version (the one you last read) to refuse a concurrent change with 409; sending the same content again changes nothing.

Scopes: `instance:read` to read, `instance:write` to write.

Backend API: `PUT /v1/instance/user_traits_schema` (create and update); `GET /v1/instance/user_traits_schema` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_user_traits_schema" "example" {

}
```

## Schema

### Optional

- `schemas` (String, JSON) A JSON Schema per target. Targets: public_metadata, unsafe_metadata, private_metadata, organization_public_metadata, organization_private_metadata. Each root has type "object". A JSON document (use `jsonencode`).
- `enforcement` (String) The `enforcement` field. One of: `off`, `warn`, `enforce`.
- `version` (Number) The version you last read; a different stored version answers 409.

### Read-Only

- `id` (String) Atlas object id.
- `updated_at` (Number) Epoch ms.

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_user_traits_schema.example user_traits_schema
```
