# A partner platform: a product that signs the events it delivers to this
# instance. The token is write-only (Terraform 1.11 or later); change
# token_wo_version to send a new one.
resource "atlas_partner_platform" "consent" {
  partner          = "consentco"
  platform_id      = "atlas"
  jwks_url         = "https://consentco.example/.well-known/jwks.json"
  signature_header = "x-consentco-signature"
  allowed_sources  = ["site-123"]

  token_wo         = var.partner_token
  token_wo_version = 1
}
