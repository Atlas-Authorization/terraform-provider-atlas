# Add one or many addresses (a bounce or complaint webhook forwarder can post here).
resource "atlas_email_suppression" "example" {
  email = "user@example.com"
}
