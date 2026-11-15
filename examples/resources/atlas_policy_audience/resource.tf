# Set a policy's audience.
resource "atlas_policy_audience" "example" {
  policy_id = "tf-policy-audience"
  audience = "everyone"
}
