# atlas_mcp_registry_entry

Add a server to the registry as approved (default), or file it as pending for review.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `POST /v1/mcp_registry` (create); `GET /v1/mcp_registry` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/mcp_registry/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_mcp_registry_entry" "example" {
  name = "tf-mcp-registry-entry"
  url = "https://example.com/url"
}
```

## Schema

### Required

- `name` (String) The `name` field.
- `url` (String) The `url` field.

### Optional

- `description` (String) The `description` field.
- `status` (String) The `status` field. One of: `approved`, `pending`.

### Read-Only

- `id` (String) Atlas object id.
- `requested_by` (String) The `requested_by` field.
- `decided_by` (String) The `decided_by` field.
- `decided_at` (Number) The `decided_at` field. (epoch ms)
- `created_at` (Number) The `created_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_mcp_registry_entry.example <id>
```
