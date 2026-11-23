# Pin a publisher key: action packs signed by it are accepted into the registry.
resource "atlas_pca_action_publisher_key" "example" {
  kid          = "acme-publisher-1"
  pk           = "<base64url Ed25519 public key>"
  display_name = "Acme publisher"
}
