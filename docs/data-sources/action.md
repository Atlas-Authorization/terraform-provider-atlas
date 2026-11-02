# atlas_action (data source)

Reads an existing action by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_action" "example" {
  id = "<id>"
}
```
