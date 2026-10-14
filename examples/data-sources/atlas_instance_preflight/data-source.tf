# Warn on every plan when the instance setup has errors.
data "atlas_instance_preflight" "this" {}

check "atlas_setup" {
  assert {
    condition     = data.atlas_instance_preflight.this.ok
    error_message = "Atlas preflight found ${data.atlas_instance_preflight.this.error_count} error(s)."
  }
}
