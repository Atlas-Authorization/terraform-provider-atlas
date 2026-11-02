# atlas_action

Create an Action for a trigger (see the Actions guide for the 22 triggers: sign-in, registration, tokens, credentials, MFA, profile, organizations, billing). Secrets are write-only. `runtime: "http"` (with `http_target` and an optional `signing_secret`) makes the Action call your own HTTPS endpoint, Standard Webhooks signed, instead of running code; a generated signing secret is returned once as `signing_secret`.

Scopes: `actions:read` to read, `actions:write` to write.

Backend API: `POST /v1/actions` (create); `GET /v1/actions/{id}` (read); `PATCH /v1/actions/{id}` (update); `DELETE /v1/actions/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_action" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `trigger` (String) The `trigger` field. Changing it forces replacement.
- `code` (String) The `code` field.
- `enabled` (Boolean) The `enabled` field.
- `secrets` (String, JSON, Sensitive) The `secrets` field. A JSON document (use `jsonencode`). Write-only: sent to Atlas, never returned, and not detected as drift.
- `runtime` (String) The `runtime` field. One of: `node`, `http`. Changing it forces replacement.
- `http_target` (String, JSON) The `http_target` field. A JSON document (use `jsonencode`).
- `signing_secret` (String) The `signing_secret` field.

### Read-Only

- `id` (String) Atlas object id.
- `secret_names` (List of String, Sensitive) The `secret_names` field. Returned once, on create.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_action.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
