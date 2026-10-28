# Trust an external issuer for this client.
resource "atlas_oauth_client_trusted_issuer" "example" {
  client_id = "tf-oauth-client-trusted-issuer"
  issuer = "https://example.com/issuer"
}
