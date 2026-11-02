# Create a rule.
resource "atlas_iga_sod_rule" "example" {
  name = "tf-iga-sod-rule"
  entitlement_a = "tf-iga-sod-rule"
  entitlement_b = "tf-iga-sod-rule"
}
