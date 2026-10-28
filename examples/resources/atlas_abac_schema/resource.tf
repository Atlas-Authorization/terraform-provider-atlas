# Set the ABAC entity schema: `entity_types` (attributes with `boolean`, `long`, `string`, `set`, `entity` and `record` types, `required: false` for optional ones, and `member_of`) and `actions` (`applies_to` principal and resource types and a `context`).
resource "atlas_abac_schema" "example" {
  schema = jsonencode({})
}
