# Give an organization access to an application (by client id or id), optionally limited to some organization roles.
resource "atlas_organization_application" "example" {
  organization_id = "tf-organization-application"
  client_id = "tf-organization-application"
}
