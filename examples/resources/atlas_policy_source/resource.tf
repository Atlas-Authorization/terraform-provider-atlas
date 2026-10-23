variable "policy_source_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

# Serve the policy text from your own system. Atlas fetches
# GET {endpoint_url}/policies, signed with the secret; your system sends signed
# policy.published / policy.retired events to the computed webhook_url.
resource "atlas_policy_source" "this" {
  enabled      = true
  endpoint_url = "https://cms.example.com/atlas"
  sync         = "webhook"

  signing_secret_wo         = var.policy_source_secret # write-only, never in state
  signing_secret_wo_version = 1                        # bump to send a new secret
}

output "policy_webhook_url" {
  value = atlas_policy_source.this.webhook_url
}
