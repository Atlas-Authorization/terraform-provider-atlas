# Register a custom subject_token_type (a urn: or https URI that is not a standard Atlas or IETF token type) for a confidential client.
resource "atlas_oauth_client_token_exchange_profile" "example" {
  client_id = "tf-oauth-client-token-exchange-profile"
}
