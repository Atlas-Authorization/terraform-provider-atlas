# atlas_oauth_publisher_client

Attach an app (its public client_id) to a verified publisher. The app's homepage_url must be an https URL on the publisher's domain, and an app belongs to one publisher. With a shared sector the app must be pairwise and takes the publisher's sector; if it already has another sector, send allow_subject_change: true to confirm that users' subject identifiers in it change. Attaching an app already attached returns the publisher unchanged.

Scopes: `oauth_clients:write` to read, `oauth_clients:write` to write.

Backend API: `PUT /v1/oauth_publishers/{id}/clients/{client_id}` (create and update); no read route: refresh keeps the last applied values; `DELETE /v1/oauth_publishers/{id}/clients/{client_id}` (destroy).

## Example Usage

```hcl
resource "atlas_oauth_publisher_client" "example" {
  publisher_id = "tf-oauth-publisher-client"
  client_id = "tf-oauth-publisher-client"
}
```

## Schema

### Required

- `publisher_id` (String) The publisher id this object belongs to. Changing it forces replacement.
- `client_id` (String) The client id this object belongs to. Changing it forces replacement.

### Optional

- `allow_subject_change` (Boolean) Confirm that attaching a pairwise app that has another sector changes every user's subject identifier in it. Default false.

### Read-Only

- `id` (String) Atlas object id.
- `name` (String) Shown on the consent screen of every app of the publisher: "Published by <name>".
- `domain` (String) The domain the publisher controls, for example example.com. Fixed at create.
- `shared_sector` (Boolean) Give every pairwise app of the publisher one sector (the domain), so a user has the same subject identifier in all of them. Needs a verified domain. Turning it on does not touch apps already attached; attaching a pairwise app sets its sector.
- `skip_consent` (Boolean) A user who already consented to another app of the publisher for the same scopes is not asked again. Only openid, email and profile are covered; every other scope is always asked.
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

The import id is `<publisher_id>/<client_id>`.

```sh
terraform import atlas_oauth_publisher_client.example <publisher_id>/<client_id>
```
