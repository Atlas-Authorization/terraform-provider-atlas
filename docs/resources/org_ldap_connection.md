# atlas_org_ldap_connection

Create an organization-owned LDAP / AD connection (url, bind_dn, base_dn, user_filter, attributes). The service-account bind_password is write-only; one connection per organization.

Scopes: `sso_connections:read` to read, `sso_connections:write` to write.

Backend API: `POST /v1/org_ldap_connections` (create); `GET /v1/org_ldap_connections/{id}` (read); `PATCH /v1/org_ldap_connections/{id}` (update); `DELETE /v1/org_ldap_connections/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_org_ldap_connection" "example" {

}
```

## Schema

### Optional

- `organization_id` (String) The owning organization (required on create).
- `status` (String) The `status` field. One of: `draft`, `active`, `disabled`.
- `url` (String) ldap:// or ldaps:// directory URL.
- `bind_dn` (String) Service-account DN Atlas binds as to search.
- `base_dn` (String) Subtree searched under.
- `user_filter` (String) Search filter with a {{username}} placeholder.
- `start_tls` (Boolean) The `start_tls` field.
- `require_consent_screen` (Boolean) Show the person a confirmation ("sign in as ...") before an LDAP sign-in completes. The sign-in response carries consent_required and a consent_token instead of a ticket.
- `attributes` (String, JSON) Directory attribute → Atlas field mapping. `email` is required. A JSON document (use `jsonencode`).
- `bind_password` (String, Sensitive) Service-account password. Write-only; never returned. Write-only: sent to Atlas, never returned, and not detected as drift.
- `default_role_id` (String) The `default_role_id` field.

### Read-Only

- `id` (String) Atlas object id.
- `has_bind_password` (Boolean) Whether a bind password is set — never the password.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_org_ldap_connection.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
