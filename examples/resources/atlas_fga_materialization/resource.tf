# Precompute list-objects for every user of `user_type` on `object_type#relation`, so a huge list question is one indexed read.
resource "atlas_fga_materialization" "example" {
  store_id = "tf-fga-materialization"
}
