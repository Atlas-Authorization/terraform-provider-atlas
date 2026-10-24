# atlas_integration

A generic identity integration (preset id or custom id). Presets also have atlas_chatly_integration, atlas_forumulate_integration, atlas_cookiemunch_integration and atlas_referalli_integration.

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_integration" "billing" {
  integration = "billing-app"
  name        = "Billing"
  audience    = "billing:prod"
  events_url  = "https://billing.example.com/atlas/events"
}
```
