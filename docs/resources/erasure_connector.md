# atlas_erasure_connector

Track a downstream system. Send catalog_id to take a catalog entry, or key and name for your own. mode ack (default): the system acknowledges through POST /v1/users/{id}/erasure_acks under ack_source. manual: an operator marks it done on each request. api: Atlas calls the vendor's deletion API (stripe, hubspot, mailchimp, segment, intercom, zendesk, brevo, customerio) with your own key (credential, sealed, never returned) for every address of the person when an erasure request is carried out; the result is recorded per request. 409 when the key exists. Honours Idempotency-Key.

Scopes: `data_subject_requests:read` to read, `data_subject_requests:write` to write.

Backend API: `POST /v1/erasure_connectors` (create); `GET /v1/erasure_connectors/{id}` (read); `PATCH /v1/erasure_connectors/{id}` (update); `DELETE /v1/erasure_connectors/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_erasure_connector" "example" {

}
```

## Schema

### Optional

- `key` (String) A slug (lowercase letters, digits, - and _). Defaults to the catalog id.
- `catalog_id` (String) An entry of GET /v1/erasure_connector_catalog; fills key and name.
- `name` (String) The `name` field.
- `mode` (String) The `mode` field. One of: `ack`, `manual`, `api`.
- `vendor` (String) mode api: the vendor to call (defaults to the catalog id when that is one of these). One of: `stripe`, `hubspot`, `mailchimp`, `segment`, `intercom`, `zendesk`, `brevo`, `customerio`, `null`.
- `config` (String, JSON) mode api: the vendor's non-secret settings; see api_vendor.settings in the catalog. A JSON document (use `jsonencode`).
- `credential` (String, Sensitive) mode api: your own key for the vendor. Sealed at rest and never returned; send it again to replace it. Write-only: sent to Atlas, never returned, and not detected as drift.
- `ack_source` (String) The `source` the system uses when it acknowledges (defaults to key).
- `enabled` (Boolean) The `enabled` field.
- `notes` (String) The `notes` field.

### Read-Only

- `id` (String) Atlas object id.
- `has_credential` (Boolean) Whether a vendor key is saved. The key itself is never returned.
- `created_at` (Number) Epoch ms.
- `updated_at` (Number) Epoch ms.

## Import

The import id is `<id>`.

```sh
terraform import atlas_erasure_connector.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
