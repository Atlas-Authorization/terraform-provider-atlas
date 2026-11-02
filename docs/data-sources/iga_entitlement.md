# atlas_iga_entitlement (data source)

Reads an existing iga entitlement by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_iga_entitlement" "example" {
  id = "<id>"
}
```
