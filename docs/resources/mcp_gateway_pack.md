# atlas_mcp_gateway_pack

Create a tool pack. With the gateway's `packs_enforced` on, a caller sees and may call only the tools of packs assigned to them.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `POST /v1/mcp_gateways/{id}/packs` (create); `GET /v1/mcp_gateways/{id}/packs/{packId}` (read); `PATCH /v1/mcp_gateways/{id}/packs/{packId}` (update); `DELETE /v1/mcp_gateways/{id}/packs/{packId}` (destroy).

## Example Usage

```hcl
resource "atlas_mcp_gateway_pack" "example" {
  mcp_gateway_id = "tf-mcp-gateway-pack"
}
```

## Schema

### Required

- `mcp_gateway_id` (String) The mcp gateway id this object belongs to. Changing it forces replacement.

### Optional

- `name` (String) The `name` field.
- `description` (String) The `description` field.
- `tools` (List of String) The `tools` field.
- `agent_ids` (List of String) Agents (machine ids) that get these tools.
- `org_ids` (List of String) Members of these organizations get these tools.
- `groups` (List of String) Members of these groups (id or slug) get these tools.

### Read-Only

- `id` (String) Atlas object id.
- `gateway_id` (String) The `gateway_id` field.
- `created_at` (Number) The `created_at` field. (epoch ms)

## Import

The import id is `<mcp_gateway_id>/<id>`.

```sh
terraform import atlas_mcp_gateway_pack.example <mcp_gateway_id>/<id>
```
