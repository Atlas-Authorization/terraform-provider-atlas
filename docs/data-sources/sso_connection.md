# atlas_sso_connection (data source)

An enterprise SSO connection, looked up by id.

Every attribute except `id` is computed. Scope: `sso_connections:read`.

## Example Usage

```hcl
data "atlas_sso_connection" "example" {
  id = "<id>"
}
```

## Schema

### Required

- `id` (String) The id to look up.

### Read-Only

- `organization_id` (String) The `organization_id` field.
- `type` (String) The `type` field.
- `status` (String) The `status` field.
- `oidc_issuer` (String) The `oidc_issuer` field.
- `oidc_client_id` (String) The `oidc_client_id` field.
- `has_secret` (Boolean, Sensitive) The `has_secret` field.
- `saml_idp_entity_id` (String) The `saml_idp_entity_id` field.
- `saml_idp_sso_url` (String) The `saml_idp_sso_url` field.
- `saml_sp_entity_id` (String) The `saml_sp_entity_id` field.
- `has_saml_certificate` (Boolean) The `has_saml_certificate` field.
- `saml_allow_idp_initiated` (Boolean) The `saml_allow_idp_initiated` field.
- `saml_sign_authn_requests` (Boolean) The `saml_sign_authn_requests` field.
- `saml_want_response_signed` (Boolean) The `saml_want_response_signed` field.
- `saml_single_logout` (Boolean) The `saml_single_logout` field.
- `saml_idp_slo_url` (String) The `saml_idp_slo_url` field.
- `saml_idp_slo_response_url` (String) The `saml_idp_slo_response_url` field.
- `saml_idp_slo_binding` (String) The `saml_idp_slo_binding` field. One of: `redirect`, `post`.
- `saml_idp_slo_soap_url` (String) The `saml_idp_slo_soap_url` field.
- `saml_require_encrypted_assertions` (Boolean) The `saml_require_encrypted_assertions` field.
- `require_consent_screen` (Boolean) The `require_consent_screen` field.
- `has_saml_encryption_key` (Boolean) Whether an SP encryption key pair is set up (the private key is never returned).
- `has_previous_saml_encryption_key` (Boolean) A rotation is in progress: the previous key is still advertised and accepted.
- `saml_encryption_rotated_at` (Number) When the current encryption key was installed (epoch ms).
- `manual_endpoints` (String, JSON) The `manual_endpoints` field. A JSON document (use `jsonencode`).
- `has_discourse_secret` (Boolean, Sensitive) The `has_discourse_secret` field.
- `discourse_provider_url` (String) The `discourse_provider_url` field.
- `allowed_domains` (List of String) The `allowed_domains` field.
- `claim_role_mappings` (String, JSON) The `claim_role_mappings` field. A JSON document (use `jsonencode`).
- `default_role_id` (String) The `default_role_id` field.
- `provisioning_mode` (String) The `provisioning_mode` field. One of: `member`, `end_user`.
- `attribute_mapping` (String, JSON) The configured field mapping (only the set keys). A JSON document (use `jsonencode`).
- `oidc_scopes` (List of String) The `oidc_scopes` field.
- `profile_mapping` (String, JSON) Declarative profile mapping from the upstream userinfo / id_token claims to Atlas user fields and metadata. Unmapped fields keep the provider defaults. A remapped `email` is unverified unless `email_verified` is mapped too. null clears it. A JSON document (use `jsonencode`).
- `token_endpoint_auth_method` (String) How Atlas authenticates to the upstream token endpoint. private_key_jwt signs a client assertion with an Atlas-held key published at client_jwks_uri. One of: `client_secret_basic`, `client_secret_post`, `client_secret_jwt`, `private_key_jwt`, `none`.
- `client_assertion_audience` (String) The `client_assertion_audience` field.
- `logout` (String, JSON) Upstream logout. OpenID Connect providers only. A JSON document (use `jsonencode`).
- `idp_initiated_login` (String, JSON) OIDC third-party initiated login at initiate_login_uri. OpenID Connect providers only. A JSON document (use `jsonencode`).
- `upstream_urls` (String, JSON) OIDC connections: the URLs to register at the IdP. A JSON document (use `jsonencode`).
- `domain_warnings` (String, JSON) Routing domains of an organization-owned connection that are not verified domains of the organization (saved before verification was required). Verify them or remove them. A JSON document (use `jsonencode`).
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).
