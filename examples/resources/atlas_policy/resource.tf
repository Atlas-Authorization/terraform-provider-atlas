# Terms of service everyone accepts at sign-up, and existing users re-accept
# (within 14 days) whenever a new version takes effect.
resource "atlas_policy" "tos" {
  key        = "tos"
  title      = "Terms of service"
  display    = "scroll" # an unskippable reader
  required   = true
  applies_to = "sign_in_and_sign_up"
  reaccept   = "grace"
  grace_days = 14
}

# A privacy policy accepted once, at sign-up, with a checkbox.
resource "atlas_policy" "privacy" {
  key   = "privacy"
  title = "Privacy policy"
}

# The policy asks nobody anything until a version is published. Versions are
# content: publish them from the dashboard, the Backend API or SDKs, or serve
# them from your own system with atlas_policy_source.
