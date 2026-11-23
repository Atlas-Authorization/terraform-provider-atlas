# atlas_pca_action_namespace_pin

Pin a namespace to one stored version: that version is active and later publishes do not replace it. Destroying the resource unpins it (later publishes activate again). The API has no read route for a pin, so refresh keeps the last applied values.

Scopes: `pca:read` to read, `pca:write` to write.

Backend API: `PUT /v1/pca/action-registry/namespaces/{ns}/pin` (create and update); no read route: refresh keeps the last applied values; `DELETE /v1/pca/action-registry/namespaces/{ns}/pin` (destroy).

## Example Usage

```hcl
resource "atlas_pca_action_namespace_pin" "example" {
  ns      = "billing"
  version = "1.4.0"
}
```

## Schema

### Required

- `ns` (String) The action namespace. Changing it forces replacement.
- `version` (String) The stored version to activate and pin.

### Read-Only

- `id` (String) Atlas object id.
- `active_version` (String) The version that is active after the pin.

## Import

The import id is `<ns>`.

```sh
terraform import atlas_pca_action_namespace_pin.example <ns>
```
