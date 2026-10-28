# atlas_access_policy (data source)

Reads an existing access policy by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_access_policy" "example" {
  id = "<id>"
}
```
