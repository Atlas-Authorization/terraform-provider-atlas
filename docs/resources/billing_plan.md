# atlas_billing_plan

A billing plan a tenant defines for their app's users (or orgs). The plan's
`slug` becomes the `pla` session claim and its `features` the `fea` claim, so
entitlement gating works off the plan a subject is on. A plan with no
`stripe_price_id` is the **free tier** (`free` is then `true`). A subscription's
status is written only by the verified Stripe webhook, never by Terraform.

Backend API (BAPI): `POST /v1/billing/plans` (create),
`GET /v1/billing/plans/:id` (read), `PATCH /v1/billing/plans/:id` (update),
`DELETE /v1/billing/plans/:id` (delete).

## Example Usage

```hcl
resource "atlas_billing_plan" "pro" {
  name            = "Pro"
  slug            = "pro"
  stripe_price_id = "price_123"
  audience        = "org"
  interval        = "month"
  currency        = "usd"
  features        = ["seats", "sso", "audit_log"]
  limits = {
    max_seats   = 25
    max_devices = 5
  }
  pricing_model = "per_seat"
  trial_days    = 14
}

# The free tier: a plan with no Stripe price.
resource "atlas_billing_plan" "free" {
  name     = "Free"
  slug     = "free"
  features = ["dashboard"]
}
```

## Schema

### Required

- `name` (String) — Human-readable plan name.
- `slug` (String) — URL-safe, instance-unique plan key — surfaced as the `pla`
  session claim.

### Optional

- `stripe_price_id` (String) — The recurring Stripe price a subscriber is
  charged. Omit for the free tier.
- `audience` (String) — Who may subscribe: `user` or `org`. Defaults to `user`.
- `interval` (String) — Billing interval: `month` or `year`. Defaults to `month`.
- `amount` (String) — A free-form display string for the price, stored verbatim
  by Atlas — not parsed or converted. Use whatever units you display (e.g.
  `12.00` for $12, or `1200` for cents); Atlas does not interpret it. Pair it
  with `currency` and `interval`.
- `currency` (String) — ISO currency code (e.g. `usd`). Defaults to `usd`.
- `features` (List of String) — Feature keys this plan grants, surfaced as the
  `fea` session claim.
- `limits` (Map of Number) — Numeric caps the plan grants, keyed by limit name
  (e.g. `{ max_seats = 25, max_devices = 5 }`). Where `features` answer "can
  they?", limits answer "how many?"; they are surfaced on the verified token and
  merged under any per-organization override. Unset keeps the plan's current
  limits (none for a new plan); `{}` clears them.
- `active` (Boolean) — Whether the plan is offered. Defaults to `true`.
- `pricing_model` (String) — `flat`, `per_seat` or `metered`. Defaults to `flat`.
- `trial_days` (Number) — Free trial length in days (0–3650), or omitted for none.
- `stripe_meter_id` (String) — Stripe meter id, for a `metered` plan.
- `usage_unit` (String) — Usage unit label, for a `metered` plan.

### Read-Only

- `id` (String) — Atlas object id.
- `free` (Boolean) — True when the plan has no `stripe_price_id`.
- `created_at` (Number) — Creation time (epoch ms).
- `updated_at` (Number) — Last update time (epoch ms).

## Import

```sh
terraform import atlas_billing_plan.pro plan_123
```
