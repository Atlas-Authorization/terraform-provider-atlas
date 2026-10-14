# Override the copy of a single built-in transactional email. Atlas validates
# {{placeholders}} at save time; destroying reverts to the built-in copy.
resource "atlas_email_template" "verification" {
  name    = "verification_code"
  subject = "Your verification code"
  text    = "Your code is {{code}}. Enter it to finish signing in."
}
