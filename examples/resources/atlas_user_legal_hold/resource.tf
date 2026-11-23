# Place a legal hold on a user.
resource "atlas_user_legal_hold" "example" {
  user_id = "tf-user-legal-hold"
  reason = "tf-user-legal-hold"
}
