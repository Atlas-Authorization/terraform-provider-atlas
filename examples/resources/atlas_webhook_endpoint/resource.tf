# A webhook endpoint. Change rotate_secret_trigger to rotate the signing secret.
resource "atlas_webhook_endpoint" "events" {
  url                   = "https://hooks.example.com/atlas"
  enabled_events        = ["user.created", "organization.updated"]
  rotate_secret_trigger = "2026-10"
}
