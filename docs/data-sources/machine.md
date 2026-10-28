# atlas_machine (data source)

Reads an existing machine by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_machine" "example" {
  id = "<id>"
}
```
