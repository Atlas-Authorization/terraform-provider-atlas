# Attach an app (its public client_id) to a verified publisher.
resource "atlas_oauth_publisher_client" "example" {
  publisher_id = "tf-oauth-publisher-client"
  client_id = "tf-oauth-publisher-client"
}
