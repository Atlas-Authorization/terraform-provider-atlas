# atlas_forumulate_integration

The Forumulate identity integration (a preset of atlas_integration).

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_forumulate_integration" "docs" {
  audience = "forumulate:org_123"
}
```
