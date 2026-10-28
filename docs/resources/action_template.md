# atlas_action_template

Create a gallery template (slug, name, author_name, summary, trigger, licence…), optionally with its first version (version, code, secret names). Versions stay drafts until submitted for review.

Scopes: `actions:read` to read, `actions:write` to write.

Backend API: `POST /v1/action_templates` (create); `GET /v1/action_templates/{id}` (read); `PATCH /v1/action_templates/{id}` (update); `DELETE /v1/action_templates/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_action_template" "example" {

}
```

## Schema

### Optional

- `slug` (String) The `slug` field.
- `name` (String) The `name` field.
- `author_name` (String) The `author_name` field.
- `summary` (String) The `summary` field.
- `description` (String) The `description` field.
- `category` (String) The `category` field.
- `tags` (List of String) The `tags` field.
- `trigger` (String) The `trigger` field.
- `license` (String) The `license` field.
- `repository_url` (String) The `repository_url` field.
- `version` (String) The `version` field. Not returned by the API, so it is not detected as drift. Changing it forces replacement.
- `code` (String) The `code` field. Not returned by the API, so it is not detected as drift. Changing it forces replacement.
- `secrets` (String, JSON, Sensitive) The `secrets` field. A JSON document (use `jsonencode`). Write-only: sent to Atlas, never returned, and not detected as drift. Changing it forces replacement.
- `changelog` (String) The `changelog` field. Not returned by the API, so it is not detected as drift. Changing it forces replacement.
- `submit` (Boolean) The `submit` field. Not returned by the API, so it is not detected as drift. Changing it forces replacement.

### Read-Only

- `id` (String) Atlas object id.
- `status` (String) The `status` field. One of: `draft`, `in_review`, `published`, `unlisted`, `rejected`, `banned`.
- `latest_version` (String) The `latest_version` field.
- `verified` (Boolean) The `verified` field.
- `featured` (Boolean) The `featured` field.
- `install_count` (Number) The `install_count` field.
- `view_count` (Number) The `view_count` field.
- `published_at` (Number) The `published_at` field. (epoch ms)
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).
- `versions` (String, JSON) The `versions` field. A JSON document (use `jsonencode`).

## Import

The import id is `<id>`.

```sh
terraform import atlas_action_template.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
