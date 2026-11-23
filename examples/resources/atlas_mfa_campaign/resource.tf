# Start an MFA rollout campaign: ask everyone, a group, an organization or an organization role to enrol a second factor.
resource "atlas_mfa_campaign" "example" {
  name = "tf-mfa-campaign"
}
