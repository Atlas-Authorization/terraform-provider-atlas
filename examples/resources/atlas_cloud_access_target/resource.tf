# Add a cloud access target.
resource "atlas_cloud_access_target" "example" {
  provider_value = "aws"
  name = "tf-cloud-access-target"
  config = jsonencode({})
}
