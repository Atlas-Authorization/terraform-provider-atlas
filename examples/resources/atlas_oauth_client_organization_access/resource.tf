# Set who may use an application.
resource "atlas_oauth_client_organization_access" "example" {
  client_id = "tf-oauth-client-organization-access"
  mode = "open"
}
