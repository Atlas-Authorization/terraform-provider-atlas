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

## `cmp` block

An optional consent manager (any CMP) the hosted sign-in pages load when the instance has no `cookie_consent` of its own. Omit it to bring none; removing it clears it.

- `script_url` (required): the CMP script (https).
- `script_id`, `site_id`, `blocking_mode`, `settings_preset` (`cookiemunch`, `cookiebot`, `onetrust` or `custom`): optional.
- `extra_origins`: extra https origins the CMP calls (at most 10; stored as bare origins).
- `consent_api` (nested block): the consent API the CMP exposes on the page. `global` (required) is the `window` property holding it; `has_method` defaults to `has`; `request_method` is the method that asks the visitor (unset: Atlas never asks); `category_map` maps Atlas category names to the CMP's own.

The preset integrations (`atlas_chatly_integration`, `atlas_forumulate_integration`, `atlas_cookiemunch_integration`, `atlas_referalli_integration`) take the same block. The Cookie Munch preset has a derived default that is not shown in state unless you configure the block.
```hcl
resource "atlas_integration" "consent" {
  integration = "my-cmp"
  name        = "My CMP"
  role        = "consent"
  audience    = "my-cmp:prod"

  cmp {
    script_url = "https://cmp.example.com/banner.js"
    site_id    = "site-123"

    consent_api {
      global         = "__cmp"
      request_method = "request"
      category_map   = { analytics = "statistics" }
    }
  }
}
```
