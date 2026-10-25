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

## `per_call_audience` block

Lets a token mint name its own audience per call (`getToken({ for, cbid })`, `?audience=`, or `?<param>=`), so one integration can serve many values (for example one audience per site). A preset already allows its own parameter for the configured value only (chatly `channel`, forumulate and referalli `org_id`, cookiemunch `cbid`); set the block to allow more values. Omit it to keep the preset's default; set `per_call_audience_disabled = true` to turn it off (cannot be combined with the block).

- `param`: the query parameter that carries the value (a preset has its own).
- `template`: custom integrations only (required there): the audience with `{param}` in it once, such as `acme:{tenant}`.
- `allowed`: values of `param` a call may name besides the configured one (at most 50).
- `any_matching`: allow any value that fits `audience_format` (a custom integration needs `audience_format`).

The preset integrations take the same block. A preset's default is not shown in state unless you configure the block; the `atlas_integration` data source always shows what is in effect.

```hcl
resource "atlas_integration" "desk" {
  integration = "desk"
  name        = "Desk"
  audience    = "desk:main"

  per_call_audience {
    param    = "tenant"
    template = "desk:{tenant}"
    allowed  = ["main", "eu", "us"]
  }
}
```
