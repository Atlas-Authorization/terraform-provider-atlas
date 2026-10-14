# The hosted sign-in page design for the instance.
resource "atlas_instance_design" "this" {
  design = jsonencode({
    preset      = "midnight"
    colorScheme = "auto"
    palette     = { primary = "#4f46e5" }
    shape       = { radius = 12 }
  })
}
