# atlas_mcp_gateway (data source)

An MCP tool gateway, looked up by id.

Every attribute except `id` is computed. Scope: `machines:read`.

## Example Usage

```hcl
data "atlas_mcp_gateway" "example" {
  id = "<id>"
}
```

## Schema

### Required

- `id` (String) The id to look up.

### Read-Only


