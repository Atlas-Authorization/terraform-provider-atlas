# Pin a namespace to one stored version: that version is active and later publishes do not replace it.
resource "atlas_pca_action_namespace_pin" "example" {
  ns      = "billing"
  version = "1.4.0"
}
