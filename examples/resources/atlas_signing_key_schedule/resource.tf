# Rotate the instance signing key automatically.
resource "atlas_signing_key_schedule" "example" {
  interval_days = 1
}
