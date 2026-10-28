# Create an app-level user role: a role not tied to any organization, granting permission keys.
resource "atlas_app_role" "example" {
  key = "tf-app-role"
  name = "tf-app-role"
}
