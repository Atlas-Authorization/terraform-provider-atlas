# Set the seat limit (max_allowed_memberships; null = unlimited) and the warning level.
resource "atlas_organization_seats" "example" {
  organization_id = "tf-organization-seats"
}
