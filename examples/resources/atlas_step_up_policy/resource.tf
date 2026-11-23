# Create a named step-up policy: how recent and how strong a re-authentication a sensitive action needs.
resource "atlas_step_up_policy" "example" {
  name = "tf-step-up-policy"
}
