# atlas_pca_action_source

Track a document (an OpenAPI spec, an MCP tool list and so on) as a source of agent actions: each sync turns what it lists into drafts for review, on a schedule or on demand. Destroying the resource stops tracking the source; drafts already made stay.

Scopes: `pca:read` to read, `pca:write` to write.

Backend API: `POST /v1/pca/action-registry/sources` (create); `GET /v1/pca/action-registry/sources` (read: the list is scanned for the id); `PATCH /v1/pca/action-registry/sources/{id}` (update); `DELETE /v1/pca/action-registry/sources/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_pca_action_source" "example" {
  ns       = "billing"
  from     = "openapi"
  url      = "https://example.com/openapi.json"
  schedule = "daily"
}
```

## Schema

### Required

- `ns` (String) The action namespace the source feeds. Changing it forces replacement.
- `from` (String) The kind of document at the URL: the import kind (for example `openapi`). Not returned by the API (it returns `kind`), so it is not detected as drift. Changing it forces replacement.
- `url` (String) Where the document is fetched from (https). Not returned by the API (it returns `uri`), so it is not detected as drift.

### Optional

- `verb` (String) For a source that lists one operation, the verb it becomes.
- `schedule` (String) How often the source is re-read. One of: `off`, `hourly`, `daily`.

### Read-Only

- `id` (String) Atlas object id.
- `kind` (String) The import kind the API recorded.
- `uri` (String) The document location the API recorded.
- `last_synced_at` (String) When the source was last read.
- `next_run_at` (String) When the next scheduled read is due.
- `last_status` (String) The outcome of the last read.
- `last_error` (String) The error of the last read, if it failed.
- `failures` (Number) Consecutive failed reads.
- `operations` (Number) How many operations the source lists.
- `created_at` (String) Creation time.

## Import

The import id is `<id>`.

```sh
terraform import atlas_pca_action_source.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
