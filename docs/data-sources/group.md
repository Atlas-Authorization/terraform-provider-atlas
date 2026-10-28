# atlas_group (data source)

Reads an existing group by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_group" "example" {
  id = "<id>"
}
```
