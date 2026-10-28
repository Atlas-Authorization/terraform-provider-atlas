# atlas_mcp_gateway_upstream

Add an upstream. An mcp upstream's tools are re-served as `<namespace>__<tool>`; an openapi upstream turns each operation of the document into a tool and calls the API itself (through a connection's relay when auth_kind is connection). At most 20 per gateway.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `POST /v1/mcp_gateways/{id}/upstreams` (create); `GET /v1/mcp_gateways/{id}/upstreams/{upstreamId}` (read); `PATCH /v1/mcp_gateways/{id}/upstreams/{upstreamId}` (update); `DELETE /v1/mcp_gateways/{id}/upstreams/{upstreamId}` (destroy).

## Example Usage

```hcl
resource "atlas_mcp_gateway_upstream" "example" {
  mcp_gateway_id = "tf-mcp-gateway-upstream"
}
```

## Schema

### Required

- `mcp_gateway_id` (String) The mcp gateway id this object belongs to. Changing it forces replacement.

### Optional

- `namespace` (String) 1 to 32 lowercase letters and digits, starting with a letter.
- `kind` (String) The `kind` field. One of: `mcp`, `openapi`, `actions`.
- `action_names` (List of String) actions upstreams: expose only these actions.
- `url` (String) The MCP server URL, or the API base URL (defaults to the document's first server).
- `openapi` (String, JSON) openapi upstreams: an OpenAPI 3.x document. Each operation becomes a tool named by its operationId. A JSON document (use `jsonencode`). Not returned by the API, so it is not detected as drift.
- `auth_kind` (String) static sends `authorization` as the Authorization header; connection uses a connection's credential. One of: `none`, `static`, `connection`.
- `authorization` (String) Write-only Authorization header value for auth_kind static. Not returned by the API, so it is not detected as drift.
- `connection_id` (String) auth_kind connection: the connection to use.
- `connection_owner` (String) org uses connection_id itself; agent_owner uses the calling agent's (or user's) own connection to the same provider, created unconnected on first use.
- `connect_redirect_url` (String) Where a person returns after connecting; the link in an authorization_required answer.
- `stream` (Boolean) mcp upstreams: pass server-sent events through.
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `gateway_id` (String) The `gateway_id` field.
- `has_authorization` (Boolean) The `has_authorization` field.
- `tool_count` (Number) openapi upstreams: how many operations became tools.
- `created_at` (Number) The `created_at` field. (epoch ms)

## Import

The import id is `<mcp_gateway_id>/<id>`.

```sh
terraform import atlas_mcp_gateway_upstream.example <mcp_gateway_id>/<id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
