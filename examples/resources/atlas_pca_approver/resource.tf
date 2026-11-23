# Settings of a passkey approver that already exists: its quorum, display name and status.
resource "atlas_pca_approver" "example" {
  approver_id = "pcaapp_example"
  quorum      = 1
}
