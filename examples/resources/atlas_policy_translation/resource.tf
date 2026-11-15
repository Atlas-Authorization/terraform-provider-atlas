# Add or replace the text of a version in one language (content_markdown or content_html, sanitised like the version itself).
resource "atlas_policy_translation" "example" {
  policy_id = "tf-policy-translation"
  version_id = "tf-policy-translation"
  locale = "tf-policy-translation"
}
