# Define a custom SCIM resource type: a name, an endpoint served at /scim/v2/<endpoint>, a schema URN and attribute definitions (type, multiValued, required, mutability).
resource "atlas_scim_resource_type" "example" {
  name = "tf-scim-resource-type"
  endpoint = "https://example.com/endpoint"
  schema_uri = "https://example.com/schema_uri"
}
