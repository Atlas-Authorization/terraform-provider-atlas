# Create a bundle: several catalog entitlements, request form fields, and a policy (who may ask, guest expiry).
resource "atlas_iga_bundle" "example" {
  key = "tf-iga-bundle"
  name = "tf-iga-bundle"
  entitlements = ["example"]
}
