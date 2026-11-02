# atlas_fga_store (data source)

Reads an existing fga store by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_fga_store" "example" {
  id = "<id>"
}
```
