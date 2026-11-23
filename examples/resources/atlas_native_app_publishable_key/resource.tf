# Restrict which registered native apps may use the publishable key (an empty list means any).
resource "atlas_native_app_publishable_key" "example" {
  app_ids = ["nativeApp_example"]
}
