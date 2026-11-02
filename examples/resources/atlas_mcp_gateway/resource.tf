resource "atlas_mcp_gateway" "billing" {
  name                   = "Billing tools"
  upstream_url           = "https://mcp.internal.example.com/mcp"
  upstream_authorization = "Bearer ${var.billing_mcp_token}"
  default_policy         = "deny"

  tools = jsonencode({
    lookup_invoice = { mask = { fields = [], presets = ["email"] } }
    refund = {
      scopes     = ["billing:refund"]
      conditions = [{ arg = "amount", op = "lte", value = 100 }]
      approval   = true
    }
  })
}

output "billing_gateway_url" {
  value = atlas_mcp_gateway.billing.url
}
