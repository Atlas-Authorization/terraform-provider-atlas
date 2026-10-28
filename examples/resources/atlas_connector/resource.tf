# Create a community connector profile (SEO profile + OAuth config + share modes).
resource "atlas_connector" "example" {
  slug = "tf-connector"
  display_name = "tf-connector"
}
