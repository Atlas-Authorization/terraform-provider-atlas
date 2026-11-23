# atlas_pca_principal

Enrol a principal public key, binding it to an Atlas identity (user, org or machine). A grant whose issuer is an enrolled active principal registers as the verified tier. Destroying the resource disables the principal: its existing verified grants are denied and no new verified grant can be registered under it.

Scopes: `pca:read` to read, `pca:write` to write.

Backend API: `POST /v1/pca/principals` (create); `GET /v1/pca/principals` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/pca/principals/{principalPub}` (destroy).

## Example Usage

```hcl
resource "atlas_pca_principal" "example" {
  principal_pub = "<base64url Ed25519 public key>"
  subject_type  = "user"
  subject_id    = "user_123"
}
```

## Schema

### Required

- `principal_pub` (String) The principal base64url Ed25519 public key to enrol.
- `subject_type` (String) What the key is bound to. One of: `user`, `org`, `machine`.
- `subject_id` (String) The id of the user, org or machine, which must exist in this instance.

### Read-Only

- `id` (String) Atlas object id.
- `status` (String) The principal status: active or disabled.

## Import

The import id is `<id>`.

```sh
terraform import atlas_pca_principal.example <id>
```
