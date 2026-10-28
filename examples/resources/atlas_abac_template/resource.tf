# Create a policy template: policy source whose scope uses the slots `?principal` and/or `?resource` (`principal == ?principal`, `resource in ?resource`, `resource is Doc in ?resource`).
resource "atlas_abac_template" "example" {
  name = "tf-abac-template"
  source = "tf-abac-template"
}
