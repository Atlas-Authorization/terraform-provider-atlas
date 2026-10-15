# atlas_cookie_consent

The cookie-consent banner on your hosted sign-in, sign-up and account pages.
Works with any consent-management platform (Cookiebot, CookieMunch, Usercentrics,
…): Atlas injects your CMP script and adds its origins to the page's Content
Security Policy.

A singleton per instance. Each apply replaces the whole setting; destroying the
resource turns the banner off. If the banner is turned off outside Terraform,
the next plan re-creates it.

API: `GET`/`PUT`/`DELETE /v1/instance/cookie_consent`.

## Example Usage

```hcl
resource "atlas_cookie_consent" "this" {
  script_url    = "https://consent.cookiebot.com/uc.js"
  script_id     = "Cookiebot"
  site_id       = "00000000-0000-0000-0000-000000000000"
  blocking_mode = "auto"
  extra_origins = ["https://consentcdn.cookiebot.com"]

  # A "Cookie settings" link in the page footer that reopens the CMP.
  settings_link = {
    label  = "Cookie settings"
    preset = "cookiebot"
  }
}
```

With a CMP that listens for its own attributes, use the `custom` preset (or
leave `preset` unset) and set `attributes`:

```hcl
  settings_link = {
    label      = "Privacy choices"
    attributes = { "data-cc" = "show-settings", "aria-label" = "Open cookie settings" }
  }
```

## Schema

### Required

- `script_url` (String) — Your CMP script URL. Must be https; its origin is added to the page CSP.

### Optional

- `script_id` (String) — The `<script id>` some CMPs need.
- `site_id` (String) — Your CMP site id (sent as `data-cbid`).
- `blocking_mode` (String) — The CMP prior-blocking mode, e.g. `auto` (sent as `data-blockingmode`).
- `extra_origins` (List of String) — Other https URLs the CMP loads from or calls; each origin is added to the page CSP.
- `settings_link` (Attributes) — A **Cookie settings** link in the footer of the hosted pages that reopens your CMP's preferences, so a visitor can change or withdraw consent after the banner is gone. Unset = no link. See [below for nested schema](#nestedatt--settings_link).

### Read-Only

- `id` (String) — Always `cookie_consent`.

<a id="nestedatt--settings_link"></a>
### Nested Schema for `settings_link`

Required:

- `label` (String) — The link text, 1–60 characters, e.g. `Cookie settings`.

Optional:

- `preset` (String) — The CMP's reopen trigger: `cookiemunch` (adds `data-cc="show-settings"`), `cookiebot` (calls `Cookiebot.renew()`), `onetrust` (calls `OneTrust.ToggleInfoDisplay()`) or `custom` (adds nothing; use `attributes`).
- `attributes` (Map of String) — Extra HTML attributes your CMP's click handler listens for. Allowed names: `data-*`, `aria-*`, `id`, `class` and `href` (an https URL or a `#fragment`), lower case; at most 10, values up to 200 characters. `data-atlas-*` and event handlers (`on*`) are refused. Set a `preset`, `attributes`, or both.

Read-Only:

- `rendered_attributes` (Map of String) — What Atlas renders on the link: `attributes` with the preset's trigger applied.

The provider checks these rules at plan time.

## Import

```sh
terraform import atlas_cookie_consent.this cookie_consent
```
