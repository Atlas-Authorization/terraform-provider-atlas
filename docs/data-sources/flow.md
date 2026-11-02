# atlas_flow (data source)

Reads an existing flow by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_flow" "example" {
  id = "<id>"
}
```
