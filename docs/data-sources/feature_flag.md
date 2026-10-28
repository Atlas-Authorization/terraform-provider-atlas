# atlas_feature_flag (data source)

Reads an existing feature flag by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_feature_flag" "example" {
  id = "<id>"
}
```
