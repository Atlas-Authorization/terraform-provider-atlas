# atlas_cookiemunch_integration

The Cookie Munch identity integration (a preset of atlas_integration).

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_cookiemunch_integration" "consent" {
  audience = "cookiemunch:dv-abc"
}
```
