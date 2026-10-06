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

  # Native-app passkeys, as a typed block (folded into the auth_config patch
  # under `nativeApps`). Atlas serves the matching apple-app-site-association and
  # assetlinks.json on the instance Frontend API host. This takes precedence over
  # any nativeApps set in the raw auth_config above; you may still set nativeApps
  # in raw JSON instead of using this block.
  native_apps = {
    apple_app_ids = ["LB4397Q8XJ.com.acme.app"]
    android_apps = [
      {
        package_name             = "com.acme.app"
        sha256_cert_fingerprints = ["AB:CD:EF:...:12:34"] # keytool/Play SHA-256
      },
    ]
  }
}

# The effective (defaults-expanded) auth config, for inspection.
output "effective_auth_config" {
  value = atlas_instance_config.this.auth_config_resolved
}
