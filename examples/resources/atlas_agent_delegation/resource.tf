# Delegate authority to an agent.
resource "atlas_agent_delegation" "example" {
  child_machine_id = "tf-agent-delegation"
  scopes = ["example"]
}
