# atlas_billing_coupon

Create a discount code. Set exactly one of `percent_off` or `amount_off` (+ `currency`). It is created on your own Stripe account as a coupon and promotion code, so Stripe enforces and applies it; a provider without discount support refuses with 422.

Scopes: `billing:read` to read, `billing:write` to write.

Backend API: `POST /v1/billing/coupons` (create); `GET /v1/billing/coupons/{id}` (read); `PATCH /v1/billing/coupons/{id}` (update); `DELETE /v1/billing/coupons/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_billing_coupon" "example" {

}
```

## Schema

### Optional

- `code` (String) The `code` field. Changing it forces replacement.
- `name` (String) The `name` field.
- `percent_off` (Number) The `percent_off` field. Changing it forces replacement.
- `amount_off` (Number) The `amount_off` field. Changing it forces replacement.
- `currency` (String) The `currency` field. Changing it forces replacement.
- `duration` (String) The `duration` field. One of: `once`, `repeating`, `forever`. Changing it forces replacement.
- `duration_months` (Number) The `duration_months` field. Changing it forces replacement.
- `max_redemptions` (Number) The `max_redemptions` field.
- `expires_at` (Number) The `expires_at` field. (epoch ms) Changing it forces replacement.
- `plan_ids` (List of String) The `plan_ids` field. Changing it forces replacement.
- `first_time_only` (Boolean) The `first_time_only` field. Changing it forces replacement.
- `provider_value` (String) The `provider` field. Not returned by the API, so it is not detected as drift. Changing it forces replacement.

### Read-Only

- `id` (String) Atlas object id.
- `redeemed_count` (Number) The `redeemed_count` field.
- `active` (Boolean) The `active` field.
- `providers_value` (List of String) The `providers` field.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_billing_coupon.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
