# Create a custom OAuth 2.0 / OIDC provider (Atlas is the relying party).
resource "atlas_custom_oauth_provider" "example" {
  provider_key = "tf-custom-oauth-provider"
  display_name = "tf-custom-oauth-provider"
}
