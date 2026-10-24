# atlas_status_subscription

Which Atlas status incidents concern this instance (singleton).

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_status_subscription" "this" {
  min_impact = "major"
  regions    = ["eu"]
}
```
