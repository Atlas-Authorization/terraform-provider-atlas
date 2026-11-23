# Configure the managed sync that turns organization memberships, org roles and groups into FGA tuples, so a model can say `viewer: member from organization` without mirroring anything by hand.
resource "atlas_fga_sync" "example" {

}
