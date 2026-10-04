terraform {
  required_providers {
    atlas = {
      source = "Atlas-Authorization/atlas"
    }
  }
}

# The provider reads its credentials from the environment by default:
#   export ATLAS_SECRET_KEY=sk_live_...
#   export ATLAS_API_URL=https://api.atlasauth.net   # optional; this is the default
#
# You may also set them inline (api_url shown; keep secret_key in the env):
provider "atlas" {
  api_url = "https://api.atlasauth.net"
  # secret_key = "sk_live_..."  # prefer ATLAS_SECRET_KEY
}

# ── A "Sign in with Atlas" relying party ─────────────────────────────────────
resource "atlas_oauth_client" "dashboard" {
  name                       = "Acme Dashboard"
  redirect_uris              = ["https://app.acme.example.com/callback"]
  allowed_scopes             = ["openid", "profile", "email"]
  grant_types                = ["authorization_code", "refresh_token"]
  token_endpoint_auth_method = "client_secret_basic"
  logo_uri                   = "https://app.acme.example.com/logo.png"
}

# The confidential client secret is revealed once, on create, and stored
# (sensitive) in state. Consume it from another resource, not by printing it.
output "dashboard_client_id" {
  value = atlas_oauth_client.dashboard.client_id
}

# ── An enterprise SSO connection (OIDC) ──────────────────────────────────────
resource "atlas_sso_connection" "okta" {
  type               = "oidc"
  status             = "active"
  oidc_issuer        = "https://acme.okta.com"
  oidc_client_id     = "0oa1example"
  oidc_client_secret = var.okta_client_secret # write-only; never read back
  allowed_domains    = ["acme.example.com"]

  claim_role_mappings {
    claim    = "groups"
    value    = "Engineering"
    role_key = "org:member"
  }
}

variable "okta_client_secret" {
  type      = string
  sensitive = true
}

# ── A social sign-in provider (GitHub) ───────────────────────────────────────
# Configures a provider from the catalog (~193 keys: google, github, ...). The
# client secret is write-only. The redirect_uri is derived from your instance
# host — register that exact string at the provider console.
resource "atlas_oauth_provider" "github" {
  provider_key = "github" # not `provider` — that is a reserved TF meta-argument

  credentials = {
    client_id     = "Iv1.abc123"
    client_secret = var.github_client_secret # write-only; never read back
  }

  scopes        = ["read:user", "user:email"]
  enabled       = true
  allow_sign_in = true
  allow_sign_up = true
}

variable "github_client_secret" {
  type      = string
  sensitive = true
}

output "github_redirect_uri" {
  value = atlas_oauth_provider.github.redirect_uri
}

# ── A resource server (API) for machine-to-machine tokens ────────────────────
resource "atlas_resource_server" "billing_api" {
  identifier        = "https://billing.acme.example.com"
  name              = "Billing API"
  token_ttl_seconds = 3600
  signing_alg       = "RS256"

  scopes {
    value       = "read:invoices"
    description = "Read invoices"
  }
  scopes {
    value       = "write:invoices"
    description = "Create and modify invoices"
  }
}

# ── A JWT template (custom claims injected into issued tokens) ────────────────
resource "atlas_jwt_template" "supabase" {
  name = "supabase"
  claims = {
    role      = "authenticated"
    user_role = "{{user.public_metadata.role}}"
  }
}

# ── A webhook endpoint (signing secret revealed once, on create) ─────────────
resource "atlas_webhook_endpoint" "events" {
  url            = "https://hooks.acme.example.com/atlas"
  enabled_events = ["user.created", "user.updated", "organization.created"]
}

# ── A custom organization role ───────────────────────────────────────────────
resource "atlas_role" "billing_admin" {
  key         = "billing_admin"
  name        = "Billing Admin"
  description = "Manages billing and invoices"
  permissions = ["org:billing:manage", "org:invoices:write"]
}

# ── An organization (created_by must be an existing user id) ──────────────────
resource "atlas_organization" "acme" {
  name                    = "Acme, Inc."
  slug                    = "acme"
  created_by              = var.founder_user_id
  max_allowed_memberships = 50
}

variable "founder_user_id" {
  type = string
}

# ── A custom instance domain (point a CNAME at cname_target, then verify) ─────
resource "atlas_domain" "fapi" {
  role = "fapi"
  host = "auth.acme.example.com"
}

output "fapi_cname_target" {
  value = atlas_domain.fapi.cname_target
}

# ── Data sources: read existing config by id ─────────────────────────────────
data "atlas_oauth_client" "existing" {
  id = atlas_oauth_client.dashboard.id
}

data "atlas_resource_server" "existing" {
  id = atlas_resource_server.billing_api.id
}

data "atlas_oauth_provider" "google" {
  provider_key = "google"
}
