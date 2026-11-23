# Add a fallback provider (up to 5 per channel).
resource "atlas_messaging_provider_fallback" "example" {
  channel = "tf-messaging-provider-fallback"
  transport = "tf-messaging-provider-fallback"
}
