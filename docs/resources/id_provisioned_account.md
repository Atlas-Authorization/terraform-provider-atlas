# atlas_id_provisioned_account

A **test or review account** for a Sign in with Atlas app: a real Atlas account (it signs in on `atlasauth.net` with its email and password) that is **confined to the one app** it was created for. Every other app, the Atlas dashboard and Atlas account settings refuse it. Use it for directory and app-store reviewer logins, end-to-end tests in CI, demos and partner testing.

Needs the provider's `platform_key` with the scope `platform:atlas_id_apps:provision`, or `atlas_id_apps:<app id>:provision` to provision for one app only.

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_id_provisioned_account" "reviewer" {
  app_id = atlas_id_app.product.id
  email  = "reviewer@product.example"
  label  = "app-store reviewer"

  two_factor_exempt = true

  password_wo         = var.reviewer_password
  password_wo_version = 1
}
```

## The password

`password_wo` (at least 12 characters) is **write-only**: it is sent to Atlas on create and whenever `password_wo_version` changes, and it is never stored in state. It needs Terraform 1.11 or later. Leave it out and Atlas generates a password, which lands in the sensitive `generated_password` (Atlas returns it once, so treat the state as sensitive). Changing `password_wo_version` rotates the password in place; existing sessions stay signed in until the account is removed.

## Guardrails

- The account works with `app_id` only. Everywhere else it is refused.
- `two_factor_exempt = true` skips the second factor for this app. Anyone who holds the password can sign in to the app as this account, so use it for reviewer logins only and share the password with the review team only.
- Omit `expires_at` for a standing reviewer login; set it (RFC 3339) for CI or a demo and Atlas removes the account by itself.
- Each account holds one email address that no other Atlas account uses. With `provisioned_verified_domains_only` on the app, the address must be on a domain you verified.
- An account may hold up to 25 test accounts across all its apps; creates and password rotations are rate limited.
- Everything except the password is fixed at creation; changing it replaces the account. Destroying the resource removes the account and ends its sessions and grants.

## Import

```
terraform import atlas_id_provisioned_account.reviewer oac_123/iacct_456
```

The password cannot be imported; set `password_wo` and `password_wo_version` to rotate in a known one.
