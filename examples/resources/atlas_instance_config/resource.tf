# The singleton instance configuration: allowed origins + auth config.
resource "atlas_instance_config" "this" {
  allowed_origins = [
    "https://app.acme.example.com",
    "https://admin.acme.example.com",
  ]

  # auth_config is a partial patch merged server-side; manage the object you set.
  auth_config = jsonencode({
    session = {
      inactivityTimeout = 1800000 # 30 minutes, in ms
    }
  })
}

# The effective (defaults-expanded) auth config, for inspection.
output "effective_auth_config" {
  value = atlas_instance_config.this.auth_config_resolved
}
