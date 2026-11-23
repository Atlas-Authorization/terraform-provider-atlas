# Notification channels, contact details and status of one member of an existing approver.
resource "atlas_pca_approver_member" "example" {
  approver_id = "pcaapp_example"
  member_id   = "pcamem_example"
  channels    = ["email"]
}
