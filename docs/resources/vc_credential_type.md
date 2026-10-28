# atlas_vc_credential_type

Define a credential type: its vct, and how each claim is filled from a user’s attributes. Issued as SD-JWT VC.

Scopes: `vc:read` to read, `vc:write` to write.

Backend API: `POST /v1/vc/credential_types` (create); `GET /v1/vc/credential_types/{id}` (read); no update route: every change replaces the object; `DELETE /v1/vc/credential_types/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_vc_credential_type" "example" {
  vct = "tf-vc-credential-type"
  name = "tf-vc-credential-type"
  claims = jsonencode({})
}
```

## Schema

### Required

- `vct` (String) The credential type identifier (a URN or URL); for mso_mdoc the docType, for example org.iso.18013.5.1.mDL.
- `name` (String) The `name` field.
- `claims` (String, JSON) The `claims` field. A JSON document (use `jsonencode`).

### Optional

- `format` (String) Default dc+sd-jwt. mso_mdoc issues an ISO 18013-5 mdoc (ES256, holder key P-256).
- `namespace` (String) mso_mdoc only: the namespace the claims live in (default the docType).
- `iaca_roots` (List of String) mso_mdoc only: PEM IACA root certificates (ISO 18013-5 Annex B). With roots set, a presentation must carry an x5chain leading to one of them and the issuer signature must verify under its Document Signer key; bare trusted keys are then not accepted.
- `revocation_policy` (String) mso_mdoc only: off, best_effort (default) or required. See the credential type object.
- `signer_chain` (List of String) mso_mdoc only: PEM Document Signer certificate first, then intermediates, published as x5chain. The first certificate must be for this instance’s mdoc signing key (the ES256 key in its key set), otherwise 400.
- `lifetime_seconds` (Number) Default one year.
- `require_wallet_attestation` (Boolean) The `require_wallet_attestation` field.
- `wallet_attestation_jwks` (String, JSON) JWK Set of accepted wallet providers; required with require_wallet_attestation. A JSON document (use `jsonencode`). Not returned by the API, so it is not detected as drift.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_vc_credential_type.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
