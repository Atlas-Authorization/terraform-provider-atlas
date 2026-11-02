# atlas_sso_connection

An enterprise SSO connection (`oidc`, `saml` or `discourse`). Secrets — the OIDC
client secret, the SAML IdP certificate and the Discourse shared secret — are
write-only: the API never returns them, so presence is tracked only via the
`has_*` booleans and by whatever is set in configuration.

Backend API (BAPI): `POST /v1/sso_connections` (create),
`GET /v1/sso_connections/:id` (read), `PATCH /v1/sso_connections/:id` (update),
`DELETE /v1/sso_connections/:id` (delete).

## Example Usage

```hcl
resource "atlas_sso_connection" "acme_oidc" {
  type               = "oidc"
  status             = "active"
  oidc_issuer        = "https://acme.okta.com"
  oidc_client_id     = "0oa1abcd"
  oidc_client_secret = var.acme_oidc_secret
  allowed_domains    = ["acme.com"]
}
```

## Schema

### Required

- `type` (String) — Connection protocol: `oidc`, `saml` or `discourse`. Immutable — forces replacement.

### Optional

- `status` (String) — `draft`, `active` or `disabled`. Activation requires the protocol's mandatory fields.
- `organization_id` (String) — Organization this connection is scoped to.
- `oidc_issuer` (String) — OIDC issuer URL.
- `oidc_client_id` (String) — OIDC client id at the identity provider.
- `oidc_client_secret` (String, Sensitive) — OIDC client secret. Write-only; never returned by the API.
- `saml_idp_entity_id` (String) — SAML IdP entity id.
- `saml_idp_sso_url` (String) — SAML IdP single sign-on URL.
- `saml_idp_certificate` (String, Sensitive) — SAML IdP signing certificate (PEM). Write-only.
- `saml_sp_entity_id` (String) — SAML service-provider entity id.
- `saml_allow_idp_initiated` (Boolean) — Allow IdP-initiated SAML sign-in.
- `saml_sign_authn_requests` (Boolean) — Sign outgoing SAML AuthnRequests.
- `saml_want_response_signed` (Boolean) — Require the SAML response itself to be signed.
- `discourse_secret` (String, Sensitive) — Discourse SSO shared secret. Write-only.
- `discourse_provider_url` (String) — Discourse provider URL.
- `allowed_domains` (List of String) — Email domains this connection applies to.
- `default_role_id` (String) — Role granted when no claim mapping matches.
- `claim_role_mappings` (Block List) — Map an IdP claim value onto an Atlas role key. Each block has `claim`, `value` and `role_key` (all required).

### Read-Only

- `id` (String) — Atlas object id.
- `has_secret` (Boolean) — Whether an OIDC client secret is stored.
- `has_saml_certificate` (Boolean) — Whether a SAML IdP certificate is stored.
- `has_discourse_secret` (Boolean) — Whether a Discourse shared secret is stored.
- `created_at` (Number) — Creation time (epoch milliseconds).
- `updated_at` (Number) — Last update time (epoch milliseconds).

## Import

```sh
terraform import atlas_sso_connection.acme_oidc ssoc_123
```

Write-only secrets are not recovered on import.
