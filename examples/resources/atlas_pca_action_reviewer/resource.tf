# A reviewer of agent action changes.
resource "atlas_pca_action_reviewer" "example" {
  display_name = "Security reviewer"
  email        = "reviewer@example.com"
}
