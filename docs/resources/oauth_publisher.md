# atlas_oauth_publisher

Claim a publisher name for a domain. Returns the DNS TXT record (record_name and record_value) to publish; then call verify. Nothing changes for any app until the domain is verified and the app is attached. Audited as oauth_publisher.created; emits oauth_publisher.updated.

Scopes: `oauth_clients:read` to read, `oauth_clients:write` to write.

Backend API: `POST /v1/oauth_publishers` (create); `GET /v1/oauth_publishers/{id}` (read); `PATCH /v1/oauth_publishers/{id}` (update); `DELETE /v1/oauth_publishers/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_oauth_publisher" "example" {
  name = "tf-oauth-publisher"
  domain = "tf-oauth-publisher"
}
```

## Schema

### Required

- `name` (String) Shown on the consent screen of every app of the publisher: "Published by <name>".
- `domain` (String) The domain the publisher controls, for example example.com. Fixed at create. Changing it forces replacement.

### Optional

- `shared_sector` (Boolean) Give every pairwise app of the publisher one sector (the domain), so a user has the same subject identifier in all of them. Needs a verified domain. Turning it on does not touch apps already attached; attaching a pairwise app sets its sector.
- `skip_consent` (Boolean) A user who already consented to another app of the publisher for the same scopes is not asked again. Only openid, email and profile are covered; every other scope is always asked.

### Read-Only

- `id` (String) Atlas object id.
- `verified` (Boolean) True once the DNS TXT record was found. Only a verified publisher's name is shown and only a verified publisher shares a sector or skips consent.
- `verified_at` (Number) Epoch ms.
- `last_checked_at` (Number) Epoch ms of the last verify call.
- `record_name` (String) The DNS name to publish the TXT record at.
- `record_value` (String) The TXT value to publish.
- `sector_identifier` (String) The shared sector (the domain) when shared_sector is on and the domain is verified; otherwise null.
- `client_ids` (List of String) The public client ids of the attached apps.
- `created_at` (Number) Epoch ms.
- `updated_at` (Number) Epoch ms.

## Import

The import id is `<id>`.

```sh
terraform import atlas_oauth_publisher.example <id>
```
