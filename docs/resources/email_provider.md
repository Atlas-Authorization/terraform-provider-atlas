# atlas_email_provider

Your own email provider (bring your own account), so Atlas sends sign-in codes,
magic links and notifications through it instead of the shared default sender.

Set **exactly one** transport block: `resend`, `ses`, `smtp`, `brevo` or
`mailgun`. Secret settings (API keys, passwords) are **sensitive and
write-only**: Atlas stores them encrypted and never returns them, so they are
kept from your configuration; `secrets_set` reports which ones Atlas holds.
Non-secret settings are read back, so drift shows in the plan.

Removing an optional non-secret setting from a block keeps the stored value;
switch transport (or destroy and re-create) to reset it. Destroying the
resource returns the channel to the Atlas default sender.

API: `GET /v1/messaging_providers`, `PUT /v1/messaging_providers/email`,
`DELETE /v1/messaging_providers/email`.

## Example Usage

```hcl
resource "atlas_email_provider" "this" {
  from_address = "Acme <no-reply@acme.example.com>"

  resend = {
    api_key = var.resend_api_key
  }
}

# Alternatively, SMTP (one atlas_email_provider per instance):
# resource "atlas_email_provider" "this" {
resource "atlas_email_provider" "smtp_example" {
  from_address = "no-reply@acme.example.com"

  smtp = {
    host     = "smtp.acme.example.com"
    port     = 587
    username = "atlas"
    password = var.smtp_password
  }
}
```

Each transport is a nested attribute, written `resend = { ... }`.

## Schema

### Optional

- `from_address` (String) — The sender address, e.g. `Acme <no-reply@acme.com>`.
- `resend` (Attributes) — `api_key` (String, Sensitive, required).
- `ses` (Attributes) — `region` (required), `access_key_id` (Sensitive, required), `secret_access_key` (Sensitive, required), `session_token` (Sensitive).
- `smtp` (Attributes) — `host` (required), `port` (Number, default 587), `secure` (Boolean; default true on port 465, otherwise STARTTLS), `username`, `password` (Sensitive).
- `brevo` (Attributes) — `api_key` (String, Sensitive, required).
- `mailgun` (Attributes) — `api_key` (Sensitive, required), `domain` (required), `region` (`us` default, or `eu`).

### Read-Only

- `id` (String) — Always `email`.
- `transport` (String) — The configured transport.
- `status` (String) — Provider status.
- `secrets_set` (Set of String) — Which secret settings Atlas holds (never their values).
- `updated_at` (Number) — Last update time (epoch ms).

## Import

```sh
terraform import atlas_email_provider.this email
```

Secrets are not recovered on import; the next apply re-sends them from your
configuration.
