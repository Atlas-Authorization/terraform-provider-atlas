# Enrol a principal public key, binding it to an Atlas identity (user, org or machine).
resource "atlas_pca_principal" "example" {
  principal_pub = "<base64url Ed25519 public key>"
  subject_type  = "user"
  subject_id    = "user_123"
}
