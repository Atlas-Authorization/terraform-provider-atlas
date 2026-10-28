# atlas_mcp_gateway (data source)

Reads an existing mcp gateway by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_mcp_gateway" "example" {
  id = "<id>"
}
```
