# atlas_org_settings_schema

The per-instance org-settings JSON Schema registry. An instance registers ONE
JSON Schema describing the shape of its organization-level settings; Atlas then
validates every per-organization settings write against it, and versions and
audits the schema. This resource is a **singleton per instance** — its `id` is
always `organization_settings_schema`. Each change re-registers the schema and
bumps `version`. Atlas exposes no delete for the registry, so destroying this
resource only removes it from Terraform state; the registered schema stays in
place.

Backend API (BAPI): `PUT /v1/organization_settings_schema` (register / update),
`GET /v1/organization_settings_schema` (read). There is no delete route.

## Example Usage

```hcl
resource "atlas_org_settings_schema" "default" {
  schema = jsonencode({
    type = "object"
    properties = {
      brand_color   = { type = "string" }
      support_email = { type = "string", format = "email" }
      seats_limit   = { type = "integer", minimum = 1 }
    }
    required = ["support_email"]
  })
}
```

## Schema

### Required

- `schema` (String) — The JSON Schema document as a JSON string (use
  `jsonencode(...)`). Must be a JSON object describing the org-settings shape,
  e.g. `{"type":"object","properties":{...}}`.

### Read-Only

- `id` (String) — Always `organization_settings_schema` (the registry is a
  singleton per instance).
- `version` (Number) — Schema version, incremented by Atlas on every
  registration.
- `updated_at` (Number) — Last registration time (epoch ms).

## Import

```sh
terraform import atlas_org_settings_schema.default organization_settings_schema
```
