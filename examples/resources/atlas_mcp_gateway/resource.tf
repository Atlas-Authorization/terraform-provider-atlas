# Create an MCP gateway: one URL (`url` in the response) that fronts your MCP server.
resource "atlas_mcp_gateway" "example" {
  name = "tf-mcp-gateway"
}
