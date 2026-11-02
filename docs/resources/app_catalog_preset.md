# atlas_app_catalog_preset

Save a private SAML preset for apps that are not in the catalog (an id, name, entityIdTemplate and acsUrlTemplate using {param} placeholders, params with a kind of slug, host, https_url or entity, and an attribute map). It then appears in GET /v1/app_catalog (custom: true, never verified) and installs like any preset. Saving an existing id replaces it.

Scopes: `app_catalog:read` to read, `app_catalog:write` to write.

Backend API: `POST /v1/app_catalog/custom_presets` (create); `GET /v1/app_catalog/custom_presets` (read: the list is scanned for the id); no update route: every change replaces the object; `DELETE /v1/app_catalog/custom_presets/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_app_catalog_preset" "example" {
  preset = jsonencode({})
}
```

## Schema

### Required

- `preset` (String, JSON) id, name, category, iconDomain, entityIdTemplate, acsUrlTemplate, params[{key,label,kind}], nameIdField, attributes, source. A JSON document (use `jsonencode`). Not returned by the API, so it is not detected as drift.

### Read-Only

- `id` (String) Atlas object id.
- `name` (String) The `name` field.

## Import

The import id is `<id>`.

```sh
terraform import atlas_app_catalog_preset.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
