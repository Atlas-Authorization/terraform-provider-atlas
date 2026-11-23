# atlas_messaging_provider_fallback

Add a fallback provider (up to 5 per channel). When your provider errors, Atlas tries the fallbacks in order; the ledger shows each attempt. A provider refusing one recipient does not fail over. The body is validated per transport exactly like PUT /v1/messaging_providers/{channel}. Credentials are yours and write-only.

Scopes: `messaging:read` to read, `messaging:write` to write.

Backend API: `POST /v1/messaging_providers/{channel}/fallbacks` (create); `GET /v1/messaging_providers/{channel}/fallbacks` (read: the list is scanned for the id); `PUT /v1/messaging_providers/{channel}/fallbacks/{id}` (update); `DELETE /v1/messaging_providers/{channel}/fallbacks/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_messaging_provider_fallback" "example" {
  channel = "tf-messaging-provider-fallback"
  transport = "tf-messaging-provider-fallback"
}
```

## Schema

### Required

- `channel` (String) The channel this object belongs to. Changing it forces replacement.
- `transport` (String) Any transport the channel accepts, with its own `config` keys, exactly as for PUT /v1/messaging_providers/{channel}.

### Optional

- `from_address` (String) The `from_address` field.
- `config` (String, JSON) The transport's settings. Secrets are write-only; on an update an omitted secret keeps the stored one. A JSON document (use `jsonencode`).
- `position` (Number) Where it sits in the failover order (lower is tried first). New providers go last.
- `status` (String) The `status` field. One of: `active`, `disabled`.

### Read-Only

- `id` (String) Atlas object id.
- `secrets_set` (String, JSON, Sensitive) Which secret keys are set, never their values. A JSON document (use `jsonencode`). Returned once, on create.
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<channel>/<id>`.

```sh
terraform import atlas_messaging_provider_fallback.example <channel>/<id>
```
