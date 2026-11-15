# Claim a publisher name for a domain.
resource "atlas_oauth_publisher" "example" {
  name = "tf-oauth-publisher"
  domain = "tf-oauth-publisher"
}
