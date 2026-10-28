# atlas_connector

Create a community connector profile (SEO profile + OAuth config + share modes). Slug is globally unique and may not shadow a built-in.

Scopes: `connectors:read` to read, `connectors:write` to write.

Backend API: `POST /v1/connectors` (create); `GET /v1/connectors/{id}` (read); `PATCH /v1/connectors/{id}` (update); `DELETE /v1/connectors/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_connector" "example" {
  slug = "tf-connector"
  display_name = "tf-connector"
}
```

## Schema

### Required

- `slug` (String) The `slug` field.
- `display_name` (String) The `display_name` field.

### Optional

- `logo_url` (String) The `logo_url` field.
- `tagline` (String) The `tagline` field.
- `description` (String) The `description` field.
- `category` (String) The `category` field.
- `tags` (List of String) The `tags` field.
- `website_url` (String) The `website_url` field.
- `docs_url` (String) The `docs_url` field.
- `support_url` (String) The `support_url` field.
- `screenshots` (List of String) The `screenshots` field. Not returned by the API, so it is not detected as drift.
- `setup_guide` (String) The `setup_guide` field.
- `protocol` (String) The `protocol` field.
- `authorization_endpoint` (String) The `authorization_endpoint` field.
- `token_endpoint` (String) The `token_endpoint` field.
- `userinfo_endpoint` (String) The `userinfo_endpoint` field.
- `issuer` (String) The `issuer` field.
- `default_scopes` (List of String) The `default_scopes` field.
- `api_scopes` (String, JSON) The `api_scopes` field. A JSON document (use `jsonencode`).
- `credential_fields` (String, JSON) The `credential_fields` field. A JSON document (use `jsonencode`).
- `authorization_params` (String, JSON) The `authorization_params` field. A JSON document (use `jsonencode`). Not returned by the API, so it is not detected as drift.
- `quirks` (String, JSON) The `quirks` field. A JSON document (use `jsonencode`). Not returned by the API, so it is not detected as drift.
- `share_modes` (List of String) The `share_modes` field. One of: `prefill`, `full_shared`, `full_idp`.
- `idp_instance_id` (String) The `idp_instance_id` field. Not returned by the API, so it is not detected as drift.
- `idp_issuer` (String) The `idp_issuer` field.
- `idp_auto_register` (Boolean) The `idp_auto_register` field.

### Read-Only

- `id` (String) Atlas object id.
- `shared_client_id` (String) The `shared_client_id` field.
- `shared_revoked` (Boolean) The `shared_revoked` field.
- `status` (String) The `status` field. One of: `draft`, `published`, `unlisted`, `banned`.
- `verified` (Boolean) The `verified` field.
- `featured` (Boolean) The `featured` field.
- `install_count` (Number) The `install_count` field.
- `view_count` (Number) The `view_count` field.
- `rating_avg` (Number) The `rating_avg` field.
- `rating_count` (Number) The `rating_count` field.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Last updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_connector.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
