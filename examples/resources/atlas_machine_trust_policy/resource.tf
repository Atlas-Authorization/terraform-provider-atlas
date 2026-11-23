# Create a trust policy: let a token from an external workload platform (GitHub Actions, Kubernetes, AWS, GCP or any OIDC issuer) be exchanged for a machine token of one machine, with no stored secret.
resource "atlas_machine_trust_policy" "example" {
  machine_id = "tf-machine-trust-policy"
  name = "tf-machine-trust-policy"
  kind = "github_actions"
  audience = "tf-machine-trust-policy"
  subject_pattern = "tf-machine-trust-policy"
}
