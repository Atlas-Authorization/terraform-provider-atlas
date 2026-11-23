# Create an organization template: custom roles (with existing permissions), policy, role set and default role, seat limit and warning level, and metadata.
resource "atlas_org_template" "example" {
  key = "tf-org-template"
  name = "tf-org-template"
}
