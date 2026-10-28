# Add a server to the registry as approved (default), or file it as pending for review.
resource "atlas_mcp_registry_entry" "example" {
  name = "tf-mcp-registry-entry"
  url = "https://example.com/url"
}
