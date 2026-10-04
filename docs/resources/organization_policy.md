# atlas_organization_policy

The security policy for one organization. Each field is an **optional
tightening** on top of the instance policy, merged server-side; a field left
unset inherits the instance behaviour (and keeps whatever was last applied), and
setting a value enforces it.

Atlas exposes no read-only GET for org policy (the org projection omits it), so
the provider reads the current policy via an idempotent empty merge. A refresh
therefore records an `organization.policy.updated` audit entry even when nothing
changed.

Backend API (BAPI): `PATCH /v1/organizations/:id/policy` (create / read /
update). The read is an empty-body PATCH that merges nothing and returns the
resolved policy.

## Example Usage

```hcl
resource "atlas_organization_policy" "acme" {
  organization_id         = atlas_organization.acme.id
  require_mfa             = true
  sso_required            = true
  allowed_sign_in_methods = ["password", "oauth_google", "saml"]
  ip_allowlist            = ["203.0.113.0/24"]
  max_session_age_seconds = 3600
}
```

## Schema

### Required

- `organization_id` (String) — The organization whose policy this manages.
  Immutable — changing it forces replacement.

### Optional

- `require_mfa` (Boolean) — Require an MFA-verified session to hold this org
  active.
- `sso_required` (Boolean) — Force enterprise SSO for identifiers whose email
  domain maps to this org (requires a verified org domain to take effect).
- `session_idle_override_ms` (Number) — Override the session idle window while
  this org is active, in milliseconds.
- `allowed_sign_in_methods` (Set of String) — Restrict which sign-in strategies
  a member may use. Each entry is a family (`password`, `oauth`, `saml`, …) or a
  specific provider (`oauth_google`). Empty/unset = every method.
- `ip_allowlist` (Set of String) — CIDR / IP allow-list gating sign-in for
  members of this org. Empty/unset = unrestricted.
- `max_session_age_seconds` (Number) — Cap the absolute session lifetime for
  this org's members, in seconds (300 – 7776000). An org may only shorten, never
  extend.

### Read-Only

- `id` (String) — Equals `organization_id`.

## Import

```sh
terraform import atlas_organization_policy.acme org_123
```
