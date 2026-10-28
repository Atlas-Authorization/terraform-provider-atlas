# Set an agent's sponsor, backup sponsor and expiry (omitted fields keep their value; null clears).
resource "atlas_agent_lifecycle" "example" {
  agent_id = "tf-agent-lifecycle"
}
