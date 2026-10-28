# Create a draft campaign scoped to everything, specific entitlements, or the entitlements for an app, role, group or organization, with reviewers (the person's manager, the entitlement owner, or named users), a due date and reminder days.
resource "atlas_iga_campaign" "example" {
  name = "tf-iga-campaign"
}
