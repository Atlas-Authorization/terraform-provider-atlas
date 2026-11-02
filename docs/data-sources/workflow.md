# atlas_workflow (data source)

Reads an existing workflow by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_workflow" "example" {
  id = "<id>"
}
```
