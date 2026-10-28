# atlas_network_zone (data source)

Reads an existing network zone by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_network_zone" "example" {
  id = "<id>"
}
```
