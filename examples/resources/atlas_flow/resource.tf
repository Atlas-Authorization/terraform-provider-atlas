# Create a sign-in journey (version 1, active unless activate is false).
resource "atlas_flow" "example" {
  name = "tf-flow"
  definition = jsonencode({})
}
