# Migrating from Better Auth without signing anyone out: while the bridge is
# open, a returning user's Better Auth session is exchanged once for an Atlas
# session.
variable "better_auth_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "atlas_better_auth_bridge" "this" {
  enabled       = true
  ends_at       = "2026-12-31T00:00:00Z"
  cookie_prefix = "better-auth"

  # Sessions created after the export: ask the old app, server to server.
  live_fallback = {
    enabled           = true
    base_url          = "https://app.example.com/api/auth"
    secret_wo         = var.better_auth_secret # write-only, never in state
    secret_wo_version = 1                      # bump to send a new secret
  }
}
