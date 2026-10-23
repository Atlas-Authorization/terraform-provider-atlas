# Put the accepted version of every required policy on session tokens as the
# `pol` claim, e.g. {"tos": "2026-10-24"}.
resource "atlas_policy_settings" "this" {
  session_claim = true
}
