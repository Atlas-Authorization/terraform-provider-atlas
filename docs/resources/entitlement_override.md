# atlas_entitlement_override

A sales-led entitlement override: pins a plan, features and/or limits onto a
user or organization **without** a Stripe subscription — for example an
Enterprise contract closed by sales. While it is live (not expired), the
override wins over the subscription-derived entitlement, from the subject's
next session.

Overrides are grant-only (there is no edit), so changing any argument replaces
the override.

API: `POST /v1/entitlement_overrides`, `GET /v1/entitlement_overrides` (read,
filtered by subject), `DELETE /v1/entitlement_overrides/:id`.

## Example Usage

```hcl
resource "atlas_entitlement_override" "acme_enterprise" {
  subject_type = "org"
  subject_id   = atlas_organization.acme.id
  plan         = "enterprise"
  features     = ["sso", "audit_log_export"]
  limits = {
    seats = 500
    projects = -1 # unlimited
  }
  expires_at = 1830297600000 # 2028-01-01, epoch ms
}
```

## Schema

### Required

- `subject_type` (String) — `user` or `org`.
- `subject_id` (String) — The user or organization id.

At least one of `plan`, `features` or `limits` is required.

### Optional

- `plan` (String) — Plan slug to grant.
- `features` (Set of String) — Feature keys to grant.
- `limits` (Map of Number) — Limit name to number (`-1` = unlimited).
- `expires_at` (Number) — When the override lapses (epoch milliseconds). Unset = never.

### Read-Only

- `id` (String) — Override id.
- `created_at` (Number) — Creation time (epoch ms).

## Import

```sh
terraform import atlas_entitlement_override.acme_enterprise org/org_123/eo_456
# or just the id:
terraform import atlas_entitlement_override.acme_enterprise eo_456
```
