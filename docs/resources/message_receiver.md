# atlas_message_receiver

Store the secret Atlas verifies this provider's callbacks with (your Mailgun webhook signing key, SendGrid public key, Resend webhook secret, Vonage signature secret, MessageBird signing key, or the SNS topic ARNs to accept). Write-only: it is never returned. Twilio needs none (it uses the auth token of your SMS provider).

Scopes: `messaging:write` to read, `messaging:write` to write.

Backend API: `PUT /v1/message_receivers/{provider}` (create and update); no read route: refresh keeps the last applied values; `DELETE /v1/message_receivers/{provider}` (destroy).

## Example Usage

```hcl
resource "atlas_message_receiver" "example" {
  provider = "tf-message-receiver"
  secret = "tf-message-receiver"
}
```

## Schema

### Required

- `provider` (String) The provider this object belongs to. Changing it forces replacement.
- `secret` (String, Sensitive) The `secret` field. Write-only: sent to Atlas, never returned, and not detected as drift.

### Read-Only

- `id` (String) Atlas object id.
- `channel` (String) The `channel` field. One of: `email`, `sms`.
- `verification` (String) `signature`: the provider signs its callbacks and a verification secret is required. `signature_or_token`: signed when a secret is set, else the URL token. `token`: the provider cannot sign; the URL carries a token derived for this instance.
- `url` (String) The URL to register on your provider account. For a token provider it carries the token, when your key may see it.
- `token_required` (Boolean, Sensitive) The `token_required` field. Returned once, on create.
- `token` (String, Sensitive) The token, for a key with `messaging:write`. Present it as a bearer header, as the basic-auth password or as `?t=`. Returned once, on create.
- `secret_set` (Boolean, Sensitive) The `secret_set` field. Returned once, on create.
- `secret_help` (String, Sensitive) What the verification secret is, or null when none is needed. Returned once, on create.
- `events` (String) The `events` field.
- `last_event_at` (Number) The `last_event_at` field. (epoch ms)
- `last_error` (String) Why the last callback was refused.

## Import

The import id is `<provider>`.

```sh
terraform import atlas_message_receiver.example <provider>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
