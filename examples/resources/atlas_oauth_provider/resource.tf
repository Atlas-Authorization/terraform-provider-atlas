# Ethereum wallet sign-in (no credentials; verifier settings only).
resource "atlas_oauth_provider" "ethereum" {
  provider_key = "siwe"
  settings = {
    domain    = "app.example.com"
    chain_ids = "1,10"
  }
  display_name = "Ethereum wallet"
  enabled      = true
}
