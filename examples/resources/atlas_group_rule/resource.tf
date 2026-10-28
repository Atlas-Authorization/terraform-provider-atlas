# Create a rule: when a user matches condition, run the actions (add to a group, grant an app role, set a trait).
resource "atlas_group_rule" "example" {
  name = "tf-group-rule"
  condition = jsonencode({})
  actions = jsonencode({})
}
