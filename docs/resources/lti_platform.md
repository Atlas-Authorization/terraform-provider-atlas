# atlas_lti_platform

Register an LTI 1.3 platform (issuer, client_id, auth_login_url, jwks_uri, deployment_ids). https URLs required; a duplicate (issuer, client_id) is a 409.

Scopes: `lti:read` to read, `lti:write` to write.

Backend API: `POST /v1/lti_platforms` (create); `GET /v1/lti_platforms/{id}` (read); `PATCH /v1/lti_platforms/{id}` (update); `DELETE /v1/lti_platforms/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_lti_platform" "example" {

}
```

## Schema

### Optional

- `issuer` (String) The `issuer` field. Changing it forces replacement.
- `client_id` (String) The `client_id` field. Changing it forces replacement.
- `auth_login_url` (String) The `auth_login_url` field.
- `jwks_uri` (String) The `jwks_uri` field.
- `deployment_ids` (List of String) The `deployment_ids` field.
- `organization_id` (String) The `organization_id` field.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_lti_platform.example <id>
```
