# Register a native app that signs users in (RFC 8252): an `ios` app (bundle_id + team_id), an `android` app (package_name + sha256_cert_fingerprints) or a `development` app for Expo Go and dev builds (allowed_users + expires_at, at most 90 days ahead; it stops working at expiry and expires_at can be extended in place).
resource "atlas_native_app" "example" {
  platform      = "ios"
  bundle_id     = "com.example.app"
  team_id       = "ABCDE12345"
  callback_urls = ["com.example.app://auth/callback"]
}
