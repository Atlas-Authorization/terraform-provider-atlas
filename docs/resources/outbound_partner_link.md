# atlas_outbound_partner_link

The credential a partner granted this instance, so Atlas can call the partner.

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_outbound_partner_link" "cookiemunch" {
  partner       = "cookiemunch"
  client_id     = var.cm_client_id
  client_secret = var.cm_client_secret
  scopes        = ["dsar:write"]
}
```

## `cmp` block

An optional consent manager (any CMP) the hosted sign-in pages load when the instance has no `cookie_consent` of its own. Omit it to bring none; removing it clears it.

- `script_url` (required): the CMP script (https).
- `script_id`, `site_id`, `blocking_mode`, `settings_preset` (`cookiemunch`, `cookiebot`, `onetrust` or `custom`): optional.
- `extra_origins`: extra https origins the CMP calls (at most 10; stored as bare origins).
- `consent_api` (nested block): the consent API the CMP exposes on the page. `global` (required) is the `window` property holding it; `has_method` defaults to `has`; `request_method` is the method that asks the visitor (unset: Atlas never asks); `category_map` maps Atlas category names to the CMP's own.

A Cookie Munch link with a `site:<id>` `resource` brings its own derived default, which is not shown in state. Removing the block sends `cmp: null`.
```hcl
resource "atlas_outbound_partner_link" "consent" {
  partner       = "consentco"
  base_url      = "https://api.consentco.example"
  client_id     = var.client_id
  client_secret = var.client_secret

  cmp {
    script_url = "https://cmp.consentco.example/banner.js"

    consent_api {
      global = "__consentco"
    }
  }
}
```
