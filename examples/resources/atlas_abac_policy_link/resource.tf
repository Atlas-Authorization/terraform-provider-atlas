# Link a template to entities.
resource "atlas_abac_policy_link" "example" {
  template_id = "tf-abac-policy-link"
  name = "tf-abac-policy-link"
}
