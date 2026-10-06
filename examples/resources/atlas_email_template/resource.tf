# Override the copy of a single built-in transactional email. Atlas validates
# {{placeholders}} at save time; destroying reverts to the built-in copy.
resource "atlas_email_template" "verification" {
  name    = "verification_code"
  subject = "Your {{app_name}} verification code"
  text    = "Your code is {{code}}. It expires in {{ttl_minutes}} minutes."
}
