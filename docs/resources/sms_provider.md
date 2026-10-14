# atlas_sms_provider

Your own SMS provider (bring your own account), so Atlas sends phone codes
through it instead of the shared default sender.

Set **exactly one** transport block (a nested attribute, `twilio = { ... }`):
`twilio`, `vonage`, `messagebird`, `plivo`, `telnyx`, `sinch`, `infobip`,
`clicksend`, `bandwidth` or `aws_sns`. Secret settings are **sensitive and
write-only**: Atlas stores them encrypted and never returns them, so they are
kept from your configuration; `secrets_set` reports which ones Atlas holds.
Non-secret settings are read back. Destroying the resource returns the channel
to the Atlas default sender.

API: `GET /v1/messaging_providers`, `PUT /v1/messaging_providers/sms`,
`DELETE /v1/messaging_providers/sms`.

## Example Usage

```hcl
resource "atlas_sms_provider" "this" {
  twilio = {
    account_sid = var.twilio_account_sid
    auth_token  = var.twilio_auth_token
    from        = "+15555550100"
  }
}
```

## Schema

### Optional

- `from_address` (String) — The sender number or ID (takes precedence over a transport's `from`).
- `twilio` (Attributes) — `account_sid` (required), `auth_token` (Sensitive, required), `from`, `whatsapp_from`, `voice_from`.
- `vonage` (Attributes) — `api_key` (required), `api_secret` (Sensitive, required), `from`.
- `messagebird` (Attributes) — `access_key` (Sensitive, required), `from`.
- `plivo` (Attributes) — `auth_id` (required), `auth_token` (Sensitive, required), `from`.
- `telnyx` (Attributes) — `api_key` (Sensitive, required), `from`.
- `sinch` (Attributes) — `service_plan_id` (required), `api_token` (Sensitive, required), `region` (e.g. `us`, `eu`), `from`.
- `infobip` (Attributes) — `base_url` (https, required), `api_key` (Sensitive, required), `from`.
- `clicksend` (Attributes) — `username` (required), `api_key` (Sensitive, required), `from`.
- `bandwidth` (Attributes) — `account_id`, `application_id`, `username` (all required), `password` (Sensitive, required), `from`.
- `aws_sns` (Attributes) — `region`, `access_key_id` (required), `secret_access_key` (Sensitive, required), `session_token` (Sensitive), `from`.

### Read-Only

- `id` (String) — Always `sms`.
- `transport` (String) — The configured transport.
- `status` (String) — Provider status.
- `secrets_set` (Set of String) — Which secret settings Atlas holds (never their values).
- `updated_at` (Number) — Last update time (epoch ms).

## Import

```sh
terraform import atlas_sms_provider.this sms
```
