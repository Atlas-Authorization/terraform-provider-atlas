# atlas_connection_trigger

Subscribe to provider events. webhook: Atlas creates a verified ingress endpoint and, for providers whose API allows it (GitHub), registers the webhook there (filter: owner, repo); verified deliveries are forwarded to target_url. Other providers get an ingress_url to paste into the provider and need `scheme`. poll: a read-only catalog action runs on an interval and each new item becomes a `connection.trigger_fired` event.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `POST /v1/connections/{id}/triggers` (create); `GET /v1/connections/{id}/triggers/{triggerId}` (read); `PATCH /v1/connections/{id}/triggers/{triggerId}` (update); `DELETE /v1/connections/{id}/triggers/{triggerId}` (destroy).

## Example Usage

```hcl
resource "atlas_connection_trigger" "example" {
  connection_id = "tf-connection-trigger"
  event = "tf-connection-trigger"
}
```

## Schema

### Required

- `connection_id` (String) The connection id this object belongs to. Changing it forces replacement.
- `event` (String) The provider's event name (webhook) or a label of your own (poll). Changing it forces replacement.

### Optional

- `mode` (String) The `mode` field. One of: `webhook`, `poll`. Changing it forces replacement.
- `filter` (String, JSON) GitHub: owner and repo. A JSON document (use `jsonencode`). Changing it forces replacement.
- `target_url` (String) webhook: https URL that receives verified deliveries. Changing it forces replacement.
- `scheme` (String) webhook, providers Atlas cannot register for: the provider's signature scheme. Not returned by the API, so it is not detected as drift. Changing it forces replacement.
- `signing_secret` (String, Sensitive) The `signing_secret` field. Write-only: sent to Atlas, never returned, and not detected as drift. Changing it forces replacement.
- `options` (String, JSON) The `options` field. A JSON document (use `jsonencode`). Not returned by the API, so it is not detected as drift. Changing it forces replacement.
- `poll` (String, JSON) poll: { action, input, items_path, id_field, interval_seconds (60 to 86400) }. A JSON document (use `jsonencode`). Changing it forces replacement.

### Read-Only

- `id` (String) Atlas object id.
- `status` (String) The `status` field. One of: `active`, `paused`, `failed`.
- `provider_registered` (Boolean) True when Atlas registered the webhook at the provider.
- `ingress_url` (String) Where the provider posts, for webhook triggers.
- `next_poll_at` (Number) The `next_poll_at` field. (epoch ms)
- `last_error` (String) The `last_error` field.
- `last_fired_at` (Number) The `last_fired_at` field. (epoch ms)
- `fired_count` (Number) The `fired_count` field.
- `forward_secret` (String, Sensitive) Shown once on creation of a webhook trigger: verifies the deliveries Atlas sends to target_url. Returned once, on create.
- `created_at` (Number) The `created_at` field. (epoch ms)

## Import

The import id is `<connection_id>/<id>`.

```sh
terraform import atlas_connection_trigger.example <connection_id>/<id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
