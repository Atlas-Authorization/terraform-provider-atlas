# atlas_custom_oauth_provider (data source)

Reads an existing custom oauth provider by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_custom_oauth_provider" "example" {
  id = "<id>"
}
```
