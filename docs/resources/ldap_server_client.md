# atlas_ldap_server_client

Create a service account for the LDAPS interface: base_dn (dc=/o=/ou= parts, unique across Atlas), bind_cn, allowed_cidrs (empty = nobody may connect), allow_user_bind and size_limit. The secret is returned once as secret.

Scopes: `ldap_server:read` to read, `ldap_server:write` to write.

Backend API: `POST /v1/ldap_server_clients` (create); `GET /v1/ldap_server_clients/{id}` (read); `PATCH /v1/ldap_server_clients/{id}` (update); `DELETE /v1/ldap_server_clients/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_ldap_server_client" "example" {
  name = "tf-ldap-server-client"
  base_dn = "tf-ldap-server-client"
  bind_cn = "tf-ldap-server-client"
}
```

## Schema

### Required

- `name` (String) The `name` field.
- `base_dn` (String) e.g. dc=acme,dc=com
- `bind_cn` (String) The service account cn under base_dn. Not returned by the API, so it is not detected as drift.

### Optional

- `allowed_cidrs` (List of String) Source networks allowed to bind (empty = anywhere).
- `allow_user_bind` (Boolean) The `allow_user_bind` field.
- `size_limit` (Number) The `size_limit` field.
- `status` (String) The `status` field. One of: `active`, `disabled`.

### Read-Only

- `id` (String) Atlas object id.
- `bind_dn` (String) The `bind_dn` field.
- `user_dn_template` (String) The `user_dn_template` field.
- `secret_prefix` (String, Sensitive) The `secret_prefix` field. Returned once, on create.
- `last_bind_at` (Number) The `last_bind_at` field. (epoch ms)
- `created_at` (Number) The `created_at` field. (epoch ms)
- `secret` (String, Sensitive) On create and rotate only: the bind password, shown once. Returned once, on create.

## Import

The import id is `<id>`.

```sh
terraform import atlas_ldap_server_client.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
