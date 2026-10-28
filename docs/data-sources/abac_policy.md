# atlas_abac_policy (data source)

Reads an existing abac policy by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_abac_policy" "example" {
  id = "<id>"
}
```
