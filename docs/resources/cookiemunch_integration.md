# atlas_cookiemunch_integration

The Cookie Munch identity integration (a preset of atlas_integration).

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_cookiemunch_integration" "consent" {
  audience = "cookiemunch:dv-abc"
}
```

## `cmp` block

An optional consent manager (any CMP) the hosted sign-in pages load when the instance has no `cookie_consent` of its own. Omit it to bring none; removing it clears it.

- `script_url` (required): the CMP script (https).
- `script_id`, `site_id`, `blocking_mode`, `settings_preset` (`cookiemunch`, `cookiebot`, `onetrust` or `custom`): optional.
- `extra_origins`: extra https origins the CMP calls (at most 10; stored as bare origins).
- `consent_api` (nested block): the consent API the CMP exposes on the page. `global` (required) is the `window` property holding it; `has_method` defaults to `has`; `request_method` is the method that asks the visitor (unset: Atlas never asks); `category_map` maps Atlas category names to the CMP's own.

Same block as `atlas_integration`. The preset's derived default is not shown in state unless you configure the block.
```hcl
resource "atlas_cookiemunch_integration" "consent" {
  audience    = "cookiemunch:example"

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
