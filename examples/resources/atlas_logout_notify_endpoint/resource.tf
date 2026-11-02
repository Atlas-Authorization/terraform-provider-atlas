# Register a logout fan-out endpoint.
resource "atlas_logout_notify_endpoint" "example" {
  url = "https://example.com/url"
}
