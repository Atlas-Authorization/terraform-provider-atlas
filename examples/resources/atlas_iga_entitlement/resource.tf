# Add an entitlement to the catalog: a role, organization membership, group, resource role, installed-app assignment or custom item, with an owner, a risk level and approval rules.
resource "atlas_iga_entitlement" "example" {
  key = "tf-iga-entitlement"
  name = "tf-iga-entitlement"
  kind = "role"
  target = jsonencode({})
}
