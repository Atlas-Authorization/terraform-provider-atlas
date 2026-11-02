# atlas_fga_store

Create a fine-grained authorization store — an isolated namespace for authorization models and relationship tuples.

Scopes: `fga:read` to read, `fga:write` to write.

Backend API: `POST /v1/fga/stores` (create); `GET /v1/fga/stores/{id}` (read); no update route: every change replaces the object; `DELETE /v1/fga/stores/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_fga_store" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_fga_store.example <id>
```
