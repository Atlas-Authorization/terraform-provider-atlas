# A sales-led Enterprise grant for one organization, without Stripe.
resource "atlas_entitlement_override" "acme_enterprise" {
  subject_type = "org"
  subject_id   = atlas_organization.acme.id
  plan         = "enterprise"
  features     = ["sso"]
  limits       = { seats = 500 }
}
