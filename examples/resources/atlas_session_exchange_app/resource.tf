# Register a session-exchange app and its backchannel_logout_uri.
resource "atlas_session_exchange_app" "example" {
  name = "tf-session-exchange-app"
  backchannel_logout_uri = "https://example.com/backchannel_logout_uri"
}
