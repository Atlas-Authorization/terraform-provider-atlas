# atlas_oauth_client_profile

Create or replace a custom client policy profile (rules: PKCE, PAR, sender constraint, signed request object, allowed auth methods and signing algorithms, https-only redirects, allowed claims). Built-in names are reserved.

Scopes: `oauth_clients:write` to read, `oauth_clients:write` to write.

Backend API: `PUT /v1/oauth_client_profiles/{name}` (create and update); no read route: refresh keeps the last applied values; `DELETE /v1/oauth_client_profiles/{name}` (destroy).

## Example Usage

```hcl
resource "atlas_oauth_client_profile" "example" {
  name = "tf-oauth-client-profile"
}
```

## Schema

### Required

- `name` (String) The name this object belongs to. Changing it forces replacement.

### Optional

- `description` (String) The `description` field.
- `require_pkce` (Boolean) The `require_pkce` field.
- `require_par` (Boolean) Authorization requests must be pushed (RFC 9126).
- `sender_constraint` (String) The token request must prove DPoP, mutual TLS, or either (any). One of: `none`, `dpop`, `mtls`, `any`.
- `require_signed_request` (Boolean) Authorization requests must be a signed request object (RFC 9101), pushed or direct.
- `token_endpoint_auth_methods` (List of String) Allowed client authentication methods; null = any.
- `signing_algs` (List of String) JWS algorithms accepted for request objects and client assertions; null = any asymmetric.
- `https_redirects_only` (Boolean) The `https_redirects_only` field.
- `allowed_claims` (List of String) Claim names a `claims` request may ask for; null = any.

### Read-Only

- `id` (String) Atlas object id.

## Import

The import id is `<name>`.

```sh
terraform import atlas_oauth_client_profile.example <name>
```
