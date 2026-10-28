# Replace the stored assertions for a model version.
resource "atlas_fga_assertions" "example" {
  store_id = "tf-fga-assertions"
  model_id = "tf-fga-assertions"
}
