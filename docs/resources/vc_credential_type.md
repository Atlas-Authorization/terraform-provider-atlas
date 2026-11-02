# atlas_vc_credential_type

A verifiable credential type Atlas issues to wallets (OpenID4VCI) and verifies
(OpenID4VP): an SD-JWT VC or an ISO 18013-5 mdoc, with each claim filled from a
user attribute. A credential type is immutable: any change forces replacement.

Backend API (BAPI): `POST /v1/vc/credential_types` (create),
`GET /v1/vc/credential_types/:id` (read), `DELETE /v1/vc/credential_types/:id`
(delete). The API key needs the `vc:read` and `vc:write` scopes.

## Example Usage

```hcl
resource "atlas_vc_credential_type" "mdl" {
  name      = "Driving licence"
  format    = "mso_mdoc"
  vct       = "org.iso.18013.5.1.mDL"
  namespace = "org.iso.18013.5.1"

  claims = [
    { name = "family_name", source = "last_name" },
    { name = "birth_date", source = "public_metadata.birth_date", value_type = "full-date" },
  ]
}
```

## Schema

### Required

- `vct` (String) — The credential type identifier (a URN or URL); for `mso_mdoc` the docType, for example `org.iso.18013.5.1.mDL`.
- `name` (String) — Display name (1 to 120 characters).
- `claims` (List of Object) — The claims of the credential, in order. Each has `name` (Required), `source` (Required: `user_id`, `email`, `first_name`, `last_name`, `username` or `public_metadata.<key>`), `disclosable` (Optional, default `true`) and `value_type` (Optional: `string`, `number`, `boolean` or `full-date`).

### Optional

- `format` (String) — `dc+sd-jwt` (default) or `mso_mdoc`.
- `namespace` (String) — `mso_mdoc` only: the namespace the claims live in. Defaults to the docType.
- `iaca_roots` (List of String) — `mso_mdoc` only: PEM IACA root certificates (ISO 18013-5 Annex B). With roots set, a presentation must carry an `x5chain` leading to one of them and the issuer signature must verify under its Document Signer key; bare trusted keys are then not accepted.
- `signer_chain` (List of String) — `mso_mdoc` only: PEM Document Signer certificate first, then intermediates, published as `x5chain` in issued mdocs. The first certificate must be for the instance's mdoc signing key (its ES256 key).
- `lifetime_seconds` (Number) — How long an issued credential is valid, in seconds (60 to 315360000). Defaults to one year.

### Read-Only

- `id` (String) — Atlas object id; also the credential configuration id in the issuer metadata.
- `require_wallet_attestation` (Boolean) — Whether the wallet must present an attestation. Set with the Backend API.

## Import

```sh
terraform import atlas_vc_credential_type.mdl vct_123
```
