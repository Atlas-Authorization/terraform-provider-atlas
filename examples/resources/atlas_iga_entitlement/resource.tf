resource "atlas_iga_entitlement" "prod_admin" {
  key           = "prod-admin"
  name          = "Production admin"
  kind          = "role"
  target        = jsonencode({ role_key = "admin" })
  owner_user_id = "user_01J8EXAMPLE"
  risk_level    = "high"

  approval_rules = jsonencode({
    steps = [
      { type = "manager", mode = "any", timeout_seconds = 86400, escalate_to = { type = "owner" } },
      { type = "group", ids = ["secops"], mode = "all" },
    ]
  })

  max_duration_seconds = 28800
}
