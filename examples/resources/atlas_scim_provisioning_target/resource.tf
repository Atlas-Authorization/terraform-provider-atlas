# Create an outbound SCIM provisioning target, optionally from a preset (preset + preset_params instead of base_url).
resource "atlas_scim_provisioning_target" "example" {
  name = "tf-scim-provisioning-target"
}
