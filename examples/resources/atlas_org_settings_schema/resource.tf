# The per-instance org-settings JSON Schema registry (a singleton). Atlas
# validates every per-organization settings write against it and versions it.
resource "atlas_org_settings_schema" "default" {
  schema = jsonencode({
    type = "object"
    properties = {
      brand_color   = { type = "string" }
      support_email = { type = "string", format = "email" }
      seats_limit   = { type = "integer", minimum = 1 }
    }
    required = ["support_email"]
  })
}
