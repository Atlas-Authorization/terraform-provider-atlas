# atlas_scim_resource_type

Define a custom SCIM resource type: a name, an endpoint served at /scim/v2/<endpoint>, a schema URN and attribute definitions (type, multiValued, required, mutability). Directory tokens then create, read, filter, patch and delete resources of it, per organization.

Scopes: `scim:read` to read, `scim:write` to write.

Backend API: `POST /v1/scim_resource_types` (create); `GET /v1/scim_resource_types/{id}` (read); no update route: every change replaces the object; `DELETE /v1/scim_resource_types/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_scim_resource_type" "example" {
  name = "tf-scim-resource-type"
  endpoint = "https://example.com/endpoint"
  schema_uri = "https://example.com/schema_uri"
}
```

## Schema

### Required

- `name` (String) The `name` field.
- `endpoint` (String) Letters and digits; not Users, Groups, Bulk, Me, Schemas, ResourceTypes or ServiceProviderConfig.
- `schema_uri` (String) A URN, e.g. urn:example:params:scim:schemas:Device.

### Optional

- `description` (String) The `description` field.
- `attributes` (String, JSON) The `attributes` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) The `created_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_scim_resource_type.example <id>
```
