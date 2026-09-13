# atlas_domain

A custom instance domain (the frontend-api `fapi` host or the `accounts` host).
Point a CNAME at the returned `cname_target`, then verify out of band. The
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
```

## Schema

### Required

- `role` (String) — Domain role: `fapi` or `accounts`. Immutable — forces replacement.
- `host` (String) — The hostname to serve (e.g. `auth.example.com`). Immutable — forces replacement.

### Read-Only

- `id` (String) — Atlas object id.
- `status` (String) — Provisioning status (`pending`, `active`, ...).
- `cname_target` (String) — The CNAME target to point the host at.
- `live` (Boolean) — Whether the domain is fully live (DNS + certificate + cookie checks pass).
- `cookie_domain` (String) — The registrable cookie domain derived from the host, if resolvable.
- `failure_reason` (String) — Most recent verification failure reason, if any.
- `certificate_expires_at` (Number) — Certificate expiry (epoch ms), once issued.

## Import

```sh
terraform import atlas_domain.auth dom_123
```
