# Issue a one-time code that lets an instance of the account adopt this app with its own secret key.
resource "atlas_id_app_adoption_code" "example" {
  atlas_id_app_id = "tf-atlas-id-app-adoption-code"
}
