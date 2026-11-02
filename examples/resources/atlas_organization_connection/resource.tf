# Invite a partner organization to use an application the supplier organization holds.
resource "atlas_organization_connection" "example" {
  supplier_organization_id = "tf-organization-connection"
  partner_organization_id = "tf-organization-connection"
  application = "tf-organization-connection"
}
