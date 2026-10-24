# atlas_referalli_integration

The Referalli identity integration (a preset of atlas_integration).

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_referalli_integration" "referrals" {
  audience = "referalli:org_123"
}
```
