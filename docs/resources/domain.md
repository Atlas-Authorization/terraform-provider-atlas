# atlas_domain

A custom instance domain (the frontend-api `fapi` host or the `accounts` host).
Point a CNAME at the returned `cname_target` — or, with `method = "path"`, mount
the Frontend API under `base_path` on a host you already run and serve
`verification_content` at `verification_url` (no DNS) — then verify out of band. The
Backend API has **no update route**, so a change to `role` or `host` replaces the
domain.

Backend API (BAPI): `POST /v1/domains` (create), `GET /v1/domains` (list — the
read filters this list by id), `DELETE /v1/domains/:id` (delete).

## Example Usage

```hcl
resource "atlas_domain" "auth" {
  role = "fapi"
  host = "auth.example.com"
}

output "cname_target" {
  value = atlas_domain.auth.cname_target
}

# Path mount: the Frontend API under /__atlas on your own host, no DNS record.
resource "atlas_domain" "app" {
  role      = "fapi"
  host      = "app.example.com"
  method    = "path"
  base_path = "/__atlas"
}

output "verification_file" {
  value = {
    url  = atlas_domain.app.verification_url
    body = atlas_domain.app.verification_content
  }
}
```

## Schema

### Required

- `role` (String) — Domain role: `fapi` or `accounts`. Immutable — forces replacement.
- `host` (String) — The hostname to serve (e.g. `auth.example.com`). Immutable — forces replacement.

### Optional

- `method` (String) — `cname` (default; a CNAME to Atlas) or `path` (the Frontend API served under `base_path` on your own host, proven by an https file). Immutable — forces replacement.
- `base_path` (String) — `method = "path"` only: the prefix the Frontend API is served under (default `/__atlas`). Immutable — forces replacement.

### Read-Only

- `id` (String) — Atlas object id.
- `status` (String) — Provisioning status (`pending`, `active`, ...).
- `cname_target` (String) — The CNAME target to point the host at (null for a path mount).
- `live` (Boolean) — Whether the domain is fully live (DNS + certificate + cookie checks pass).
- `cookie_domain` (String) — The registrable cookie domain derived from the host, if resolvable.
- `failure_reason` (String) — Most recent verification failure reason, if any.
- `certificate_expires_at` (Number) — Certificate expiry (epoch ms), once issued.
- `routing` (String) — Path mount, once verified: `atlas_ingress` (Atlas routes the prefix) or `customer_proxy` (add the reverse-proxy rule yourself).
- `frontend_api_url` (String) — Path mount: the Frontend API URL (and, once verified, the issuer), `https://<host><base_path>`.
- `verification_url` (String) — Path mount: the https URL to serve the verification file at.
- `verification_token` (String) — Path mount: the verification token, bound to this instance, host and base path.
- `verification_content` (String) — Path mount: the exact body to serve at `verification_url`.

## Import

```sh
terraform import atlas_domain.auth dom_123
```
