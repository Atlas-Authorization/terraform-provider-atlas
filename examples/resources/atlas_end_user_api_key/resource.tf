# A key-pair API key for a user. Bring your own public key so no private key is ever written to Terraform state.
resource "tls_private_key" "partner" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P256"
}

resource "atlas_end_user_api_key" "partner" {
  subject_type = "user"
  subject_id   = "user_123"
  scopes       = ["orders:read"]
  credential   = "key_pair"
  audience     = "https://api.example.com"
  public_key   = tls_private_key.partner.public_key_pem
}

output "key_id" {
  value = atlas_end_user_api_key.partner.key_id
}
