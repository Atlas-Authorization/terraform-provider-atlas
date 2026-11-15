# atlas_billing_connect_account

Create the Stripe Connect (Express) account a user or organization is paid out to, on YOUR Stripe account. One account per subject: a second call returns it (200). Idempotent on `idempotency_key`. Send the person to a link from `POST /v1/billing/connect/accounts/{id}/onboarding_link` to finish Stripe's onboarding.

Scopes: `billing:read` to read, `billing:write` to write.

Backend API: `POST /v1/billing/connect/accounts` (create); `GET /v1/billing/connect/accounts/{id}` (read); no update route: every change replaces the object; no delete route: destroy forgets the object.

## Example Usage

```hcl
resource "atlas_billing_connect_account" "example" {
  subject_type = "user"
  subject_id = "tf-billing-connect-account"
  idempotency_key = "tf-billing-connect-account"
}
```

## Schema

### Required

- `subject_type` (String) The `subject_type` field. One of: `user`, `org`.
- `subject_id` (String) The `subject_id` field.
- `idempotency_key` (String) The `idempotency_key` field. Not returned by the API, so it is not detected as drift.

### Optional

- `email` (String) Prefills the Stripe onboarding. Not returned by the API, so it is not detected as drift.
- `country` (String) Two-letter ISO country, upper case. It cannot be changed later.

### Read-Only

- `id` (String) Atlas object id.
- `provider_account_id` (String) The Stripe account id (acct_...).
- `details_submitted` (Boolean) The `details_submitted` field.
- `charges_enabled` (Boolean) The `charges_enabled` field.
- `payouts_enabled` (Boolean) True once Stripe accepts transfers to the account.
- `requirements_due` (List of String) What Stripe still needs from the account holder.
- `refreshed_at` (Number) The `refreshed_at` field. (epoch ms)
- `created_at` (Number) Epoch milliseconds. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_billing_connect_account.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
