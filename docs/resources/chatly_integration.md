# atlas_chatly_integration

The Chatly identity integration (a preset of atlas_integration).

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_chatly_integration" "support" {
  audience = "chatly:support"
}
```
