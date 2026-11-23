# Track a document (an OpenAPI spec, an MCP tool list and so on) as a source of agent actions: each sync turns what it lists into drafts for review, on a schedule or on demand.
resource "atlas_pca_action_source" "example" {
  ns       = "billing"
  from     = "openapi"
  url      = "https://example.com/openapi.json"
  schedule = "daily"
}
