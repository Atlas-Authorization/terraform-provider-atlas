# atlas_wsfed_connection

Add a WS-Federation connection. Give the provider's federation metadata (metadata_url or metadata_xml) or its issuer, passive endpoint and signing certificates, and the realm. Sign-in starts at POST /v1/client/sign_ins/wsfed.

Scopes: `sso_connections:read` to read, `sso_connections:write` to write.

Backend API: `POST /v1/wsfed_connections` (create); `GET /v1/wsfed_connections/{id}` (read); `PATCH /v1/wsfed_connections/{id}` (update); `DELETE /v1/wsfed_connections/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_wsfed_connection" "example" {

}
```

## Schema

### Optional

- `organization_id` (String) The `organization_id` field.
- `status` (String) The `status` field. One of: `draft`, `active`, `disabled`.
- `metadata_url` (String) An https URL serving the provider's federation metadata (for AD FS /FederationMetadata/2007-06/FederationMetadata.xml); fills issuer, passive_endpoint and signing_certificates. Not returned by the API, so it is not detected as drift.
- `metadata_xml` (String) The federation metadata document itself. Not returned by the API, so it is not detected as drift.
- `issuer` (String) The `issuer` field.
- `passive_endpoint` (String) The `passive_endpoint` field.
- `signing_certificates` (String) One to five PEM certificates (several during a rollover; a token verifies against any of them). Not returned by the API, so it is not detected as drift.
- `realm` (String) The `realm` field.
- `allowed_domains` (List of String) The `allowed_domains` field.
- `attribute_mapping` (String, JSON) Claim URI to profile field: email, email_verified, first_name, last_name, username. A JSON document (use `jsonencode`).
- `freshness_minutes` (Number) The `freshness_minutes` field.
- `auth_method` (String) The `auth_method` field.

### Read-Only

- `id` (String) Atlas object id.
- `signing_certificate_count` (Number) The `signing_certificate_count` field.
- `acs_url` (String) Enter this as the relying-party trust's WS-Federation passive endpoint (the reply URL).
- `logout_start_url` (String) The `logout_start_url` field.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_wsfed_connection.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
