# atlas_sso_onboarding_profile

Create a self-service SSO onboarding profile: allowed connection types and an optional bound organization.

Scopes: `sso_connections:read` to read, `sso_connections:write` to write.

Backend API: `POST /v1/sso_onboarding_profiles` (create); `GET /v1/sso_onboarding_profiles/{id}` (read); no update route: every change replaces the object; `DELETE /v1/sso_onboarding_profiles/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_sso_onboarding_profile" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `allowed_connection_types` (List of String) The `allowed_connection_types` field. One of: `oidc`, `saml`.
- `organization_id` (String) The `organization_id` field.
- `company_name` (String) The `company_name` field.
- `allow_scim` (Boolean) The `allow_scim` field.
- `intents` (List of String) The `intents` field. One of: `sso`, `directory_sync`, `domain_verification`, `log_stream`, `audit_log`, `certificate_renewal`, `api_keys`, `webhooks`.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_sso_onboarding_profile.example <id>
```
