# atlas_redirect_url

An allowlisted redirect URL — one of the exact URLs an OAuth / SSO flow may hand
control back to. Managing the allowlist as code keeps the set of valid redirect
targets under review-and-apply.

The Backend API **normalises** the stored value on create (a bare
`scheme://host[:port]`, dropping a trailing `/` path) and has **no update
route**, so `url` is immutable — changing it forces replacement. Supply an
already-normalised absolute http(s) URL to avoid a perpetual diff.

Backend API (BAPI): `POST /v1/redirect_urls` (create),
`GET /v1/redirect_urls/:id` (read), `DELETE /v1/redirect_urls/:id` (delete).

## Example Usage

```hcl
resource "atlas_redirect_url" "app_callback" {
  url = "https://app.acme.example.com/callback"
}
```

## Schema

### Required

- `url` (String) — Absolute http(s) URL to allowlist. Immutable — changing it
  forces replacement.

### Read-Only

- `id` (String) — Atlas object id.
- `created_at` (Number) — Creation time (epoch ms).

## Import

```sh
terraform import atlas_redirect_url.app_callback rurl_123
```
