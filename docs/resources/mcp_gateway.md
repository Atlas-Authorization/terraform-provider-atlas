# atlas_mcp_gateway

An MCP tool gateway: one URL that fronts your MCP server for Atlas agents. The
agent presents its agent token; the gateway lists only the tools it may use and
enforces per-tool scopes, argument conditions, human approval and response
masking on each call, and audits every call.

Backend API (BAPI): `POST /v1/mcp_gateways` (create), `GET /v1/mcp_gateways/:id`
(read), `PATCH /v1/mcp_gateways/:id` (update), `DELETE /v1/mcp_gateways/:id`
(delete). The API key needs the `machines:read` and `machines:write` scopes.

## Example Usage

```hcl
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
```

## Argument Reference

- `name` (Required) Display name (1 to 100 characters); also the server name agents see.
- `upstream_url` (Required) The http(s) URL of your MCP server (streamable HTTP).
- `upstream_authorization` (Optional, Sensitive) The `Authorization` header value sent to your MCP server. Write-only: stored encrypted and never read back. Remove it to clear it.
- `default_policy` (Optional) `deny` (default): only tools listed in `tools` are reachable. `allow`: every upstream tool is, unless its policy says otherwise.
- `tools` (Optional) Per-tool policy as a JSON object keyed by tool name (use `jsonencode`): `allow`, `scopes`, `conditions` (`{ arg, op, value }` with `eq`, `neq`, `in`, `not_in`, `lte`, `gte`, `present`), `approval` and `mask` (`fields`, `presets`). Compared semantically; replaced as a whole on each update.
- `enabled` (Optional) Whether the gateway answers. Defaults to `true`.

## Attribute Reference

- `id` The gateway id.
- `url` The gateway URL agents connect to.
- `has_upstream_authorization` Whether an upstream authorization is stored.

## Import

```shell
terraform import atlas_mcp_gateway.billing mgw_…
```
