# atlas_organization_application

Give an organization access to an application (by client id or id), optionally limited to some organization roles. Idempotent; send `enabled: false` to pause access without removing it. Only matters while the application is `restricted`.

Scopes: `organizations:write` to read, `organizations:write` to write.

Backend API: `PUT /v1/organizations/{id}/applications/{clientId}` (create and update); no read route: refresh keeps the last applied values; `DELETE /v1/organizations/{id}/applications/{clientId}` (destroy).

## Example Usage

```hcl
resource "atlas_organization_application" "example" {
  organization_id = "tf-organization-application"
  client_id = "tf-organization-application"
}
```

## Schema

### Required

- `organization_id` (String) The organization id this object belongs to. Changing it forces replacement.
- `client_id` (String) The client id this object belongs to. Changing it forces replacement.

### Optional

- `roles` (List of String) The `roles` field.
- `enabled` (Boolean) The `enabled` field.

### Read-Only

- `id` (String) Atlas object id.
- `application_id` (String) The application's id.
- `application_name` (String) The `application_name` field.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<organization_id>/<client_id>`.

```sh
terraform import atlas_organization_application.example <organization_id>/<client_id>
```
