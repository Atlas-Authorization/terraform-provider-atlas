# Point an organization at a role set and/or give it its own default role (null clears either).
resource "atlas_organization_role_policy" "example" {
  organization_id = "tf-organization-role-policy"
}
