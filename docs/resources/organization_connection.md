# atlas_organization_connection

Invite a partner organization to use an application the supplier organization holds. Returns a one-time `token` for the partner's administrator; the invitation lasts 14 days.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `POST /v1/organization_connections` (create); `GET /v1/organization_connections` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/organization_connections/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_organization_connection" "example" {
  supplier_organization_id = "tf-organization-connection"
  partner_organization_id = "tf-organization-connection"
  application = "tf-organization-connection"
}
```

## Schema

### Required

- `supplier_organization_id` (String) The `supplier_organization_id` field.
- `partner_organization_id` (String) The `partner_organization_id` field.
- `application` (String) The `application` field. Not returned by the API, so it is not detected as drift.

### Optional

- `roles` (List of String) The `roles` field.

### Read-Only

- `id` (String) Atlas object id.
- `application_id` (String) The `application_id` field.
- `status` (String) The `status` field. One of: `pending`, `active`, `revoked`, `expired`.
- `expires_at` (Number) The `expires_at` field. (epoch ms)
- `accepted_by` (String) The `accepted_by` field.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `token` (String, Sensitive) The `token` field. Returned once, on create.
- `note` (String) The `note` field.

## Import

The import id is `<id>`.

```sh
terraform import atlas_organization_connection.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
