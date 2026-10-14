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

### Read-Only

- `id` (String) — Always `cookie_consent`.

## Import

```sh
terraform import atlas_cookie_consent.this cookie_consent
```
