# atlas_chat_integration

Connect a Slack workspace or a Teams bot using your own app. Slack credentials are checked with the workspace before saving.

Scopes: `service_connectors:read` to read, `service_connectors:write` to write.

Backend API: `POST /v1/chat_integrations` (create); `GET /v1/chat_integrations/{id}` (read); `PATCH /v1/chat_integrations/{id}` (update); `DELETE /v1/chat_integrations/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_chat_integration" "example" {
  provider_value = "slack"
  name = "tf-chat-integration"
  credentials = jsonencode({})
}
```

## Schema

### Required

- `provider_value` (String) The `provider` field. One of: `slack`, `teams`. Changing it forces replacement.
- `name` (String) The `name` field. Changing it forces replacement.
- `credentials` (String, JSON, Sensitive) Slack: bot_token and signing_secret. Teams: app_id and app_password (your Azure Bot). Stored in the vault, write-only. A JSON document (use `jsonencode`). Write-only: sent to Atlas, never returned, and not detected as drift.

### Optional

- `config` (String, JSON) The `config` field. A JSON document (use `jsonencode`).
- `routing` (String, JSON) The `routing` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `enabled` (Boolean) The `enabled` field.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_chat_integration.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
