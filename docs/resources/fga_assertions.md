# atlas_fga_assertions

Replace the stored assertions for a model version.

Scopes: `fga:read` to read, `fga:write` to write.

Backend API: `PUT /v1/openfga/stores/{store}/assertions/{model}` (create and update); `GET /v1/openfga/stores/{store}/assertions/{model}` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_fga_assertions" "example" {
  store_id = "tf-fga-assertions"
  model_id = "tf-fga-assertions"
}
```

## Schema

### Required

- `store_id` (String) The store id this object belongs to. Changing it forces replacement.
- `model_id` (String) The model id this object belongs to. Changing it forces replacement.

### Optional

- `assertions` (String, JSON) The `assertions` field. A JSON document (use `jsonencode`).

### Read-Only

- `id` (String) Atlas object id.

## Import

The import id is `<store_id>/<model_id>`.

```sh
terraform import atlas_fga_assertions.example <store_id>/<model_id>
```
