# Mint a SCIM provisioning token for an organization.
resource "atlas_scim_token" "example" {
  organization_id = "tf-scim-token"
}
