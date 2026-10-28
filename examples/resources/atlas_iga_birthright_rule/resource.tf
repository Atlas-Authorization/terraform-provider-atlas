# Create a birthright rule: everyone matching `match` (a group, organization and role, app role or verified email domain; all conditions must hold) holds the listed entitlements.
resource "atlas_iga_birthright_rule" "example" {
  name = "tf-iga-birthright-rule"
  match = jsonencode({})
  entitlement_ids = ["example"]
}
