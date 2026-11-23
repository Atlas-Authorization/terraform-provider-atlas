# Human-approval settings for agent actions: the approval channels, the default channels for approver members, the CIBA backchannel and the Slack webhook.
resource "atlas_pca_approvals_settings" "example" {
  channels = jsonencode(["email"])
}
