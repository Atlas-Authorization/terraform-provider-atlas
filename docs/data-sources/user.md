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
- `status` (String) One account state, the same value as the ID token's account_status: banned (an active ban; deactivation by SCIM, directory sync or a workflow is a ban), suspended (a lock, suspension directive or sign-in lockout is in force), else active.
- `deleted_at` (Number) Epoch-ms; non-null once soft-deleted.
- `last_sign_in_at` (Number) Epoch-ms.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).
- `version` (Number) Strictly increases with every change to the user. Webhook payloads carry the same value: apply an update only if its version is higher than the one you hold.
- `referral` (String, JSON) How the user arrived through a referral program (referral attribution on). Null only when the user was never reported to the referral platform; otherwise `status` says where attribution stands (`pending` until the platform answers). Present on read, list, create and update responses. A JSON document (use `jsonencode`).
- `referral_code` (String) The user's own referral code: set once the referral platform enrolled the user as a partner (its answer carried a code) or an admin set one (PATCH /v1/users/{id} referral_code). null when the user is not enrolled, referrals are off, or the code was cleared.
- `warnings` (List of String) PATCH answers only, when there is something to know: e.g. the referral platform has not confirmed a new referral_code yet (always so in a dry run, which does not ask it).
- `referral_link` (String) The user's own share link: the link the referral platform returned for the partner, else the referral settings' link_template with {code} replaced by the URL-encoded referral_code. null while there is no referral_code, or no platform link and no link_template.
- `external_accounts` (String, JSON) Every linked provider identity, however it was linked (a sign-in, a connect flow, identities on create, a bulk import or identities/import); the same set GET /v1/users/{id}/identities lists. Present on read, list, create, update and password responses. A JSON document (use `jsonencode`).
