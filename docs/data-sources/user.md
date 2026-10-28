# atlas_user (data source)

A user in the instance, looked up by id.

Every attribute except `id` is computed. Scope: `users:read`.

## Example Usage

```hcl
data "atlas_user" "example" {
  id = "<id>"
}
```

## Schema

### Required

- `id` (String) The id to look up.

### Read-Only

- `username` (String) The `username` field.
- `first_name` (String) The `first_name` field.
- `last_name` (String) The `last_name` field.
- `image_url` (String) The `image_url` field.
- `email_addresses` (String, JSON) The `email_addresses` field. A JSON document (use `jsonencode`).
- `primary_email_address_id` (String) The `primary_email_address_id` field.
- `primary_email` (String) The primary address as a flat string; null until a verified primary exists.
- `external_id` (String) The relying party's own id for this user, if set/imported.
- `password_enabled` (Boolean, Sensitive) Whether the account has a usable password (never the hash).
- `public_metadata` (String, JSON) The `public_metadata` field. A JSON document (use `jsonencode`).
- `mfa_enabled` (Boolean) The `mfa_enabled` field.
- `required_factors` (List of String) Factor classes this user must always use (POL-04); empty = none.
- `is_anonymous` (Boolean) Whether this identity is still an anonymous guest.
- `banned` (Boolean) The `banned` field.
- `ban_reason` (String) The `ban_reason` field.
- `ban_note` (String) The `ban_note` field.
- `locked` (Boolean) The `locked` field.
- `deleted_at` (Number) Epoch-ms; non-null once soft-deleted.
- `last_sign_in_at` (Number) Epoch-ms.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).
- `version` (Number) Strictly increases with every change to the user. Webhook payloads carry the same value: apply an update only if its version is higher than the one you hold.
- `referral` (String, JSON) How the user arrived through a referral program (referral attribution on). Null only when the user was never reported to the referral platform; otherwise `status` says where attribution stands (`pending` until the platform answers). Present on read, list, create and update responses. A JSON document (use `jsonencode`).
- `referral_code` (String) The user's own referral code, when they are a referrer.
- `referral_link` (String) The user's own referral link, when they are a referrer.
