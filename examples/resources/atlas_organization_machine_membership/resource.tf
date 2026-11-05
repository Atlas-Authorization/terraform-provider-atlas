# Make a machine a member of an organization with a role.
resource "atlas_organization_machine_membership" "example" {
  organization_id = "tf-organization-machine-membership"
}
