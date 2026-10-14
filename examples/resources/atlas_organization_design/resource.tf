# An organization's own colours on its hosted pages.
resource "atlas_organization_design" "acme" {
  organization_id = atlas_organization.acme.id
  design = jsonencode({
    palette = { primary = "#0f766e" }
  })
}
