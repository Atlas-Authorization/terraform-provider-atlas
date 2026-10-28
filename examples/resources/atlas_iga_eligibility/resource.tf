# Make a person eligible for an entitlement: they hold nothing until they activate it.
resource "atlas_iga_eligibility" "example" {
  user_id = "tf-iga-eligibility"
  entitlement = "tf-iga-eligibility"
}
