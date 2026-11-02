# atlas_mcp_gateway

Create an MCP gateway: one URL (`url` in the response) that fronts your MCP server. `tools` maps a tool name to its policy (allow, scopes the agent token must hold, argument conditions, approval, response masking); `default_policy` (deny by default) decides tools not listed. upstream_authorization is write-only and sent to the upstream as its Authorization header.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `POST /v1/mcp_gateways` (create); `GET /v1/mcp_gateways/{id}` (read); `PATCH /v1/mcp_gateways/{id}` (update); `DELETE /v1/mcp_gateways/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_mcp_gateway" "example" {
  name = "tf-mcp-gateway"
}
```

## Schema

### Required

- `name` (String) The `name` field.

### Optional

- `upstream_url` (String) The single upstream MCP server; omit it for a gateway made only of upstreams (see /upstreams).
- `packs_enforced` (Boolean) A caller sees only the tools of packs assigned to them.
- `registry_only` (Boolean) Upstreams must be approved in the registry.
- `guard` (String, JSON) Guardrails: injection (block prompt-injection in results), pii (email, card, ssn, phone: redacted in results, refused in arguments), block_patterns, pin_tools (pin tool definitions and block changed ones), max_result_chars. A JSON document (use `jsonencode`).
- `user_auth` (String, JSON) End-user OAuth: enabled, allow_agents, groups (group name to tool list; * = everyone), required_scopes. Turns the endpoint into an RFC 9728 protected resource. A JSON document (use `jsonencode`).
- `upstream_authorization` (String, Sensitive) The `upstream_authorization` field. Write-only: sent to Atlas, never returned, and not detected as drift.
- `default_policy` (String) The `default_policy` field. One of: `allow`, `deny`.
- `enabled` (Boolean) The `enabled` field.
- `abac` (String, JSON) Authorize every tool call through the ABAC policy engine as well as the tool policy. The principal is the agent (type Agent, with name, tier, owner_user_id, organization_id, scopes, data), the action is the tool name, and the resource is the tool (type McpTool, with gateway_id, gateway_name, tool and args, the call arguments). `policy_ids` limits which policies are consulted (empty: every enabled one). With `enabled` and no permitting policy, every call is denied. A JSON document (use `jsonencode`).
- `tools` (String, JSON) Tool name to { allow?, scopes?, conditions?: [{ arg, op: eq|neq|lte|gte|in|not_in|present, value }], approval?, mask?: { fields?, presets?: email|card|ssn|phone } }. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `has_upstream_authorization` (Boolean) The `has_upstream_authorization` field.
- `resource` (String) The `resource` field.
- `resource_metadata_url` (String) The `resource_metadata_url` field.
- `url` (String) The `url` field.
- `created_at` (Number) The `created_at` field.

## Import

The import id is `<id>`.

```sh
terraform import atlas_mcp_gateway.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
