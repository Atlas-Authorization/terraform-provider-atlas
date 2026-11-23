# atlas_pca_action_publisher_key

Pin a publisher key: action packs signed by it are accepted into the registry. Destroying the resource revokes the key. The API has no read route for a single key, so refresh keeps the last applied values.

Scopes: `pca:read` to read, `pca:write` to write.

Backend API: `POST /v1/pca/action-registry/publisher-keys` (create); no read route: refresh keeps the last applied values; no update route: every change replaces the object; `DELETE /v1/pca/action-registry/publisher-keys/{kid}` (destroy).

## Example Usage

```hcl
resource "atlas_pca_action_publisher_key" "example" {
  kid          = "acme-publisher-1"
  pk           = "<base64url Ed25519 public key>"
  display_name = "Acme publisher"
}
```

## Schema

### Required

- `kid` (String) The key id: 1 to 128 of letters, digits and `._:-`. Changing it forces replacement.
- `pk` (String) The base64url Ed25519 public key (32 bytes).

### Optional

- `display_name` (String) A label for the key (at most 120 characters).

### Read-Only

- `id` (String) Atlas object id.

## Import

The import id is `<id>`.

```sh
terraform import atlas_pca_action_publisher_key.example <id>
```
