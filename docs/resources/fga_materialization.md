# atlas_fga_materialization

Precompute list-objects for every user of `user_type` on `object_type#relation`, so a huge list question is one indexed read. The first build runs now; the worker rebuilds after tuple writes. A stored answer is served only while it is current for the store revision and model, before any tuple expiry, and for a plain `type:id` user with no request context or contextual tuples; otherwise list-objects is answered live. A model that uses conditions cannot be materialized.

Scopes: `fga:read` to read, `fga:write` to write.

Backend API: `POST /v1/fga/stores/{id}/materializations` (create); `GET /v1/fga/stores/{id}/materializations` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/fga/stores/{id}/materializations/{mid}` (destroy).

## Example Usage

```hcl
resource "atlas_fga_materialization" "example" {
  store_id = "tf-fga-materialization"
}
```

## Schema

### Required

- `store_id` (String) The store id this object belongs to. Changing it forces replacement.

### Optional

- `object_type` (String) The `object_type` field.
- `relation` (String) The `relation` field.
- `user_type` (String) The `user_type` field.

### Read-Only

- `id` (String) Atlas object id.
- `status` (String) The `status` field.
- `built_revision` (Number) The `built_revision` field.
- `built_at` (Number) The `built_at` field.
- `valid_until` (Number) The `valid_until` field.
- `users` (Number) The `users` field.
- `rows` (Number) The `rows` field.
- `error` (String) The `error` field.

## Import

The import id is `<store_id>/<id>`.

```sh
terraform import atlas_fga_materialization.example <store_id>/<id>
```
