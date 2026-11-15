# atlas_agent_delegation

Delegate authority to an agent. A root delegation (delegator_user_id) lets the agent act for that user; a child (parent_delegation_id) lets an agent pass part of what it holds to a sub-agent. A child can only narrow its parent: its scopes must be a subset, its lifetime ends no later, and the chain is limited by the instance's maxDelegationDepth. Loops are refused. The agent's tokens then carry delegation and act claims, are narrowed to the delegated scopes, and stop verifying when the delegation is revoked. Needs a recent MFA sign-in. Honours Idempotency-Key.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `POST /v1/agent_delegations` (create); `GET /v1/agent_delegations/{id}` (read); no update route: every change replaces the object; `POST /v1/agent_delegations/{id}/revoke` (destroy).

## Example Usage

```hcl
resource "atlas_agent_delegation" "example" {
  child_machine_id = "tf-agent-delegation"
  scopes = ["example"]
}
```

## Schema

### Required

- `child_machine_id` (String) The `child_machine_id` field.
- `scopes` (List of String) The `scopes` field.

### Optional

- `delegator_user_id` (String) For a root delegation.
- `parent_delegation_id` (String) For a child delegation.
- `resource` (String) A resource server identifier.
- `expires_in_seconds` (Number) The `expires_in_seconds` field. Not returned by the API, so it is not detected as drift.
- `expires_at` (String) ISO 8601 or epoch ms, within 30 days. Default: the instance's defaultDelegationTtlSeconds.
- `organization_id` (String) The `organization_id` field.

### Read-Only

- `id` (String) Atlas object id.
- `root_delegation_id` (String) The `root_delegation_id` field.
- `parent_machine_id` (String) The agent that delegated (null for a delegation from a user).
- `depth` (Number) The `depth` field.
- `status` (String) The `status` field. One of: `active`, `revoked`.
- `revoked_at` (Number) Epoch ms.
- `revoked_by` (String) The `revoked_by` field.
- `revoked_reason` (String) The `revoked_reason` field.
- `posture` (String, JSON) The delegator's posture when the root delegation was made (mfa_at, risk, device). A JSON document (use `jsonencode`).
- `last_decision` (String) The `last_decision` field. One of: `allow`, `deny`, `revoked`, `null`.
- `last_evaluated_at` (Number) Epoch ms.
- `created_at` (Number) Epoch ms.
- `evaluations` (String, JSON) The `evaluations` field. A JSON document (use `jsonencode`).

## Import

The import id is `<id>`.

```sh
terraform import atlas_agent_delegation.example <id>
```

Write-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.
