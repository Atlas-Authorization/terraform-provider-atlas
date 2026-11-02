# Create a feature flag: a typed value (`boolean`, `string`, `int` or `json`) with a `default_value`, an `off_value` (served while `enabled` is false; defaults to false for a boolean, else the default) and ordered targeting `rules` by org, user, plan and percentage.
resource "atlas_feature_flag" "example" {
  key = "tf-feature-flag"
  type = "boolean"
  default_value = true
}
