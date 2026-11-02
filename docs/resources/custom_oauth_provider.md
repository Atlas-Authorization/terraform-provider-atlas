# atlas_custom_oauth_provider

Create a custom OAuth 2.0 / OIDC provider (Atlas is the relying party). OIDC: give the issuer and the endpoints are discovered (a bad issuer is a 400 now, not a broken button later). OAuth 2.0: give the authorize, token and userinfo endpoints. Set client authentication (client_secret_basic/post/jwt, private_key_jwt with an Atlas-held key, or none), a declarative profile_mapping, upstream logout and third-party initiated login. The response names the URLs to register at the provider. Starts disabled unless `enabled: true` (which needs a client id, a secret for the secret methods, and a signing key for private_key_jwt).

Scopes: `oauth_providers:read` to read, `oauth_providers:write` to write.

Backend API: `POST /v1/custom_oauth_providers` (create); `GET /v1/custom_oauth_providers/{id}` (read); `PATCH /v1/custom_oauth_providers/{id}` (update); `DELETE /v1/custom_oauth_providers/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_custom_oauth_provider" "example" {
  provider_key = "tf-custom-oauth-provider"
  display_name = "tf-custom-oauth-provider"
}
```

## Schema

### Required

- `provider_key` (String) 2-48 lowercase letters, digits or _; not a built-in provider. The sign-in `provider` value. Immutable. Changing it forces replacement.
- `display_name` (String) The `display_name` field.

### Optional

- `icon_url` (String) The `icon_url` field.
- `protocol` (String) oidc (default): id_token verified, discovery available. oauth2: hand-entered endpoints, profile from userinfo.
- `issuer` (String) OIDC: the issuer (the id_token `iss`); endpoints are discovered from it.
- `discovery_url` (String) A non-standard discovery document URL.
- `authorization_endpoint` (String) The `authorization_endpoint` field.
- `token_endpoint` (String) The `token_endpoint` field.
- `userinfo_endpoint` (String) The `userinfo_endpoint` field.
- `jwks_uri` (String) The `jwks_uri` field.
- `end_session_endpoint` (String) The `end_session_endpoint` field.
- `revocation_endpoint` (String) RFC 7009 revoke endpoint, used when a connection on this provider is disconnected. May carry {name} host templates.
- `host_params_schema` (String, JSON) For a data-only provider: the connect-time host_params the {name} templates in the endpoints take. Each value must match its pattern as a whole and use only letters, digits, dot, dash and underscore; the resulting endpoints must be https URLs on public hostnames. A JSON document (use `jsonencode`).
- `data_only` (Boolean) Only data connections may use this provider: it is never a sign-in method, a sign-in button or a login strategy. Required for endpoints with {name} templates. Default false.
- `scopes` (List of String) The `scopes` field.
- `scope_separator` (String) The `scope_separator` field.
- `pkce` (Boolean) Send PKCE (S256). Default true.
- `authorization_params` (String, JSON) Extra authorize parameters (protocol parameters are reserved). A JSON document (use `jsonencode`).
- `token_params` (String, JSON) Extra token-request parameters. A JSON document (use `jsonencode`).
- `userinfo_auth` (String) The `userinfo_auth` field. One of: `header`, `query`.
- `token_endpoint_auth_method` (String) How Atlas authenticates to the upstream token endpoint. private_key_jwt signs a client assertion with an Atlas-held key published at client_jwks_uri. One of: `client_secret_basic`, `client_secret_post`, `client_secret_jwt`, `private_key_jwt`, `none`.
- `client_assertion_audience` (String) The `aud` of a client assertion; default: the token endpoint.
- `profile_source` (String) Where the profile comes from. merged (default): userinfo over id_token claims.
- `profile_mapping` (String, JSON) Declarative profile mapping from the upstream userinfo / id_token claims to Atlas user fields and metadata. Unmapped fields keep the provider defaults. A remapped `email` is unverified unless `email_verified` is mapped too. null clears it. A JSON document (use `jsonencode`).
- `logout` (String, JSON) Upstream logout. OpenID Connect providers only. A JSON document (use `jsonencode`).
- `idp_initiated_login` (String, JSON) OIDC third-party initiated login at initiate_login_uri. OpenID Connect providers only. A JSON document (use `jsonencode`).
- `client_id` (String) The `client_id` field.
- `client_secret` (String, Sensitive) Write-only; never returned. Write-only: sent to Atlas, never returned, and not detected as drift.
- `enabled` (Boolean) The `enabled` field.
- `allow_sign_in` (Boolean) The `allow_sign_in` field.
- `allow_sign_up` (Boolean) The `allow_sign_up` field.

### Read-Only

- `id` (String) Atlas object id.
- `has_secret` (Boolean) The `has_secret` field.
- `redirect_uri` (String) Register at the provider as the redirect / callback URI.
- `client_jwks_uri` (String) Register as the client JWKS URL (private_key_jwt).
- `backchannel_logout_uri` (String) Register as the back-channel logout URI.
- `frontchannel_logout_uri` (String) Register as the front-channel logout URI.
- `post_logout_redirect_uri` (String) Register as a post-logout redirect URI.
- `initiate_login_uri` (String) Register as the initiate-login URI (third-party initiated login).
- `signing_keys` (String, JSON) The `signing_keys` field. A JSON document (use `jsonencode`).
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_custom_oauth_provider.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
