# atlas_directory_pull_source

Create a pull source for an organization: provider (entra, google_workspace, bamboohr, hibob, personio, workday, successfactors, gusto, csv, rest_json), config, write-only credentials and settings (schedule, role mappings, disable on removal).

Scopes: `directory_pull:read` to read, `directory_pull:write` to write.

Backend API: `POST /v1/directory_pull_sources` (create); `GET /v1/directory_pull_sources/{id}` (read); `PATCH /v1/directory_pull_sources/{id}` (update); `DELETE /v1/directory_pull_sources/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_directory_pull_source" "example" {
  organization_id = "tf-directory-pull-source"
  name = "tf-directory-pull-source"
  provider_value = "entra"
}
```

## Schema

### Required

- `organization_id` (String) The `organization_id` field.
- `name` (String) The `name` field.
- `provider_value` (String) The `provider` field. One of: `entra`, `google_workspace`, `bamboohr`, `hibob`, `personio`, `workday`, `successfactors`, `gusto`, `csv`, `rest_json`.

### Optional

- `status` (String) The `status` field. One of: `active`, `paused`.
- `config` (String, JSON) The `config` field. A JSON document (use `jsonencode`).
- `credentials` (String, JSON, Sensitive) Write-only. A JSON document (use `jsonencode`). Write-only: sent to Atlas, never returned, and not detected as drift.
- `default_role_id` (String) The `default_role_id` field.
- `settings` (String, JSON) Schedule, role mappings and disable-on-removal. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.
- `has_credentials` (Boolean) The `has_credentials` field.
- `has_uploaded_csv` (Boolean) The `has_uploaded_csv` field.
- `next_sync_at` (Number) The `next_sync_at` field. (epoch ms)
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_directory_pull_source.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
