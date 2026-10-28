# atlas_oauth_client_trusted_issuer

Trust an external issuer for this client. With purpose user_assertion (default) its signed JWTs (verified against the inline jwks or the https jwks_uri) can then be presented to POST /oauth2/token with grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer. subject_mapping says how the JWT sub names an Atlas user. With purpose client_auth or both (PRO-06 workload identity) a client registered with token_endpoint_auth_method workload_jwt authenticates by presenting a token from this issuer as its client_assertion; subject_pattern is then required and must name the workload. workload_kind github_actions, gcp and azure fill in the issuer and JWKS URL. Audited as oauth_trusted_issuer.created.

Scopes: `oauth_clients:read` to read, `oauth_clients:write` to write.

Backend API: `POST /v1/oauth_clients/{id}/trusted_issuers` (create); `GET /v1/oauth_clients/{id}/trusted_issuers` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/oauth_clients/{id}/trusted_issuers/{issuerId}` (destroy).

## Example Usage

```hcl
resource "atlas_oauth_client_trusted_issuer" "example" {
  client_id = "tf-oauth-client-trusted-issuer"
  issuer = "https://example.com/issuer"
}
```

## Schema

### Required

- `client_id` (String) The client id this object belongs to. Changing it forces replacement.
- `issuer` (String) The `issuer` field.

### Optional

- `jwks` (String, JSON) Inline JWK Set. Provide exactly one of jwks or jwks_uri. A JSON document (use `jsonencode`).
- `jwks_uri` (String) https URL of the issuer JWK Set. Provide exactly one of jwks or jwks_uri.
- `subject_mapping` (String) The `subject_mapping` field. One of: `user_id`, `external_id`, `email`.
- `workload_kind` (String) Fills in issuer and jwks_uri for github_actions, gcp and azure when you omit them. One of: `generic`, `spiffe`, `kubernetes`, `github_actions`, `gcp`, `azure`.
- `tenant_id` (String) Required for workload_kind azure (without an explicit issuer). Not returned by the API, so it is not detected as drift.
- `purpose` (String) The `purpose` field. One of: `user_assertion`, `client_auth`, `both`.
- `subject_pattern` (String) Required when purpose is client_auth or both. A glob with * wildcards; a bare * is refused.
- `required_claims` (String, JSON) Extra claims that must match exactly. A JSON document (use `jsonencode`).
- `audiences` (List of String) Accepted aud values. Defaults to the token endpoint, the PAR endpoint and the issuer.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).

## Import

The import id is `<client_id>/<id>`.

```sh
terraform import atlas_oauth_client_trusted_issuer.example <client_id>/<id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
