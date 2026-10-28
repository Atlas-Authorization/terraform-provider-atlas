# Define a credential type: its vct, and how each claim is filled from a user’s attributes.
resource "atlas_vc_credential_type" "example" {
  vct = "tf-vc-credential-type"
  name = "tf-vc-credential-type"
  claims = jsonencode({})
}
