# atlas_device_assurance_policy (data source)

Reads an existing device assurance policy by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

```hcl
data "atlas_device_assurance_policy" "example" {
  id = "<id>"
}
```
