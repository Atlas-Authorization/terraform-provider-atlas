# Save a private SAML preset for apps that are not in the catalog (an id, name, entityIdTemplate and acsUrlTemplate using {param} placeholders, params with a kind of slug, host, https_url or entity, and an attribute map).
resource "atlas_app_catalog_preset" "example" {
  preset = jsonencode({})
}
