# atlas_machine_trust_policy

Create a trust policy: let a token from an external workload platform (GitHub Actions, Kubernetes, AWS, GCP or any OIDC issuer) be exchanged for a machine token of one machine, with no stored secret. The external token must verify against the issuer, carry the audience, and its subject and claims must satisfy the policy. A machine holds at most 20. Names are unique per machine (409). Needs a recent MFA sign-in (it widens what can act as the machine). Honours Idempotency-Key.

Scopes: `machines:read` to read, `machines:write` to write.

Backend API: `POST /v1/machine_trust_policies` (create); `GET /v1/machine_trust_policies/{id}` (read); `PATCH /v1/machine_trust_policies/{id}` (update); `DELETE /v1/machine_trust_policies/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_machine_trust_policy" "example" {
  machine_id = "tf-machine-trust-policy"
  name = "tf-machine-trust-policy"
  kind = "github_actions"
  audience = "tf-machine-trust-policy"
  subject_pattern = "tf-machine-trust-policy"
}
```

## Schema

### Required

- `machine_id` (String) The `machine_id` field. Changing it forces replacement.
- `name` (String) Unique per machine.
- `kind` (String) Where the workload token comes from. Fixed at create. github_actions and gcp default the issuer; kubernetes and oidc need one; aws takes none.
- `audience` (String) The audience the external token must carry. For aws it is the value the workload signs into the x-atlas-audience header (1 to 128 printable characters).
- `subject_pattern` (String) A glob (* and ?) the external token's sub must match, for example repo:my-org/my-repo:ref:refs/heads/main. For aws: the role ARN, for example arn:aws:iam::123456789012:role/deployer (an assumed-role session is matched as its role). A pattern of only wildcards is refused, and each kind must name its owner (GitHub owner, AWS account, Kubernetes namespace).

### Optional

- `issuer` (String) The expected iss of the external token (https). Null for aws, and for github_actions and gcp to use their standard issuer.
- `jwks_uri` (String) Where the issuer publishes its signing keys (https). Null: found through the issuer's OpenID configuration.
- `jwks` (String, JSON) Inline public signing keys, for an issuer Atlas cannot reach (a cluster's service account issuer): { keys: [...] }. Private key parameters are refused. A JSON document (use `jsonencode`).
- `claim_conditions` (String, JSON) Every condition must hold, on top of the subject pattern. A JSON document (use `jsonencode`).
- `max_ttl_seconds` (Number) The longest a machine token minted through this policy lives (also capped by the API's token lifetime and the grant). Default 600.
- `enabled` (Boolean) A disabled policy accepts nothing (it can still be tested).

### Read-Only

- `id` (String) Atlas object id.
- `last_used_at` (Number) Epoch ms of the last exchange through this policy.
- `created_at` (Number) The `created_at` field. (epoch ms)
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

The import id is `<id>`.

```sh
terraform import atlas_machine_trust_policy.example <id>
```
